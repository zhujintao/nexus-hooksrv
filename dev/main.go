package main

import (
	"github.com/gin-gonic/gin"
	"fmt"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"strings"
)


var nexus = make(map[string]interface{})

func setupRouter() *gin.Engine {

	r := gin.Default()


	r.POST("/nexus", func(c *gin.Context) {

		c.BindJSON(&nexus)

		if a :=nexus["asset"]; a != nil {

			asset:=a.(map[string]interface{})
			fmt.Println("Asset",asset["name"],nexus["action"] )

		}

		if c :=nexus["component"]; c != nil {

			component := c.(map[string]interface{})

			if nexus["action"].(string) == "CREATED" {

				var imagesource = os.Getenv("REGISTRY_ADDRESS") + "/" + component["name"].(string) + ":" + component["version"].(string)
				var namespace = strings.Split(component["version"].(string), "_")[0]
				kubeConfig, err := rest.InClusterConfig()
				if err != nil {
					panic(err.Error())
				}
				clientset, err := kubernetes.NewForConfig(kubeConfig)
				if err != nil {
					panic(err.Error())
				}

				deployClient := clientset.AppsV1().Deployments(namespace)

				result, err := deployClient.Get(component["name"].(string), metav1.GetOptions{})

				if err == nil {
					for k, v := range result.Spec.Template.Spec.Containers {

						if v.Name == component["name"].(string) {
							result.Spec.Template.Spec.Containers[k].Image = imagesource
							_, updateErr := deployClient.Update(result)

							fmt.Println("NameSpace: ", namespace, " Image:", imagesource, " updateErr: ", updateErr)
						}
					}
				}
			}

		}
	})

	return r
}



func main() {
	gin.SetMode(gin.ReleaseMode)

	r := setupRouter()

	r.Run(":8080")


}
