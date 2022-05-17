package util

import (
	"errors"
	"fmt"
	"github.com/chill-cloud/chill-cli/pkg/config"
	"os"
	"path/filepath"
)

func SetupCwd(cwd string) (string, error) {
	d, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if cwd != "" {
		return filepath.Join(d, cwd), nil
	}
	l := filepath.SplitList(d)
	for i := len(l); i >= 0; i-- {
		_, err := os.Stat(filepath.Join(append(l[:i], config.ProjectConfigName)...))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			} else {
				return "", err
			}
		} else {
			res := filepath.Join(l[:i]...)
			if i != len(l) {
				println(fmt.Sprintf("Switched project directory to %s", res))
			}
			return res, nil
		}
	}
	return "", fmt.Errorf("no project found in file hierarchy")
}
