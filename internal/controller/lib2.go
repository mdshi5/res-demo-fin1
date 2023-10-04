package controller

import (
	"context"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (r *EntryReconciler) deleteExternalResources(pod v1.Pod) error {

	err := r.Client.Delete(context.TODO(), &pod)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	return nil
}
