package cilium

import (
	"fmt"
	"strings"

	"github.com/khulnasoft/meshkit/models/oam/core/v1alpha1"
	"gopkg.in/yaml.v2"
)

// CompHandler is the type for functions which can handle OAM components
type CompHandler func(*Cilium, v1alpha1.Component, bool, []string) (string, error)

// HandleComponents handles the processing of OAM components
func (h *Cilium) HandleComponents(comps []v1alpha1.Component, isDel bool, kubeconfigs []string) (string, error) {
	var errs []error
	var msgs []string

	compFuncMap := map[string]CompHandler{
		"CiliumMesh": handleComponentCiliumMesh,
	}

	for _, comp := range comps {
		fnc, ok := compFuncMap[comp.Spec.Type]
		if !ok {
			msg, err := handleCiliumCoreComponent(h, comp, isDel, "", "", kubeconfigs)
			if err != nil {
				errs = append(errs, err)
				continue
			}

			msgs = append(msgs, msg)
			continue
		}

		msg, err := fnc(h, comp, isDel, kubeconfigs)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		msgs = append(msgs, msg)
	}

	if err := mergeErrors(errs); err != nil {
		return mergeMsgs(msgs), err
	}

	return mergeMsgs(msgs), nil
}

// HandleApplicationConfiguration handles the processing of OAM application configuration
func (h *Cilium) HandleApplicationConfiguration(config v1alpha1.Configuration, isDel bool, kubeconfigs []string) (string, error) {
	var errs []error
	var msgs []string
	for _, comp := range config.Spec.Components {
		for _, trait := range comp.Traits {
			msgs = append(msgs, fmt.Sprintf("applied trait \"%s\" on service \"%s\"", trait.Name, comp.ComponentName))
		}
	}

	if err := mergeErrors(errs); err != nil {
		return mergeMsgs(msgs), err
	}

	return mergeMsgs(msgs), nil
}

func handleComponentCiliumMesh(h *Cilium, comp v1alpha1.Component, isDel bool, kubeconfigs []string) (string, error) {
	// Get the Cilium version from the settings
	// we are sure that the version of Cilium would be present
	// because the configuration is already validated against the schema
	version := comp.Spec.Version
	if version == "" {
		return "", fmt.Errorf("pass valid version inside service for Cilium installation")
	}
	//TODO: When no version is passed in service, use the latest Cilium version

	msg, err := h.installCilium(isDel, version, comp.Namespace, kubeconfigs)
	if err != nil {
		return fmt.Sprintf("%s: %s", comp.Name, msg), err
	}

	return fmt.Sprintf("%s: %s", comp.Name, msg), nil
}

func handleCiliumCoreComponent(
	h *Cilium,
	comp v1alpha1.Component,
	isDel bool,
	apiVersion,
	kind string,
	kubeconfigs []string) (string, error) {
	if apiVersion == "" {
		apiVersion = getAPIVersionFromComponent(comp)
		if apiVersion == "" {
			return "", ErrCiliumCoreComponentFail(fmt.Errorf("failed to get API Version for: %s", comp.Name))
		}
	}

	if kind == "" {
		kind = getKindFromComponent(comp)
		if kind == "" {
			return "", ErrCiliumCoreComponentFail(fmt.Errorf("failed to get kind for: %s", comp.Name))
		}
	}

	component := map[string]interface{}{
		"apiVersion": apiVersion,
		"kind":       kind,
		"metadata": map[string]interface{}{
			"name":        comp.Name,
			"annotations": comp.Annotations,
			"labels":      comp.Labels,
		},
		"spec": comp.Spec.Settings,
	}

	// Convert to yaml
	yamlByt, err := yaml.Marshal(component)
	if err != nil {
		err = ErrParseCiliumCoreComponent(err)
		h.Log.Error(err)
		return "", err
	}

	msg := fmt.Sprintf("created %s \"%s\" in namespace \"%s\"", kind, comp.Name, comp.Namespace)
	if isDel {
		msg = fmt.Sprintf("deleted %s config \"%s\" in namespace \"%s\"", kind, comp.Name, comp.Namespace)
	}

	return msg, h.applyManifest(yamlByt, isDel, comp.Namespace, kubeconfigs)
}

func getAPIVersionFromComponent(comp v1alpha1.Component) string {
	return comp.Annotations["pattern.meshplay.io.mesh.workload.k8sAPIVersion"]
}

func getKindFromComponent(comp v1alpha1.Component) string {
	return comp.Annotations["pattern.meshplay.io.mesh.workload.k8sKind"]
}

func mergeErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}

	var errMsgs []string

	for _, err := range errs {
		errMsgs = append(errMsgs, err.Error())
	}

	return fmt.Errorf(strings.Join(errMsgs, "\n"))
}

func mergeMsgs(strs []string) string {
	return strings.Join(strs, "\n")
}
