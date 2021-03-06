package client

import (
	"github.com/chill-cloud/chill-cli/pkg/integrations/common"
	"github.com/chill-cloud/chill-cli/pkg/service/naming"
	"os"
	"path/filepath"
	"text/template"
)

type dartIntegration struct{}

var pubspecSrc = `name: {{.ServiceName}}
description: Generated by Chill
version: 1.0.0

environment:
  sdk: '>=2.16.2 <3.0.0'

dependencies:
  grpc: ^3.0.2
`

func (g *dartIntegration) GenerateClient(protoSource string, path string, name string) error {
	newName := naming.Merge(append(naming.SplitIntoParts(name), "dart", "codegen"), "_", naming.ModeLower)

	var replacement struct {
		ServiceName string
	}
	replacement.ServiceName = newName
	tmpl, err := template.New("pubspec.yaml").Parse(pubspecSrc)
	if err != nil {
		return err
	}

	out, err := os.Create(filepath.Join(path, "pubspec.yaml"))
	if err != nil {
		return err
	}
	defer out.Close()
	err = tmpl.Execute(out, replacement)
	if err != nil {
		return err
	}
	return common.GenerateMethodsDart(path, name, protoSource, false)
}

func init() {
	Register("dart", &dartIntegration{})
}
