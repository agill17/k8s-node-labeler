package main

import (
	"agill.apps.node-labeler/pkg"
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

const (
	EnvVarResyncPeriod = "RESYNC_PERIOD"
	EnvVarDevMode = "DEV_MODE"
)

func init() {
	logger.SetFormatter(&logrus.JSONFormatter{})
	if _, ok := os.LookupEnv(EnvVarDevMode); ok {
		logger.SetFormatter(&logrus.TextFormatter{ForceColors: true})
	}
}


func main() {

	// TODO: take input that defines set of labels to add to node
	// TODO: take input to exclude nodes with k:v as labels

	client := pkg.GetClient(logger)

	// instantiate sharedInformerFactory
	sharedInformerFactory := informers.NewSharedInformerFactory(client, getResyncPeriod())

	// create informer for specific resource
	nodeInformer := sharedInformerFactory.Core().V1().Nodes()

	nodeReconciler := pkg.ReconcileNodeLabel{
		Logger:    logger,
		ClientSet: client,
	}

	// register event handlers ( we only care about add and update events )
	nodeInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			//TODO: if err -> requeue
			nodeReconciler.ReconcileNode(obj, labelsToAdd)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			//TODO: if err -> requeue
			nodeReconciler.ReconcileNode(newObj, labelsToAdd)
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
	if val, ok := os.LookupEnv(EnvVarResyncPeriod); ok {
		syncPeriodInStr = val
	}
	syncPeriod, errCastingToInt := strconv.Atoi(syncPeriodInStr)
	if errCastingToInt != nil {
		logger.Errorf("Failed to cast string to int in getRsyncPeriod: %v", errCastingToInt)
		os.Exit(1)
	}
	return time.Duration(syncPeriod) * time.Minute
}