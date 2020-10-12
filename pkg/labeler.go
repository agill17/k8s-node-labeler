package pkg

import (
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ReconcileNodeLabel struct {
	Logger *logrus.Logger
	ClientSet *kubernetes.Clientset
}

func (r *ReconcileNodeLabel) ReconcileNode (obj interface{}, desiredLabels map[string]string) error {
	nodeObjFromCache, isNode := obj.(*v1.Node)
	if !isNode {
		r.Logger.Warnf("Not a node object, therefore skipping: %T", obj)
		return errors.New("ErrNotANodeObj")
	}
	nodeObj := nodeObjFromCache.DeepCopy()

	currentLabels := nodeObj.GetLabels()
	var needsUpdate bool
	for k, v := range desiredLabels {
		if val, found := currentLabels[k]; !found || val != v {
			currentLabels[k] = v
			needsUpdate = true
		}
	}

	if needsUpdate {
		r.Logger.Infof("Updating labels for node: %v", nodeObj.GetName())
		nodeObj.SetLabels(currentLabels)
		if _, errUpdating := r.ClientSet.CoreV1().Nodes().Update(context.TODO(), nodeObj, metav1.UpdateOptions{}); errUpdating != nil {
			r.Logger.Errorf("Failed to add/update labels: %v", errUpdating)
			return errUpdating
		}

	}
	r.Logger.Infof("Node labels reconciled: %v", nodeObj.GetName())
	return nil
}
