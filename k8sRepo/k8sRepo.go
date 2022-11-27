package k8sRepo

import (
	"context"
	"main/initializers"
	"main/requests"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateVolume(worldName string, path string) (*v1.PersistentVolume, error) {
	vol := v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: worldName,
		},
		Spec: v1.PersistentVolumeSpec{
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): resource.MustParse("20Gi"),
			},
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},
			PersistentVolumeReclaimPolicy: v1.PersistentVolumeReclaimDelete,
			StorageClassName:              "local-storage",
			PersistentVolumeSource: v1.PersistentVolumeSource{
				Local: &v1.LocalVolumeSource{
					Path: path,
				},
			},
			NodeAffinity: &v1.VolumeNodeAffinity{
				Required: &v1.NodeSelector{
					NodeSelectorTerms: []v1.NodeSelectorTerm{
						{
							MatchExpressions: []v1.NodeSelectorRequirement{
								{
									Key:      "kubernetes.io/hostname",
									Operator: v1.NodeSelectorOpIn,
									Values:   []string{"docker-desktop"},
								},
							},
						},
					},
				},
			},
		},
	}
	return initializers.ClientSet.CoreV1().PersistentVolumes().Create(context.TODO(), &vol, metav1.CreateOptions{})
}

func DeletePersistentVolume(name string) error {
	deletePolicy := metav1.DeletePropagationForeground
	return initializers.ClientSet.CoreV1().PersistentVolumes().Delete(context.TODO(), name, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
}

func CreatePersistentVolumeClaim(claimName string, storageClassName string) error {
	claimReq := v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{Kind: "PersistentVolumeClaim"},
		ObjectMeta: metav1.ObjectMeta{
			Name: claimName,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			StorageClassName: &storageClassName,
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceName(v1.ResourceStorage): resource.MustParse("20Gi"),
				},
			},
		},
	}
	_, err := initializers.ClientSet.CoreV1().PersistentVolumeClaims("default").Create(context.TODO(), &claimReq, metav1.CreateOptions{})
	return err
}

func DeletePersistentVolumeClaim(claimName string) error {
	deletePolicy := metav1.DeletePropagationForeground
	return initializers.ClientSet.CoreV1().PersistentVolumeClaims("default").Delete(context.TODO(), claimName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
}

func CreateService(worldName string, port int) (*v1.Service, error) {
	service := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: worldName,
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeNodePort,
			Ports: []v1.ServicePort{
				{
					Port:     25565,
					NodePort: int32(port),
				},
			},
			Selector: map[string]string{
				"app": worldName,
			},
		},
	}
	return initializers.ClientSet.CoreV1().Services("default").Create(context.TODO(), &service, metav1.CreateOptions{})
}

func DeleteService(name string) error {
	deletePolicy := metav1.DeletePropagationForeground
	return initializers.ClientSet.CoreV1().Services("default").Delete(context.TODO(), name, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
}

func CreateDeployment(request requests.WorldRequest, name string, claimName string, image string) (*appsv1.Deployment, error) {
	var envs []v1.EnvVar
	for _, tag := range request.Tags {
		envs = append(envs, v1.EnvVar{
			Name:  tag.Key,
			Value: tag.Value,
		})
	}

	deployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Image: image,
							Name:  name,
							Env:   envs,
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 25565,
									Name:          "main",
								},
							},
							ReadinessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									Exec: &v1.ExecAction{
										Command: []string{"/usr/local/bin/mc-monitor", "status", "--host", "localhost"},
									},
								},
								InitialDelaySeconds: 20,
								PeriodSeconds:       5,
								FailureThreshold:    20,
							},
							LivenessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									Exec: &v1.ExecAction{
										Command: []string{"/usr/local/bin/mc-monitor", "status", "--host", "localhost"},
									},
								},
							},
							VolumeMounts: []v1.VolumeMount{
								{
									MountPath: "/data",
									Name:      "mc-data",
								},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: "mc-data",
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: claimName,
								},
							},
						},
					},
				},
			},
		},
	}
	return initializers.ClientSet.AppsV1().Deployments("default").Create(context.TODO(), &deployment, metav1.CreateOptions{})
}

func DeleteDeployment(name string) error {
	deletePolicy := metav1.DeletePropagationForeground
	return initializers.ClientSet.AppsV1().Deployments("default").Delete(context.TODO(), name, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
}
