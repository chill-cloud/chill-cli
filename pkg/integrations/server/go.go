package server

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type goIntegration struct{}

func (g *goIntegration) GenerateMethods(cwd string, name string, protoSource string) error {
	name = strings.ReplaceAll(name, "-", "")
	protoPath := filepath.Join(protoSource, "api")
	protos, err := filepath.Glob(filepath.Join(protoSource, "api", "*") + "/*.proto")
	if err != nil {
		return err
	}
	for i := 0; i < len(protos); i++ {
		protos[i] = strings.TrimPrefix(protos[i], protoPath+"/")
	}
	var moduleDeclarations []string

	for _, p := range protos {
		moduleDeclarations = append(moduleDeclarations, "--go_opt=M"+p+"=./"+name, "--go-grpc_opt=M"+p+"=./"+name)
	}

	targetPath := filepath.Join(cwd, "src", "internal", "generated")

	err = os.MkdirAll(targetPath, os.ModePerm)
	if err != nil {
		return err
	}

	genCmd := append([]string{
		"--go_out=" + targetPath,
		"--go-grpc_out=" + targetPath,
	}, append(
		moduleDeclarations,
		append(
			[]string{"-I" + protoPath},
			protos...)...,
	)...,
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

func (g *goIntegration) GetBaseProjectRemote() string {
	return "github.com/chill-cloud/base-project-go"
}

func init() {
	Register("go", &goIntegration{})
}
