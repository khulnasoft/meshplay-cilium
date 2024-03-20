package oam

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/khulnasoft/meshkit/models/meshmodel/core/types"
	"github.com/layer5io/meshery-adapter-library/adapter"
)

var (
	//WorkloadPath will be used by both static and component generation
	WorkloadPath        = filepath.Join(basePath, "templates", "oam", "workloads")
	MeshmodelComponents = filepath.Join(basePath, "templates", "meshmodel", "components")
	traitPath           = filepath.Join(basePath, "templates", "oam", "traits")
	basePath, _         = os.Getwd()
)

// AvailableVersions denote the component versions available statically
var AvailableVersions = map[string]bool{}
var availableVersionGlobalMutex sync.Mutex

type schemaDefinitionPathSet struct {
	oamDefinitionPath string
	jsonSchemaPath    string
	name              string
}

type meshmodelDefinitionPathSet struct {
	meshmodelDefinitionPath string
}

func RegisterMeshModelComponents(uuid, runtime, host, port string) error {
	meshmodelRDP := []adapter.MeshModelRegistrantDefinitionPath{}
	pathSets, err := loadMeshmodelComponents(MeshmodelComponents)
	if err != nil {
		return err
	}
	portint, _ := strconv.Atoi(port)
	for _, pathSet := range pathSets {
		meshmodelRDP = append(meshmodelRDP, adapter.MeshModelRegistrantDefinitionPath{
			EntityDefintionPath: pathSet.meshmodelDefinitionPath,
			Host:                host,
			Port:                portint,
			Type:                types.ComponentDefinition,
		})
	}

	return adapter.
		NewMeshModelRegistrant(meshmodelRDP, fmt.Sprintf("%s/api/meshmodel/components/register", runtime)).
		Register(uuid)
}

func loadMeshmodelComponents(basepath string) ([]meshmodelDefinitionPathSet, error) {
	res := []meshmodelDefinitionPathSet{}
	if err := filepath.Walk(basepath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		res = append(res, meshmodelDefinitionPathSet{
			meshmodelDefinitionPath: path,
		})
		availableVersionGlobalMutex.Lock()
		AvailableVersions[filepath.Base(filepath.Dir(path))] = true // Getting available versions already existing on file system
		availableVersionGlobalMutex.Unlock()
		return nil
	}); err != nil {
		return nil, err
	}

	return res, nil
}
