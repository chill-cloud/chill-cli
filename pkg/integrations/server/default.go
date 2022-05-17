package server

import (
	"fmt"
	"github.com/chill-cloud/chill-cli/pkg/logging"
)

type defaultIntegration struct{}

func (g *defaultIntegration) GenerateMethods(cwd string, name string, protoSource string) error {
	logging.Logger.Info(fmt.Sprintf("Default integration is used for service %s, so no code will be generated", name))
	return nil
}

func (g *defaultIntegration) GetBaseProjectRemote() string {
	return "github.com/chill-cloud/base-project-default"
}

func init() {
	Register(DefaultServerName, &defaultIntegration{})
}
