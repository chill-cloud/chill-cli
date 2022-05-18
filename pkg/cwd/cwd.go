package cwd

import (
	"errors"
	"fmt"
	"github.com/chill-cloud/chill-cli/pkg/config"
	"os"
	"path/filepath"
	"strings"
)

func SetupCwd(cwd string) (string, error) {
	d, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if cwd != "" {
		return filepath.Join(d, cwd), nil
	}
	l := strings.Split(d, string(os.PathSeparator))
	for i := len(l); i >= 0; i-- {
		_, err := os.Stat(strings.Join(append(l[:i], config.ProjectConfigName), string(os.PathSeparator)))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			} else {
				return "", err
			}
		} else {
			res := strings.Join(l[:i], string(os.PathSeparator))
			if i != len(l) {
				println(fmt.Sprintf("Switched project directory to %s", res))
			}
			return res, nil
		}
	}
	return "", fmt.Errorf("no project found in file hierarchy")
}
