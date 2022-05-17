package service

import (
	"github.com/chill-cloud/chill-cli/pkg/cache"
	"github.com/chill-cloud/chill-cli/pkg/version"
	"github.com/chill-cloud/chill-cli/pkg/version/constraint"
)

type Dependency interface {
	GetName() string
	GetVersion() constraint.Constraint
	GetSpecificVersion() *version.Version
	SetSpecificVersion(*version.Version) error
	Cache() cache.CachedSource
}

type LocalDependency struct {
	Name            string
	Path            string
	Version         constraint.Constraint
	SpecificVersion *version.Version
}

func (ld *LocalDependency) GetName() string {
	return ld.Name
}

func (ld *LocalDependency) GetVersion() constraint.Constraint {
	return ld.Version
}

func (ld *LocalDependency) GetSpecificVersion() *version.Version {
	return ld.SpecificVersion
}

func (ld *LocalDependency) SetSpecificVersion(v *version.Version) error {
	ld.SpecificVersion = v
	return nil
}

func (ld *LocalDependency) Cache() cache.CachedSource {
	return &cache.GitLocalSource{
		LocalPath: ld.Path,
	}
}

type RemoteDependency struct {
	Name            string
	Git             string
	Version         constraint.Constraint
	SpecificVersion *version.Version
}

func (rd *RemoteDependency) GetName() string {
	return rd.Name
}

func (rd *RemoteDependency) GetVersion() constraint.Constraint {
	return rd.Version
}

func (rd *RemoteDependency) GetSpecificVersion() *version.Version {
	return rd.SpecificVersion
}

func (rd *RemoteDependency) SetSpecificVersion(v *version.Version) error {
	rd.SpecificVersion = v
	return nil
}

func (rd *RemoteDependency) Cache() cache.CachedSource {
	return &cache.GitSource{
		Remote: rd.Git,
	}
}
