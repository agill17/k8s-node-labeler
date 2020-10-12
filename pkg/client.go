package pkg

import (
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)


func GetClient(logger *logrus.Logger) *kubernetes.Clientset {
	clientSet, errGettingClient := kubernetes.NewForConfig(getCfg(logger))
	if errGettingClient != nil {
		logger.Errorf("Failed to get clientSet: %v", errGettingClient)
		os.Exit(1)
	}
	return clientSet
}

func getCfg(logger *logrus.Logger) *rest.Config {
	cfg, errGettingCfg := config.GetConfig()
	if errGettingCfg != nil {
		logger.Errorf("Failed to get config: %v", errGettingCfg)
		os.Exit(1)
	}
	return cfg
}