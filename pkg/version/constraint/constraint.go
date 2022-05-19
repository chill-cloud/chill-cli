package constraint

import (
	"fmt"
	"github.com/chill-cloud/chill-cli/pkg/version"
	"strings"
)

type Constraint interface {
	FitConstraint(v version.Version) bool
	String() string
	DetailedString() string
}

type RangedConstraint struct {
	Lower version.Version
	Upper version.Version
}

type AnnotatedConstraint struct {
	C          Constraint
	Annotation string
}

type MajorOnlyConstraint struct {
	C Constraint
}

type MinorOnlyConstraint struct {
	C Constraint
}

type AnyConstraint struct{}

func New(lower version.Version, upper version.Version) Constraint {
	return &RangedConstraint{
		Lower: lower,
		Upper: upper,
	}
}

func NewMajorOnly(c Constraint) Constraint {
	return &MajorOnlyConstraint{c}
}

func NewMinorOnly(c Constraint) Constraint {
	return &MinorOnlyConstraint{c}
}

func ParseFromString(str string) (Constraint, error) {
	s := strings.TrimPrefix(str, "v")
	parts := strings.Split(s, ".")
	if len(parts) < 1 {
		return nil, fmt.Errorf("too few parts for string %s", str)
	} else {
		major, err := version.ParseVersionPart(parts[0])
		if err != nil {
			return nil, err
		}
		if len(parts) == 1 {
			return &AnnotatedConstraint{
				C: New(
					version.Version{Major: major, Minor: 0, Patch: 0},
					version.Version{Major: major + 1, Minor: 0, Patch: 0},
				),
				Annotation: str,
			}, nil
		}
		minor, err := version.ParseVersionPart(parts[1])
		if err != nil {
			return nil, err
		}
		if len(parts) == 2 {
			return &AnnotatedConstraint{
				C: New(
					version.Version{Major: major, Minor: minor, Patch: 0},
					version.Version{Major: major, Minor: minor, Patch: 1},
				), Annotation: str,
			}, nil
		}
		patch, err := version.ParseVersionPart(parts[2])
		if err != nil {
			return nil, err
		}
		if len(parts) == 3 {
			return &AnnotatedConstraint{
				C: New(
					version.Version{Major: major, Minor: minor, Patch: patch},
					version.Version{Major: major, Minor: minor, Patch: patch + 1},
				), Annotation: str,
			}, nil
		} else {
			return nil, fmt.Errorf("too many parts for string %s", str)
		}
	}
}

var anyConstraint = AnyConstraint{}

func Any() Constraint {
	return &anyConstraint
}

func (rc *RangedConstraint) String() string {
	return fmt.Sprintf("(>=%s, <%s)", rc.Lower.String(), rc.Upper.String())
}

func (rc *RangedConstraint) DetailedString() string {
	return fmt.Sprintf("(>=%s, <%s)", rc.Lower.String(), rc.Upper.String())
}

func (rc *RangedConstraint) FitConstraint(v version.Version) bool {
	return v.Compare(rc.Lower) >= 0 && v.Compare(rc.Upper) == -1
}

func (mc *MajorOnlyConstraint) String() string {
	return fmt.Sprintf("[major-only %s]", mc.C.String())
}

func (mc *MajorOnlyConstraint) DetailedString() string {
	return fmt.Sprintf("[major-only %s]", mc.C.String())
}

func (mc *MajorOnlyConstraint) FitConstraint(v version.Version) bool {
	if v.GetMinor() != 0 || v.GetPatch() != 0 {
		return false
	}
	return mc.C.FitConstraint(v)
}

func (mc *MinorOnlyConstraint) String() string {
	return fmt.Sprintf("[minor-only %s]", mc.C.String())
}

func (mc *MinorOnlyConstraint) DetailedString() string {
	return fmt.Sprintf("[minor-only %s]", mc.C.String())
}

func (mc *MinorOnlyConstraint) FitConstraint(v version.Version) bool {
	if v.GetPatch() != 0 {
		return false
	}
	return mc.C.FitConstraint(v)
}

func (ac *AnnotatedConstraint) String() string {
	return ac.Annotation
}

func (ac *AnnotatedConstraint) DetailedString() string {
	return ac.C.DetailedString()
}

func (ac *AnnotatedConstraint) FitConstraint(v version.Version) bool {
	return ac.C.FitConstraint(v)
}

func (ac *AnyConstraint) String() string {
	return "(*)"
}

func (ac *AnyConstraint) DetailedString() string {
	return "(any version)"
}

func (ac *AnyConstraint) FitConstraint(v version.Version) bool {
	return true
}
