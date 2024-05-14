package promYamlGen

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func readFile(name string) []byte {
	data, err := os.ReadFile(name)
	if err != nil {
		log.Panicf("Failed to read file. Error: %s", err)
	}

	return data
}

func CreateConfig(promTemplateFile string) {
	data := config{}

	promTemplateData := readFile(promTemplateFile)
	err := yaml.Unmarshal(promTemplateData, &data)
	if err != nil {
		fmt.Printf("Error %s:", err.Error())
		os.Exit(1)
	}

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		fmt.Println("Error kubeConfig", err)
		os.Exit(0)
	}

	client, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		fmt.Println("Error", err)
		os.Exit(0)
	}
	services, err := client.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{})
	for _, service := range services.Items {
		if strings.Contains(service.Name, "dl-node")  && strings.Contains(service.Name, "metrics") {
			nodeExporter := service.Name + ":9100"
			for _, portObj := range service.Spec.Ports {
				if portObj.Name == "metrics" {
					socket := service.Name + ":" + fmt.Sprintf("%v", portObj.Port)
					staticConfig := StaticConfig{
						Targets: []string{socket, nodeExporter},
					}
					config := ScrapeConfig{
						JobName:       service.Name,
						StaticConfigs: []StaticConfig{staticConfig},
					}
					data.ScrapeConfigs = append(data.ScrapeConfigs, config)
				}
			}

		}

		if strings.Contains(service.Name, "dl-disperser") && strings.Contains(service.Name, "metrics") {
			nodeExporter := service.Name + ":9100"
			for _, portObj := range service.Spec.Ports {
				if portObj.Name == "metrics" {
					socket := service.Name + ":" + fmt.Sprintf("%v", portObj.Port)
					staticConfig := StaticConfig{
						Targets: []string{socket, nodeExporter},
					}
					config := ScrapeConfig{
						JobName:       service.Name,
						StaticConfigs: []StaticConfig{staticConfig},
					}
					data.ScrapeConfigs = append(data.ScrapeConfigs, config)

				}
			}
		}
	}

	//// Write to subgraph file
	subgraphYaml, err := yaml.Marshal(&data)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println(string(subgraphYaml))
	//writeFile(subgraphFile, subgraphYaml)
	//log.Print("subgraph.yaml written")

}
