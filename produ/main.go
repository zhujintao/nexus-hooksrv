package main

import (
	"github.com/gin-gonic/gin"
	"fmt"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/api/core/v1"
	"strings"
	"os"
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

			component:=c.(map[string]interface{})

			if nexus["action"].(string) == "CREATED" {
				if strings.Contains(component["name"].(string),"-release"){
					var imagesource = os.Getenv("REGISTRY_ADDRESS") + "/" + component["name"].(string)  + ":" +  component["version"].(string)

					kubeConfig, err := rest.InClusterConfig()
					if err != nil {
						panic(err.Error())
					}
					clientset, err := kubernetes.NewForConfig(kubeConfig)
					if err != nil {
						panic(err.Error())
					}

					pod := &apiv1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name: component["name"].(string),
						},
						Spec: apiv1.PodSpec{
							Containers: []apiv1.Container{
								{
									Name: component["name"].(string),
									Image: imagesource,
								},
							},
							ImagePullSecrets: []apiv1.LocalObjectReference{
								{ Name: os.Getenv("REGISTRY_SECRET") },
							},
							RestartPolicy: "Never",
						},

					}

					_,createErr:=clientset.CoreV1().Pods("default").Create(pod)
					fmt.Println(component["name"].(string)  + ":" +  component["version"].(string), "createErr:",createErr)

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
