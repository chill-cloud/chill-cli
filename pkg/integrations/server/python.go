package server

import (
	"github.com/chill-cloud/chill-cli/pkg/integrations/common"
)

type pythonIntegration struct{}

func (g *pythonIntegration) GenerateMethods(cwd string, name string, protoSource string) error {
	return common.GenerateMethodsPython(cwd, name, protoSource, true)
}

func (g *pythonIntegration) GetBaseProjectRemote() string {
	return "github.com/chill-cloud/base-project-python"
}

func init() {
	Register("python", &pythonIntegration{})
}
