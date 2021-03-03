package main

import (
	"io/ioutil"
	"os/exec"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/skv2/codegen/util"
)

// this file is a Go script which updates the go.mod file to match the contents of Chart-template.yaml

const (
	glooMeshDependencyName          = "gloo-mesh"
	glooMeshRepo                    = "github.com/solo-io/gloo-mesh"
	glooMeshEnerpriseDependencyName = "gloo-mesh-extender"
	glooMeshEnterpriseRepo          = "github.com/solo-io/gloo-mesh-enterprise"
)

var moduleRoot = util.GetModuleRoot()

func main() {
	if err := run(); err != nil {
		log.Fatalf("%v", err)
	}
}

func run() error {
	log.Printf("updating gomod to match chart-template...")
	chartTemplateFile := filepath.Join(moduleRoot, "install", "helm", "gloo-mesh-enterprise", "Chart-template.yaml")

	chartTemplate, err := ioutil.ReadFile(chartTemplateFile)
	if err != nil {
		return err
	}

	var chartContents chartWithDependencies
	if err := yaml.Unmarshal(chartTemplate, &chartContents); err != nil {
		return err
	}

	var versions subchartVersions
	for _, dep := range chartContents.Dependencies {
		switch dep.Name {
		case glooMeshEnerpriseDependencyName:
			versions.glooMeshEnterprise = dep.Version
		}
	}

	if versions.glooMeshEnterprise == "" {
		return eris.Errorf("no gloo-mesh-enterprise version found")
	}

	return versions.updateGoMod()
}

// the versions of subcharts we import
type subchartVersions struct {
	glooMeshEnterprise string
}

func (v subchartVersions) updateGoMod() error {
	for repo, version := range map[string]string{
		glooMeshEnterpriseRepo: v.glooMeshEnterprise,
	} {
		if err := goGet(repo, version); err != nil {
			return err
		}
	}
	return nil
}

type chartWithDependencies struct {
	Dependencies []dependency `json:"dependencies"`
}

type dependency struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func goGet(repo, version string) error {
	out, err := exec.Command(
		"go",
		"get",
		"-v",
		repo+"@v"+version,
	).CombinedOutput()
	if err != nil {
		return eris.Wrapf(err, "running command (%v) failed: %v",
			[]string{"go", "get", "-v", repo + "@v" + version},
			string(out))
	}
	return nil
}
