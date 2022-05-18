package common

import "path/filepath"

func GetPathsForVisibility(protoSource string, all bool) ([]string, error) {
	if all {
		return filepath.Glob(filepath.Join(protoSource, "api", "*") + "/*.proto")
	} else {
		return filepath.Glob(filepath.Join(protoSource, "api", "public") + "/*.proto")
	}
}
