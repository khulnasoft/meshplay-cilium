package cilium

import (
	"sync"

	meshplaykube "github.com/khulnasoft/meshkit/utils/kubernetes"
	"github.com/layer5io/meshery-adapter-library/adapter"
	"github.com/layer5io/meshery-adapter-library/status"
)

func (h *Cilium) installSampleApp(del bool, namespace string, templates []adapter.Template, kubeconfigs []string) (string, error) {
	st := status.Installing
	if del {
		st = status.Removing
	}
	for _, template := range templates {
		err := h.applyManifest([]byte(template.String()), del, namespace, kubeconfigs)
		if err != nil {
			return st, ErrSampleApp(err)
		}
	}
	return status.Installed, nil
}

func (h *Cilium) applyManifest(contents []byte, isDel bool, namespace string, kubeconfigs []string) error {
	var errs []error
	var wg sync.WaitGroup
	for _, k8sconfig := range kubeconfigs {
		wg.Add(1)
		go func(k8sconfig string) {
			defer wg.Done()
			kClient, err := meshplaykube.New([]byte(k8sconfig))
			if err != nil {
				errs = append(errs, err)
				return
			}
			err = kClient.ApplyManifest(contents, meshplaykube.ApplyOptions{
				Namespace: namespace,
				Update:    true,
				Delete:    isDel,
			})
			if err != nil {
				errs = append(errs, err)
				return
			}
		}(k8sconfig)
	}
	wg.Wait()
	if len(errs) != 0 {
		return mergeErrors(errs)
	}
	return nil
}
