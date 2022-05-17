package version

import (
	"fmt"
	"strconv"
	"strings"
)

//type Version interface {
//	GetMajor() int
//	GetMinor() int
//	GetPatch() int
//	Compare(other Version) int
//	String() string
//	MayBeNext(other Version) bool
//}

type Version struct {
	Major int
	Minor int
	Patch int
}

func versionMayBeNext(from *Version, to *Version) bool {
	if from == nil {
		return true
	}
	if from != nil && to == nil {
		return false
	}
	if from.GetMajor() == to.GetMajor() {
		if from.GetMinor() == to.GetMinor() {
			return from.GetPatch()+1 == to.GetPatch()
		}
		return from.GetMinor()+1 == to.GetMinor() && to.GetPatch() == 0
	}
	return from.GetMajor()+1 == to.GetMajor() && to.GetMinor() == 0 && to.GetPatch() == 0
}

func versionToString(v *Version) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("v%d.%d.%d", v.GetMajor(), v.GetMinor(), v.GetPatch())
}

func (v *Version) GetMajor() int {
	return v.Major
}

func (v *Version) GetMinor() int {
	return v.Minor
}

func (v *Version) GetPatch() int {
	return v.Patch
}

func (v *Version) String() string {
	return versionToString(v)
}

func (v *Version) MayBeNext(other *Version) bool {
	return versionMayBeNext(v, other)
}

func New(major int, minor int, patch int) *Version {
	return &Version{Major: major, Minor: minor, Patch: patch}
}

func ParseVersionPart(part string) (int, error) {
	res, err := strconv.Atoi(part)
	if err != nil {
		return 0, err
	}
	if res < 0 {
		return 0, fmt.Errorf("part of version must not be negative")
	}
	return res, nil
}

func ParseFromString(str string) (*Version, error) {
	s := strings.TrimPrefix(str, "v")
	var res Version
	parts := strings.Split(s, ".")
	if len(parts) < 1 {
		return nil, fmt.Errorf("too few parts for string %s", str)
	}

	major, err := ParseVersionPart(parts[0])
	if err != nil {
		return nil, err
	}
	res.Major = major

	if len(parts) >= 2 {
		minor, err := ParseVersionPart(parts[1])
		if err != nil {
			return nil, err
		}
		res.Minor = minor
	}

	if len(parts) >= 3 {
		patch, err := ParseVersionPart(parts[2])
		if err != nil {
			return nil, err
		}
		res.Patch = patch
	}

	if len(parts) > 3 {
		return nil, fmt.Errorf("too much parts for string %s", str)
	}
	return &res, nil
}

func IsProduction(v Version) bool {
	return v.GetPatch() == 0
}

func (v *Version) Compare(otherV Version) int {
	if v.GetMajor() < otherV.GetMajor() {
		return -1
	}
	if v.GetMajor() > otherV.GetMajor() {
		return 1
	}
	if v.GetMinor() < otherV.GetMinor() {
		return -1
	}
	if v.GetMinor() > otherV.GetMinor() {
		return 1
	}
	if v.GetPatch() < otherV.GetPatch() {
		return -1
	}
	if v.GetPatch() > otherV.GetPatch() {
		return 1
	}
	return 0
}
