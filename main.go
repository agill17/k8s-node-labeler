package main

import (
	"agill.apps.node-labeler/pkg"
	"flag"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"os"
	"strconv"
	"time"
)

var (
	logger = logrus.New()
	defaultRsyncPeriod = "30" // minutes
	labelsToAdd = map[string]string{"foo": "bar"}
)



func init() {
	logger.SetFormatter(&logrus.JSONFormatter{})
	if _, ok := os.LookupEnv(pkg.EnvVarDevMode); ok {
		logger.SetFormatter(&logrus.TextFormatter{ForceColors: true})
	}
}


func main() {
	var configFile string
	flag.StringVar(&configFile, "conf-file", "", "Config yaml file that defines desired node labels")
	flag.Parse()

	// TODO: swap to cobra for required flags , but in the meantime this hack works
	if configFile == "" {
		logger.Errorf("--conf-file flag is required.")
		os.Exit(1)
	}

	if _, errCheckingFile := os.Stat(configFile); errCheckingFile != nil {
		logger.Errorf("Failed to check for config file: %v", errCheckingFile)
		os.Exit(1)
	}

	client := pkg.GetClient(logger)
	nodeReconciler := pkg.ReconcileNodeLabel{
		Logger:    logger,
		ClientSet: client,
	}

	// instantiate sharedInformerFactory
	sharedInformerFactory := informers.NewSharedInformerFactory(client, getResyncPeriod())

	// create informer for specific resource
	nodeInformer := sharedInformerFactory.Core().V1().Nodes()

	// register event handlers ( we only care about add and update events )
	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			//TODO: if err -> requeue
			nodeReconciler.ReconcileNode(obj, configFile)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			//TODO: if err -> requeue
			nodeReconciler.ReconcileNode(newObj, configFile)
		},
	})

	stopChan := make(chan struct{})

	// start informer
	logger.Info("Staring informers")
	sharedInformerFactory.Start(stopChan)

	// pre-heat informers cache store
	logger.Info("Waiting for caches to sync")
	sharedInformerFactory.WaitForCacheSync(stopChan)

	logger.Info("Started event handlers")
	// block main routine
	<- stopChan
}

// returns time.Duration resyncPeriod ( how often to run eventHandlers while using cached store )
// unit -> minutes
func getResyncPeriod() time.Duration {
	syncPeriodInStr := defaultRsyncPeriod
	if val, ok := os.LookupEnv(pkg.EnvVarResyncPeriod); ok {
		syncPeriodInStr = val
	}
	syncPeriod, errCastingToInt := strconv.Atoi(syncPeriodInStr)
	if errCastingToInt != nil {
		logger.Errorf("Failed to cast string to int in getRsyncPeriod: %v", errCastingToInt)
		os.Exit(1)
	}
	return time.Duration(syncPeriod) * time.Minute
}