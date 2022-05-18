package config

import "C"
import (
	"errors"
	"fmt"
	"github.com/chill-cloud/chill-cli/pkg/integrations/server"
	service2 "github.com/chill-cloud/chill-cli/pkg/service"
	"github.com/chill-cloud/chill-cli/pkg/version"
	"github.com/chill-cloud/chill-cli/pkg/version/constraint"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
)

const ProjectConfigName = "chill.yaml"
const LockConfigName = ".chill-lock.yaml"

var stages = map[string]service2.Stage{
	"development": service2.StageDevelopment,
	"production":  service2.StageProduction,
	"major":       service2.StageMajor,
}

var stagesToString = map[service2.Stage]string{
	service2.StageDevelopment: "development",
	service2.StageProduction:  "production",
	service2.StageMajor:       "major",
}

type SerializedDependency struct {
	Remote          string `yaml:"remote,omitempty"`
	Local           string `yaml:"local,omitempty"`
	Version         string `yaml:"version"`
	SpecificVersion string `yaml:"specificVersion,omitempty"`
}

type SerializedService struct {
	Name           string                          `yaml:"name,omitempty"`
	Registry       string                          `yaml:"registry,omitempty"`
	Remote         string                          `yaml:"remote,omitempty"`
	Clients        map[string]string               `yaml:"clients,omitempty"`
	BaseVersion    string                          `yaml:"baseVersion,omitempty"`
	CurrentVersion string                          `yaml:"currentVersion,omitempty"`
	Stage          string                          `yaml:"stage"`
	Integration    string                          `yaml:"integration"`
	Dependencies   map[string]SerializedDependency `yaml:"dependencies"`
	TrafficTargets map[string]int                  `yaml:"trafficTargets,omitempty"`
	Secrets        []string                        `yaml:"secrets,omitempty"`
}

const lockWarning = `# THIS IS AN AUTO-GENERATED FILE; DO NOT MODIFY!
# DO NOT EXCLUDE THIS FILE FROM THE VERSION CONTROL

`

type SerializedLockFile struct {
	Service SerializedService
}

func parseDependencies(m map[string]SerializedDependency) (map[service2.Dependency]bool, error) {
	res := map[service2.Dependency]bool{}

	for name, data := range m {
		var d service2.Dependency
		c, err := constraint.ParseFromString(data.Version)
		if err != nil {
			return nil, err
		}
		var v *version.Version
		if data.SpecificVersion != "" {
			v, err = version.ParseFromString(data.SpecificVersion)
			if err != nil {
				return nil, err
			}
		}
		if data.Remote != "" {
			d = &service2.RemoteDependency{
				Name:            name,
				Git:             data.Remote,
				Version:         c,
				SpecificVersion: v,
			}
		} else {
			d = &service2.LocalDependency{
				Name:            name,
				Path:            data.Local,
				Version:         c,
				SpecificVersion: v,
			}
		}
		res[d] = true
	}
	return res, nil
}

func ParseConfig(cwd string, file string, lock bool) (*service2.ProjectConfig, error) {
	configFile := filepath.Join(cwd, file)
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		} else {
			return nil, err
		}
	}
	var l SerializedLockFile
	err = yaml.Unmarshal(data, &l)
	s := &l.Service
	if err != nil {
		return nil, err
	}
	var c service2.ProjectConfig

	c.Name = s.Name
	c.Registry = s.Registry
	if s.Integration == "" {
		c.Integration = server.DefaultServerName
	} else {
		c.Integration = s.Integration
	}
	c.Clients = s.Clients

	if stage, ok := stages[s.Stage]; ok {
		c.Stage = stage
	} else {
		if lock {
			return nil, fmt.Errorf("unknown stage type")
		}
	}

	if s.BaseVersion != "" {
		if !lock {
			return nil, fmt.Errorf("version should not be set in a config file")
		}
		c.BaseVersion, err = version.ParseFromString(s.BaseVersion)
		if err != nil {
			return nil, err
		}
	}

	if s.BaseVersion != "" && !lock {
		return nil, fmt.Errorf("version should not be set in a config file")
	}
	if s.CurrentVersion != "" {
		c.CurrentVersion, err = version.ParseFromString(s.CurrentVersion)
		if err != nil {
			return nil, err
		}
	}

	c.Dependencies, err = parseDependencies(s.Dependencies)
	if err != nil {
		return nil, err
	}

	if s.TrafficTargets != nil {
		c.TrafficTargets = map[version.Version]int{}
		sum := 0
		for vs, p := range s.TrafficTargets {
			v, err := version.ParseFromString(vs)
			if err != nil {
				return nil, err
			}
			c.TrafficTargets[*v] = p
			sum += p
		}

		if sum != 100 {
			return nil, fmt.Errorf("sum of percents should be 100")
		}
	}

	c.Secrets = s.Secrets

	return &c, nil
}

func ProcessConfig(c *service2.ProjectConfig) (*SerializedService, error) {
	var s SerializedService
	s.Name = c.Name
	s.Registry = c.Registry
	s.Stage = stagesToString[c.Stage]
	// there might be no base version if this version is the first
	if c.BaseVersion != nil {
		s.BaseVersion = c.BaseVersion.String()
	}
	s.Clients = c.Clients
	s.Integration = c.Integration
	if c.CurrentVersion != nil {
		s.CurrentVersion = c.CurrentVersion.String()
	}
	s.Dependencies = map[string]SerializedDependency{}
	for dep, _ := range c.Dependencies {
		if remote, ok := dep.(*service2.RemoteDependency); ok {
			s.Dependencies[dep.GetName()] = SerializedDependency{
				Remote:          remote.Git,
				Version:         remote.Version.String(),
				SpecificVersion: remote.SpecificVersion.String(),
			}
		} else if local, ok := dep.(*service2.LocalDependency); ok {
			s.Dependencies[dep.GetName()] = SerializedDependency{
				Local:           local.Path,
				Version:         local.Version.String(),
				SpecificVersion: local.SpecificVersion.String(),
			}
		} else {
			return nil, fmt.Errorf("unknown type of dependency")
		}
	}
	if c.TrafficTargets != nil {
		s.TrafficTargets = map[string]int{}
		for ver, percent := range c.TrafficTargets {
			s.TrafficTargets[ver.String()] = percent
		}
	}
	s.Secrets = c.Secrets
	return &s, nil
}

func (s *SerializedService) SaveToFile(path string, lock bool) error {
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()
	res, err := yaml.Marshal(SerializedLockFile{Service: *s})
	if err != nil {
		return err
	}
	if lock {
		_, err = out.Write([]byte(lockWarning))
		if err != nil {
			return err
		}
	}
	_, err = out.Write(res)
	if err != nil {
		return err
	}
	return nil
}
