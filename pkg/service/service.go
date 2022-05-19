package service

import (
	"fmt"
	"github.com/chill-cloud/chill-cli/pkg/cache"
	"github.com/chill-cloud/chill-cli/pkg/version"
)

type Stage int

const (
	StageDevelopment Stage = iota
	StageProduction
	StageMajor
)

//type ProjectConfig interface {
//	GetName() string
//	GetRemote() string
//	GetRegistry() string
//	GetBaseVersion() version.Version
//	GetCurrentVersion() version.Version
//	GetStage() Stage
//	GetIntegration() string
//	GetDependencies() map[Dependency]bool
//}

type ChillService interface {
	GetBuildTag() string
	UpdateDependencies(ctx cache.LocalCacheContext) error
}

type IncrementalConfig interface {
	ApplyIdempotent(c *ProjectConfig) error
}

type ProjectConfig struct {
	Name           string
	Registry       string
	Clients        map[string]string
	BaseVersion    *version.Version
	CurrentVersion *version.Version
	Stage          Stage
	Integration    string
	Dependencies   map[Dependency]bool
	TrafficTargets map[version.Version]int
	Secrets        []string
}

func (pc *ProjectConfig) GetTrafficTargets() (map[version.Version]int, error) {
	if !version.IsProduction(*pc.CurrentVersion) {
		return nil, fmt.Errorf("it is not possible to set traffic targets for non-production versions")
	}
	if pc.TrafficTargets == nil {
		return map[version.Version]int{
			*pc.CurrentVersion: 100,
		}, nil
	} else {
		return pc.TrafficTargets, nil
	}
}

func (pc *ProjectConfig) ApplyIdempotent(c *ProjectConfig) error {
	if c.Name != "" && pc.Name != c.Name {
		return fmt.Errorf("it is not a good idea to rename service")
	}
	pc.BaseVersion = c.BaseVersion
	pc.CurrentVersion = c.CurrentVersion
	if pc.TrafficTargets == nil {
		pc.TrafficTargets = c.TrafficTargets
	}
	//if c.Remote != "" {
	//	pc.Remote = c.Remote
	//}
	//if c.Registry != "" {
	//	pc.Registry = c.Registry
	//}
	//if pc.Integration != "" {
	//	pc.Integration = c.Integration
	//}
	//pc.Dependencies = c.Dependencies
	return nil
}

func (pc *ProjectConfig) GetBuildTag(forceLocal bool) (string, bool) {
	isLocal := pc.Registry == "" || forceLocal
	var registry string
	if isLocal {
		registry = "dev.local"
	} else {
		registry = pc.Registry
	}
	return fmt.Sprintf("%s/%s:%s", registry, pc.Name, pc.CurrentVersion.String()), isLocal
}

func (pc *ProjectConfig) UpdateDependencies(ctx cache.LocalCacheContext) error {
	for dep := range pc.Dependencies {
		err := dep.Cache().Update(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
