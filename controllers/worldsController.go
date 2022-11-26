package controllers

import (
	"context"
	"errors"
	"fmt"
	"main/initializers"
	"main/models"
	"math/rand"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//appsv1 "k8s.io/api/apps/v1"
)

type WorldRequest struct {
	Name string `json:"name,omitempty"`
	Port int    `json:"port,omitempty"`
	Tags []Tag  `json:"tags,omitempty"`
}

type Tag struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

func WorldsCreate(c *gin.Context) {
	var err error
	var request WorldRequest
	c.Bind(&request)
	worldName := strings.Replace(strings.ToLower(request.Name), " ", "-", -1)
	claimName := fmt.Sprintf("%s-claim", worldName)

	//generate vol path
	basepath := os.Getenv("VOL_BASE_PATH")
	rand.Seed(time.Now().UnixNano())
	driveNum := rand.Int() % 2
	drive := strings.Split(os.Getenv("VOL_DRIVES"), ",")[driveNum]
	volPath := path.Join(basepath, drive, os.Getenv("VOL_DATAFOLDER"), worldName)
	err = volPathExists(volPath)
	if err != nil {
		c.Status(400)
		return
	}

	//create folder
	osDrive := fmt.Sprintf("%s:", drive)
	localFolderPath := path.Join(osDrive, os.Getenv("VOL_DATAFOLDER"), worldName)
	_, err = os.Stat(localFolderPath)
	if os.IsNotExist(err) {
		err = os.Mkdir(localFolderPath, 0755)
		if err != nil {
			c.Status(400)
			return
		}
	} else {
		c.Status(400)
		return
	}

	//create volume
	_, err = createVolume(worldName, volPath)
	if err != nil {
		c.Status(400)
		return
	}

	//create claim
	storageClassName := "local-storage"
	err = createClaim(claimName, storageClassName, err)
	if err != nil {
		c.Status(400)
		return
	}

	//create service
	_, err = createService(worldName, request.Port)
	if err != nil {
		c.Status(400)
		return
	}
	// deployment := &appsv1.Deployment{
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name: worldName,
	// 	},
	// 	Spec: appsv1.DeploymentSpec{
	// 		Strategy.Type
	// 	},
	// }
	//save records

	world := models.World{
		Name: "Test",
		Port: 30005,
		Tags: []models.Tag{
			{
				Key:   "key1",
				Value: "value1",
			},
			{
				Key:   "key2",
				Value: "value2",
			}},
	}
	result := initializers.DB.Create(&world)

	if result.Error != nil {
		c.Status(400)
		return
	}
	//check node exists
	c.JSON(200, world)
}

func createService(worldName string, port int) (*v1.Service, error) {
	service := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: worldName,
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeNodePort,
			Ports: []v1.ServicePort{
				v1.ServicePort{
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

func createClaim(claimName string, storageClassName string, err error) error {
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
	_, err = initializers.ClientSet.CoreV1().PersistentVolumeClaims("default").Create(context.TODO(), &claimReq, metav1.CreateOptions{})
	return err
}

func createVolume(worldName string, path string) (*v1.PersistentVolume, error) {
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

func volPathExists(path string) error {
	volList, err := initializers.ClientSet.CoreV1().PersistentVolumes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, vol := range volList.Items {
		if vol.Spec.StorageClassName == "local-storage" {
			if vol.Spec.PersistentVolumeSource.Local.Path == path {
				return errors.New("volume path already exists in kubernetes")
			}
		}
	}
	return nil
}
