package server

const DefaultServerName = "default"

type Integration interface {
	GenerateMethods(cwd string, name string, protoSource string) error
	GetBaseProjectRemote() string
}

var serverIntegrationMap = map[string]Integration{}

func ForName(name string) Integration {
	return serverIntegrationMap[name]
}

func Register(name string, integration Integration) {
	serverIntegrationMap[name] = integration
}
