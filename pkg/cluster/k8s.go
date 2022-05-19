package cluster

import (
	"context"
	"encoding/base64"
	"encoding/json"
	errors2 "errors"
	"fmt"
	"github.com/chill-cloud/chill-cli/pkg/logging"
	"github.com/chill-cloud/chill-cli/pkg/service/naming"
	"github.com/chill-cloud/chill-cli/pkg/version"
	"github.com/chill-cloud/chill-cli/pkg/version/constraint"
	v13 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	knative "knative.dev/serving/pkg/client/clientset/versioned"
	v1 "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	ChillSecretKey = "chill-secret"
)

type kubernetesClusterManager struct {
	Config    *rest.Config
	Namespace string
}

func (k *kubernetesClusterManager) GetServiceIdentifier(serviceName string, version version.Version) string {
	return fmt.Sprintf("%s-v%d", serviceName, version.GetMajor())
}

func (k *kubernetesClusterManager) GetRevisionPath(serviceName string, version version.Version) string {
	return fmt.Sprintf("v%d-%d-%s-v%d", version.GetMinor(), version.GetPatch(), serviceName, version.GetMajor())
}

func (k *kubernetesClusterManager) GetServiceAndVersion(revisionPath string) (string, *version.Version, error) {
	parts := strings.Split(revisionPath, "-")
	if len(parts) < 4 {
		return "", nil, fmt.Errorf("too few parts")
	}
	if !strings.HasPrefix(parts[0], "v") {
		return "", nil, fmt.Errorf("wrong minor version format")
	}
	minor, err := strconv.Atoi(strings.TrimPrefix(parts[0], "v"))
	if err != nil {
		return "", nil, err
	}
	patch, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", nil, err
	}

	if !strings.HasPrefix(parts[len(parts)-1], "v") {
		return "", nil, fmt.Errorf("wrong major version format")
	}

	major, err := strconv.Atoi(strings.TrimPrefix(parts[len(parts)-1], "v"))
	if err != nil {
		return "", nil, err
	}

	name := naming.MergeToCanonical(parts[2 : len(parts)-1])
	return name, &version.Version{Major: major, Minor: minor, Patch: patch}, nil
}

func (k *kubernetesClusterManager) GetInternalServiceHost(serviceName string, v version.Version, c constraint.Constraint) (string, error) {
	var path string
	if annotated, ok := c.(*constraint.AnnotatedConstraint); ok {
		if ranged, ok := annotated.C.(*constraint.RangedConstraint); ok {
			if ranged.Upper.GetMajor() > ranged.Lower.GetMajor() {
				path = k.GetServiceIdentifier(serviceName, v)
			}
		}
	}
	if path == "" {
		path = k.GetRevisionPath(serviceName, v)
	}
	return fmt.Sprintf(
		"%s.%s.svc.cluster.local",
		path,
		k.Namespace,
	), nil
}

func (k *kubernetesClusterManager) GetKnative() (v1.ServingV1Interface, error) {
	deploymentsClient, err := knative.NewForConfig(k.Config)
	if err != nil {
		return nil, err
	}
	return deploymentsClient.ServingV1(), nil
}

func (k *kubernetesClusterManager) setSecret(key string, value string, mapKey string, secretType v13.SecretType) error {
	kubeClient, err := kubernetes.NewForConfig(k.Config)
	if err != nil {
		return err
	}
	secretsInterface := kubeClient.CoreV1().Secrets(k.Namespace)
	secret, err := secretsInterface.Get(context.TODO(), key, v12.GetOptions{})
	created := true
	if err != nil {
		var typedError *errors.StatusError
		switch {
		case errors2.As(err, &typedError):
			if typedError.Status().Reason == v12.StatusReasonNotFound {
				logging.Logger.Info("Service not created yet")
				created = false
			}
		default:
			return err
		}
	}
	var s *v13.Secret
	if created {
		s = secret
		s.StringData = map[string]string{mapKey: value}
		_, err := secretsInterface.Update(context.TODO(), s, v12.UpdateOptions{})
		if err != nil {
			return err
		}
	} else {
		s = &v13.Secret{
			ObjectMeta: v12.ObjectMeta{
				Name: key,
			},
			StringData: map[string]string{mapKey: value},
			Type:       secretType,
		}
		_, err := secretsInterface.Create(context.TODO(), s, v12.CreateOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (k *kubernetesClusterManager) SetSecret(key string, value string) error {
	return k.setSecret(key, value, ChillSecretKey, v13.SecretTypeOpaque)
}

func serviceRegistryKey(serviceName string) string {
	return fmt.Sprintf("chill-reg-%s", serviceName)
}

func (k *kubernetesClusterManager) SetRegistry(serviceName string, server string, login string, password string) error {
	t := base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", login, password)))
	s := fmt.Sprintf("{\"auths\":{\"%s\":{\"auth\":\"%s\"}}}", server, t)

	err := k.setSecret(serviceRegistryKey(serviceName), s, ".dockerconfigjson", v13.SecretTypeDockerConfigJson)
	if err != nil {
		return err
	}
	return nil
}

func (k *kubernetesClusterManager) GetRegistry(serviceName string, server string) (string, string, error) {
	v, err := k.getSecret(serviceRegistryKey(serviceName), ".dockerconfigjson")
	if err != nil {
		return "", "", err
	}

	var s struct {
		Auths map[string]struct{ Auth string }
	}

	err = json.Unmarshal([]byte(v), &s)
	if err != nil {
		return "", "", err
	}
	res, err := base64.URLEncoding.DecodeString(s.Auths[server].Auth)
	if err != nil {
		return "", "", err
	}
	parts := strings.Split(string(res), ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("wrong Kubernetes registry value format")
	}
	return parts[0], parts[1], nil
}

func (k *kubernetesClusterManager) getSecret(key string, mapKey string) (string, error) {
	kubeClient, err := kubernetes.NewForConfig(k.Config)
	if err != nil {
		return "", err
	}
	secretsInterface := kubeClient.CoreV1().Secrets(k.Namespace)
	secret, err := secretsInterface.Get(context.TODO(), key, v12.GetOptions{})
	if err != nil {
		return "", err
	}
	return string(secret.Data[mapKey]), nil
}

func (k *kubernetesClusterManager) GetSecret(key string) (string, error) {
	return k.getSecret(key, ChillSecretKey)
}

func GetKubeconfigPath(forceConfig string) (string, error) {
	if forceConfig == "" {
		dirname, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		forceConfig = filepath.Join(dirname, ".kube", "config")
	}
	return forceConfig, nil
}

func (k *kubernetesClusterManager) GetLockerClient() (dynamic.Interface, error) {
	kubeClient, err := dynamic.NewForConfig(k.Config)
	if err != nil {
		return nil, err
	}
	return kubeClient, nil
}

func (k *kubernetesClusterManager) GetKubernetesClient() (*kubernetes.Clientset, error) {
	kubeClient, err := kubernetes.NewForConfig(k.Config)
	if err != nil {
		return nil, err
	}
	return kubeClient, nil
}

func NewForKubernetes(forceConfig string, namespace string) (ClusterManager, error) {
	forceConfig, err := GetKubeconfigPath(forceConfig)
	if err != nil {
		return nil, err
	}
	config, err := clientcmd.BuildConfigFromFlags("", forceConfig)
	if err != nil {
		return nil, err
	}
	return &kubernetesClusterManager{
		Config:    config,
		Namespace: namespace,
	}, nil
}
