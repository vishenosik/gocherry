package versions

import (
	"errors"
	"fmt"
	"sort"
)

var (
	ErrDotFormat = errors.New("version must be in format MAJOR.MINOR")
)

type DotVersion struct {
	Major int
	Minor int
}

func NewDotVersion(version string) DotVersion {
	var v DotVersion
	v, err := ParseDotVersion(version)
	if err != nil {
		panic(err)
	}
	return v
}

// Parse parses a semantic version string into a SemanticVersion
func ParseDotVersion(version string) (DotVersion, error) {
	var double DotVersion
	_, err := fmt.Sscanf(version, "%d.%d", &double.Major, &double.Minor)
	if err != nil {
		return DotVersion{}, ErrDotFormat
	}
	return double, nil
}

// String returns the version as "MAJOR.MINOR" string
func (v DotVersion) String() string {
	return fmt.Sprintf("%d.%d", v.Major, v.Minor)
}

func (v DotVersion) In(v1, v2 DotVersion) bool {

	lower, upper := v1, v2
	if v1.Major > v2.Major || (v1.Major == v2.Major && v1.Minor > v2.Minor) {
		lower, upper = upper, lower
	}

	if lower.Major == upper.Major && lower.Minor == upper.Minor {
		if v.Major != upper.Major || v.Minor != upper.Minor {
			return false
		}
		return true
	} else if (v.Major < lower.Major || v.Major > upper.Major) ||
		(v.Major == lower.Major && v.Minor < lower.Minor) ||
		(v.Major == upper.Major && v.Minor > upper.Minor) {
		return false

	}
	return true
}

func (v DotVersion) In_(v1, v2 Interface) bool {
	converted_v1, ok := v1.(DotVersion)
	if !ok {
		return false
	}
	converted_v2, ok := v2.(DotVersion)
	if !ok {
		return false
	}

	return v.In(converted_v1, converted_v2)
}

func (v1 DotVersion) GTE(v2 DotVersion) bool {
	if v1.Major != v2.Major {
		return v1.Major >= v2.Major
	}
	if v1.Minor != v2.Minor {
		return v1.Minor >= v2.Major
	}
	return true
}

func LatestDotVersion(versions ...DotVersion) DotVersion {

	// Sort versions in descending order
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].GTE(versions[j])
	})

	return versions[0]

}
