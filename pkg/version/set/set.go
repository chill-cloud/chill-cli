package set

import (
	"fmt"
	"github.com/chill-cloud/chill-cli/pkg/version"
	"github.com/chill-cloud/chill-cli/pkg/version/constraint"
	"sort"
)

type ArrayVersionSet []version.Version

type VersionSet interface {
	GetLatestVersion(c constraint.Constraint) version.Version
	Validate() error
	GetLatestProductionVersion() *version.Version
}

func (s ArrayVersionSet) GetLatestVersion(c constraint.Constraint) *version.Version {
	var best version.Version
	was := false
	for _, v := range s {
		if c.FitConstraint(v) && (!was || v.Compare(best) > 0) {
			was = true
			best = v
		}
	}
	if !was {
		return nil
	}
	return &best
}

func (s ArrayVersionSet) GetLatestMajorVersion() int {
	best := 0
	for _, v := range s {
		if v.GetMajor() > best {
			best = v.GetMajor()
		}
	}
	return best
}

func (s ArrayVersionSet) GetLatestProductionVersion() *version.Version {
	var res version.Version
	was := false
	for _, v := range s {
		if version.IsProduction(v) && (!was || res.Compare(v) < 0) {
			was = true
			res = v
		}
	}
	if !was {
		return nil
	}
	return &res
}

func (s ArrayVersionSet) Validate() error {
	sort.Slice(s, func(i, j int) bool {
		return s[i].Compare(s[j]) < 0
	})
	for i := 0; i < len(s)-1; i++ {
		if !s[i].MayBeNext(&s[i+1]) {
			return fmt.Errorf("broken service versioning: %s cannot be followed by %s", s[i].String(), s[i+1].String())
		}
	}
	return nil
}
