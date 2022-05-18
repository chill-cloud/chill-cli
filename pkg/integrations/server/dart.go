package server

import (
	"github.com/chill-cloud/chill-cli/pkg/integrations/common"
)

type dartIntegration struct{}

func (g *dartIntegration) GenerateMethods(cwd string, name string, protoSource string) error {
	return common.GenerateMethodsDart(cwd, name, protoSource, true)
}

func (g *dartIntegration) GetBaseProjectRemote() string {
	return "github.com/chill-cloud/base-project-dart"
}

func init() {
	Register("dart", &dartIntegration{})
}
