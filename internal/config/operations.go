package config

import (
	"github.com/khulnasoft/meshkit/utils"
	"github.com/layer5io/meshery-adapter-library/adapter"
	"github.com/layer5io/meshery-adapter-library/meshes"
)

var (
	ServiceName = "service_name"
)

func getOperations(dev adapter.Operations) adapter.Operations {
	var adapterVersions []adapter.Version
	versions, _ := utils.GetLatestReleaseTagsSorted("cilium", "cilium")
	for _, v := range versions {
		adapterVersions = append(adapterVersions, adapter.Version(v))
	}

	dev[CiliumOperation] = &adapter.Operation{
		Type:                 int32(meshes.OpCategory_INSTALL),
		Description:          "Cilium Service Mesh",
		Versions:             adapterVersions,
		Templates:            []adapter.Template{},
		AdditionalProperties: map[string]string{},
	}

	return dev
}
