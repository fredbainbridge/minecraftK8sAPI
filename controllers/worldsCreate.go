package controllers

import (
	"context"
	"errors"
	"fmt"
	"main/initializers"
	"main/k8sRepo"
	"main/models"
	"main/requests"
	"math/rand"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func WorldsCreate(c *gin.Context) {
	var err error
	var request requests.WorldRequest
	c.Bind(&request)
	worldName := strings.Replace(strings.ToLower(request.Name), " ", "-", -1)
	claimName := fmt.Sprintf("%s-claim", worldName)
	driveNum := rand.Int() % 2
	drive := strings.Split(os.Getenv("VOL_DRIVES"), ",")[driveNum]
	osDrive := fmt.Sprintf("%s:", drive)
	localFolderPath := path.Join(osDrive, os.Getenv("VOL_DATAFOLDER"), worldName)

	//generate vol path
	basepath := os.Getenv("VOL_BASE_PATH")
	rand.Seed(time.Now().UnixNano())
	hostPath := path.Join(basepath, drive, os.Getenv("VOL_DATAFOLDER"), worldName)
	err = volPathExists(hostPath)
	if err != nil {
		nukeAll(worldName, claimName, localFolderPath)
		c.Status(400)
		return
	}

	//create folder
	_, err = os.Stat(localFolderPath)
	if os.IsNotExist(err) {
		err = os.Mkdir(localFolderPath, 0755)
		if err != nil {
			nukeAll(worldName, claimName, localFolderPath)
			c.Status(400)
			return
		}
	} else {
		nukeAll(worldName, claimName, localFolderPath)
		c.Status(400)
		return
	}

	//create volume
	_, err = k8sRepo.CreateVolume(worldName, hostPath)
	if err != nil {
		nukeAll(worldName, claimName, localFolderPath)
		c.Status(400)
		return
	}

	//create claim
	storageClassName := "local-storage"
	err = k8sRepo.CreatePersistentVolumeClaim(claimName, storageClassName)
	if err != nil {
		nukeAll(worldName, claimName, localFolderPath)
		c.Status(400)
		return
	}

	//create service
	_, err = k8sRepo.CreateService(worldName, request.Port)
	if err != nil {
		nukeAll(worldName, claimName, localFolderPath)
		c.Status(400)
		return
	}
	_, err = k8sRepo.CreateDeployment(request, worldName, claimName, os.Getenv("IMAGE_NAME"))
	if err != nil {
		nukeAll(worldName, claimName, localFolderPath)
		c.Status(400)
	}

	mTags := []models.Tag{}
	for _, rTag := range request.Tags {
		mTags = append(mTags, models.Tag{
			Key:   rTag.Key,
			Value: rTag.Value,
		})
	}

	volumes := []models.Volume{}
	volumes = append(volumes, models.Volume{
		HostPath:  hostPath,
		LocalPath: localFolderPath,
		Storage:   "20gi",
		Claim:     claimName,
	})
	world := models.World{
		Name:    request.Name,
		K8sName: worldName,
		Port:    request.Port,
		Tags:    mTags,
		Volumes: volumes,
	}
	result := initializers.DB.Create(&world)

	if result.Error != nil {
		c.Status(400)
		return
	}
	c.JSON(200, world)
}

func nukeAll(worldName string, claimName string, localFolderPath string) {
	var err error
	err = k8sRepo.DeleteDeployment(worldName)
	if err != nil {
		fmt.Println(err)
	}
	err = k8sRepo.DeletePersistentVolume(worldName)
	if err != nil {
		fmt.Println(err)
	}
	err = k8sRepo.DeletePersistentVolumeClaim(claimName)
	if err != nil {
		fmt.Println(err)
	}
	err = k8sRepo.DeleteService(worldName)
	if err != nil {
		fmt.Println(err)
	}
	err = os.RemoveAll(localFolderPath)
	if err != nil {
		fmt.Println(err)
	}
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
