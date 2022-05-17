package server

import (
	"fmt"
	"github.com/chill-cloud/chill-cli/pkg/service/naming"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type dartIntegration struct{}

func (g *dartIntegration) GenerateMethods(cwd string, name string, protoSource string) error {
	name = naming.Merge(naming.SplitIntoParts(name), "_", naming.ModeLower)
	protoPath := filepath.Join(protoSource, "api")
	protos, err := filepath.Glob(filepath.Join(protoSource, "api", "*") + "/*.proto")
	if err != nil {
		return err
	}
	for i := 0; i < len(protos); i++ {
		protos[i] = strings.TrimPrefix(protos[i], protoPath+"/")
	}

	targetPath := filepath.Join(cwd, "src", "lib", "chillgen", name)

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
	var outbuf, errbuf strings.Builder
	q.Stdout = &outbuf
	q.Stderr = &errbuf
	err = q.Run()
	if err != nil {
		return fmt.Errorf("Unable to generate gRPC stubs\n"+
			"stdout:\n%s"+
			"\n\n"+
			"stderr:\n%s",
			outbuf.String(), errbuf.String())

	}
	return nil
}

func (g *dartIntegration) GetBaseProjectRemote() string {
	return "github.com/chill-cloud/base-project-dart"
}

func init() {
	Register("dart", &dartIntegration{})
}
