package cluster

import (
	"github.com/chill-cloud/chill-cli/pkg/version"
	"github.com/chill-cloud/chill-cli/pkg/version/constraint"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	v1 "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1"
)

type ClusterManager interface {
	GetKnative() (v1.ServingV1Interface, error)
	GetKubernetesClient() (*kubernetes.Clientset, error)
	GetLockerClient() (dynamic.Interface, error)
	GetServiceIdentifier(serviceName string, version version.Version) string
	GetRevisionPath(serviceName string, version version.Version) string
	GetServiceAndVersion(revisionPath string) (string, *version.Version, error)
	// GetInternalServiceHost returns best host compatible with version matching constraints
	GetInternalServiceHost(serviceName string, version version.Version, c constraint.Constraint) (string, error)
	SetSecret(key string, value string) error
	GetSecret(key string) (string, error)
	SetRegistry(serviceName string, server string, login string, password string) error
	GetRegistry(serviceName string, server string) (string, string, error)
}
