package configuration

import (
	"log"
	"os"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// K8sConfig tries returns local k8s config or incluster config
func K8sConfig(configPath string) (*rest.Config, error) {
	if configPath == "" {
		log.Print("Config path is not set, using inclusterconfig")
		return rest.InClusterConfig()
	}

	if _, err := os.Stat(configPath); err != nil {
		log.Printf("Unable to find config path at '%s', using inclusterconfig", configPath)
		return rest.InClusterConfig()
	}

	log.Printf("Using connfigPath '%s' to connect to cluster", configPath)
	return clientcmd.BuildConfigFromFlags("", configPath)
}
