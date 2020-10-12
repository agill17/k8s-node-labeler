package pkg

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/workqueue"
	"os"
	"reflect"
)

type ReconcileNodeLabel struct {
	Logger *logrus.Logger
	ClientSet *kubernetes.Clientset
	Queue workqueue.DelayingInterface
}

func (r *ReconcileNodeLabel) ReconcileNode (obj interface{}, confFile string) error {
	configObj, errParsingConf := NewDesiredConf(confFile)
	if errParsingConf != nil {
		r.Logger.Errorf("Failed to parse config file: %v", errParsingConf)
		os.Exit(1)
	}

	nodeObjFromCache, isNode := obj.(*v1.Node)
	if !isNode {
		r.Logger.Warnf("Not a node object, therefore skipping: %T", obj)
		return errors.New("ErrNotANodeObj")
	}
	nodeObj := nodeObjFromCache.DeepCopy()
	currentLabels := nodeObj.GetLabels()
	currentAnnotations := nodeObj.GetAnnotations()
	desiredLabels := configObj.DesiredLabels

	// this represents labels that were applied as "desired state". Meaning they were desired at some point.
	var appliedDesiredLabels map[string]string
	if val, ok := currentAnnotations[LastAppliedLabelsAnnotationKey]; ok {
		err := json.Unmarshal([]byte(val), &appliedDesiredLabels)
		if err != nil {
			r.Logger.Errorf("Failed to unmarshall applied desired label annotations to a map: %v", err)
			return err
		}
	}

	// a user could remove a label from conf, we need to reconcile that delete from node labels
	// in-case new desiredLabels do not match what was applied as "desired labels", we need to reconcile that
	if !reflect.DeepEqual(appliedDesiredLabels, desiredLabels) {
		labelKeysToDelete := getLabelsToRemove(appliedDesiredLabels, desiredLabels)
		for _, e := range labelKeysToDelete {
			delete(currentLabels, e)
		}
	}


	for k, v := range desiredLabels {
		if val, found := currentLabels[k]; !found || val != v {
			currentLabels[k] = v
		}
	}

	if !reflect.DeepEqual(currentLabels, nodeObjFromCache.GetLabels()) {
		r.Logger.Infof("Updating labels for node: %v", nodeObj.GetName())
		currentDesiredLabels, err := json.Marshal(desiredLabels)
		if err != nil {
			r.Logger.Errorf("Failed to convert to json: %v", err)
			return err
		}
		currentAnnotations[LastAppliedLabelsAnnotationKey] = string(currentDesiredLabels)
		nodeObj.SetAnnotations(currentAnnotations)
		nodeObj.SetLabels(currentLabels)

		if _, errUpdating := r.ClientSet.CoreV1().Nodes().Update(context.TODO(), nodeObj, metav1.UpdateOptions{}); errUpdating != nil {
			r.Logger.Errorf("Failed to add/update labels: %v", errUpdating)
			return errUpdating
		}
	}
	r.Logger.Infof("Node labels reconciled: %v", nodeObj.GetName())
	return nil
}


func getLabelsToRemove(appliedDesiredLabels map[string]string, desiredLabels map[string]string) []string {
	var labelsToDelete []string
	for k, _ := range appliedDesiredLabels {
		if _, found := desiredLabels[k]; !found {
			labelsToDelete = append(labelsToDelete, k)
		}
	}
	return labelsToDelete
}