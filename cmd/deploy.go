package cmd

import (
	"context"
	errors2 "errors"
	"fmt"
	"github.com/chill-cloud/chill-cli/pkg/cache"
	"github.com/chill-cloud/chill-cli/pkg/cluster"
	"github.com/chill-cloud/chill-cli/pkg/config"
	cwd2 "github.com/chill-cloud/chill-cli/pkg/cwd"
	"github.com/chill-cloud/chill-cli/pkg/logging"
	"github.com/chill-cloud/chill-cli/pkg/service/naming"
	util "github.com/chill-cloud/chill-cli/pkg/util"
	"github.com/chill-cloud/chill-cli/pkg/version"
	"github.com/chill-cloud/chill-cli/pkg/version/set"
	"github.com/spf13/cobra"
	"github.com/werf/lockgate"
	"github.com/werf/lockgate/pkg/distributed_locker"
	"github.com/werf/werf/pkg/kubeutils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	"time"
)

func RunDeploy(cmd *cobra.Command, args []string) error {
	cwd, err := cwd2.SetupCwd(Cwd)
	if err != nil {
		return err
	}

	cfg, err := config.ParseConfig(cwd, config.LockConfigName, true)
	if err != nil {
		return err
	}

	s, err := cache.NewLocalSourceOfTruth(cwd)
	if err != nil {
		return err
	}

	clean, err := s.IsClean()

	if err != nil {
		return err
	}

	if clean != nil {
		return fmt.Errorf("uncommitted or untracked changes detected; commit or stash them")
	}

	if !forceFrozen {
		frozen, err := s.IsFrozen()

		if err != nil {
			return err
		}

		if !frozen {
			return fmt.Errorf("version must be frozen before deploying")
		}
	}

	clusterManager, err := cluster.NewForKubernetes(Kubeconfig, KubeNamespace)
	if err != nil {
		return fmt.Errorf("unable to build cluster client")
	}
	knative, err := clusterManager.GetKnative()
	if err != nil {
		return fmt.Errorf("unable to build Knative client")
	}
	ver := cfg.CurrentVersion
	name := fmt.Sprintf("%s-v%d", cfg.Name, ver.GetMajor())
	imageName, _ := cfg.GetBuildTag(ForceLocal)

	k8sClient, err := clusterManager.GetKubernetesClient()
	if err != nil {
		return err
	}

	configMapName := fmt.Sprintf("chill-lock-%s", name)

	_, err = kubeutils.GetOrCreateConfigMapWithNamespaceIfNotExists(k8sClient, KubeNamespace, configMapName)
	if err != nil {
		return err
	}

	lockerClient, err := clusterManager.GetLockerClient()

	if err != nil {
		return err
	}

	locker := distributed_locker.NewKubernetesLocker(
		lockerClient, schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "configmaps",
		},
		configMapName, KubeNamespace,
	)

	for i := 0; i < attemptsLimit; i++ {
		locked, h, err := locker.Acquire("lock", lockgate.AcquireOptions{Shared: false, Timeout: time.Duration(lockTimeout) * time.Second})
		if err != nil {
			return err
		}
		if locked {
			defer func() {
				err := locker.Release(h)
				if err != nil {
					panic(err)
				}
			}()
			existingService, err := knative.Services(KubeNamespace).Get(context.TODO(), name, metav1.GetOptions{})
			created := true
			if err != nil {
				var typedErr *errors.StatusError
				switch {
				case errors2.As(err, &typedErr):
					if typedErr.Status().Reason == metav1.StatusReasonNotFound {
						logging.Logger.Info("Service not created yet")
						created = false
					}
				default:
					return err
				}
			} else {
				logging.Logger.Info(fmt.Sprintf("Service %s present, version: %s", existingService.Name, existingService.ResourceVersion))
			}

			var trafficList []servingv1.TrafficTarget
			if version.IsProduction(*cfg.CurrentVersion) {
				targets, err := cfg.GetTrafficTargets()
				if err != nil {
					return err
				}
				trafficList = []servingv1.TrafficTarget{
					{
						LatestRevision: util.BoolPtr(true),
						Percent:        util.Int64Ptr(int64(targets[*cfg.CurrentVersion])),
						Tag:            fmt.Sprintf("v%d-%d", ver.GetMinor(), ver.GetPatch()),
					},
				}
				var versions []version.Version
				for _, t := range existingService.Status.RouteStatusFields.Traffic {
					var minor, patch int
					_, err := fmt.Sscanf(t.Tag, "v%d-%d", &minor, &patch)
					if err != nil {
						return err
					}
					versions = append(versions, version.Version{Major: cfg.CurrentVersion.GetMajor(), Minor: minor, Patch: patch})
					trafficList = append(trafficList, servingv1.TrafficTarget{
						RevisionName: t.RevisionName,
						Percent: util.Int64Ptr(int64(targets[version.Version{
							Major: cfg.CurrentVersion.GetMajor(),
							Minor: minor,
							Patch: patch,
						}])),
						ConfigurationName: t.ConfigurationName,
						Tag:               t.Tag,
					})
				}
				latest := set.ArrayVersionSet(versions).GetLatestProductionVersion()
				if !latest.MayBeNext(cfg.CurrentVersion) {
					println("Wrong deploying order; retry might help")
					continue
				}
			} else {
				trafficList = []servingv1.TrafficTarget{
					{
						LatestRevision: util.BoolPtr(true),
						Percent:        util.Int64Ptr(0),
						Tag:            fmt.Sprintf("v%d-%d", ver.GetMinor(), ver.GetPatch()),
					},
				}
				for _, t := range existingService.Status.RouteStatusFields.Traffic {
					trafficList = append(trafficList, servingv1.TrafficTarget{
						RevisionName:      t.RevisionName,
						Percent:           t.Percent,
						ConfigurationName: t.ConfigurationName,
						Tag:               t.Tag,
					})
				}
			}
			envList := []v1.EnvVar{
				{
					Name:  "CHILL_SELF_NAME",
					Value: cfg.Name,
				},
				{
					Name:  "CHILL_SELF_VERSION",
					Value: cfg.CurrentVersion.String(),
				},
			}
			for dep := range cfg.Dependencies {
				specificVersion := dep.GetSpecificVersion()
				if specificVersion == nil {
					return fmt.Errorf("specific version must be set for service %s", dep.GetName())
				}
				host, err := clusterManager.GetInternalServiceHost(dep.GetName(), *specificVersion, dep.GetVersion())
				if err != nil {
					return err
				}
				envList = append(envList, v1.EnvVar{
					Name:  naming.NameToEnv(dep.GetName()),
					Value: host,
				})
			}
			var volumes []v1.Volume
			var volumeMounts []v1.VolumeMount
			for _, s := range cfg.Secrets {
				envList = append(envList, v1.EnvVar{
					Name: naming.SecretToEnv(s),
					ValueFrom: &v1.EnvVarSource{
						SecretKeyRef: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{Name: s},
							Key:                  cluster.ChillSecretKey,
						},
					},
				})
				volumes = append(volumes, v1.Volume{
					Name: fmt.Sprintf("chill-secret-mount-%s", s),
					VolumeSource: v1.VolumeSource{
						Secret: &v1.SecretVolumeSource{
							SecretName: s,
						},
					},
				})
				volumeMounts = append(volumeMounts, v1.VolumeMount{
					Name:      fmt.Sprintf("chill-secret-mount-%s", s),
					MountPath: naming.SecretToMountPath(s),
				})
			}
			service := servingv1.Service{
				ObjectMeta: existingService.ObjectMeta,
				Spec: servingv1.ServiceSpec{
					RouteSpec: servingv1.RouteSpec{
						Traffic: trafficList,
					},
					ConfigurationSpec: servingv1.ConfigurationSpec{
						Template: servingv1.RevisionTemplateSpec{
							Spec: servingv1.RevisionSpec{
								PodSpec: v1.PodSpec{
									Volumes: volumes,
									Containers: []v1.Container{
										{
											Image:           imageName,
											ImagePullPolicy: v1.PullAlways,
											Ports: []v1.ContainerPort{
												{
													Name:          "h2c",
													Protocol:      v1.ProtocolTCP,
													ContainerPort: 80,
												},
											},
											Env:          envList,
											VolumeMounts: volumeMounts,
										},
									},
									ImagePullSecrets: []v1.LocalObjectReference{
										{
											Name: fmt.Sprintf("chill-reg-%s", cfg.Name),
										},
									},
								},
							},
						},
					},
				},
			}

			if created {
				service.SetResourceVersion(existingService.GetResourceVersion())
				service.ObjectMeta = existingService.ObjectMeta
				_, err = knative.Services(KubeNamespace).Update(context.TODO(), &service, metav1.UpdateOptions{})
				if err != nil {
					return fmt.Errorf("Knative server error while creating: %w\n", err)
				}
			} else {
				service.ObjectMeta = metav1.ObjectMeta{
					Name: name,
				}
				_, err = knative.Services(KubeNamespace).Create(context.TODO(), &service, metav1.CreateOptions{})

				if err != nil {
					return fmt.Errorf("Knative server error while updating: %w\n", err)
				}
			}
			return nil
		} else {
			logging.Logger.Info(fmt.Sprintf("Lock %d failed, waiting...", i))
			time.Sleep(time.Duration(sleepTime) * time.Second)
		}
	}
	return fmt.Errorf("Failed to take a lock\n")
}

var forceFrozen bool
var attemptsLimit int
var sleepTime int
var lockTimeout int

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploys your service into the cluster",
	RunE:  RunDeploy,
}

func init() {
	rootCmd.AddCommand(deployCmd)

	deployCmd.Flags().BoolVarP(&forceFrozen, "force-frozen", "f", false, "Force mark state as frozen")
	deployCmd.Flags().IntVar(&attemptsLimit, "attempts-limit", 5, "Number of tries to take a lock")
	deployCmd.Flags().IntVar(&sleepTime, "sleep-time", 10, "Timeout between tries to take a lock")
	deployCmd.Flags().IntVar(&lockTimeout, "lock-timeout", 30, "Lock timeout")
}
