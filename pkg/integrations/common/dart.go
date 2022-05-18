package common

import (
	"github.com/chill-cloud/chill-cli/pkg/service/naming"
	"github.com/chill-cloud/chill-cli/pkg/util"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func GenerateMethodsDart(cwd string, name string, protoSource string, visibility bool) error {
	name = naming.Merge(naming.SplitIntoParts(name), "_", naming.ModeLower)
	protoPath := filepath.Join(protoSource, "api")
	protos, err := GetPathsForVisibility(protoSource, visibility)
	if err != nil {
		return err
	}
	for i := 0; i < len(protos); i++ {
		protos[i] = strings.TrimPrefix(protos[i], protoPath+"/")
	}

	targetPath := filepath.Join(cwd, "lib", "chillgen", name)

	err = os.MkdirAll(targetPath, os.ModePerm)
	if err != nil {
		return err
	}

	genCmd := append([]string{
		"--dart_out=grpc:" + targetPath,
	}, append(
		[]string{"-I" + protoPath},
		protos...)...,
	)
	q := exec.Command(
		"protoc",
		genCmd...,
	)
	err = util.RunCmdDetailed(q)
	if err != nil {
		return err
	}
	return nil
}
