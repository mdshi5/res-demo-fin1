/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	webappv1 "demo1/api/v1"
)

// EntryReconciler reconciles a Entry object
type EntryReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=webapp.shi.io,resources=entries,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=webapp.shi.io,resources=entries/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=webapp.shi.io,resources=entries/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Entry object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.0/pkg/reconcile
func (r *EntryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	myFinalizerName := "batch.tutorial.kubebuilder.io/finalizer"
	entry := &webappv1.Entry{}
	errfind := r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, entry)
	if errfind == nil {
		l.Info("resource entry found")
	}

	nodeRes, err := r.reconcileNodeSelector(ctx, entry, l)
	if err != nil {
		l.Info("error in node creation")
	}

	l.Info("resource node created", "name:", nodeRes.Name, "namespace:", nodeRes.Namespace)
	// examine DeletionTimestamp to determine if object is under deletion
	if entry.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !controllerutil.ContainsFinalizer(entry, myFinalizerName) {
			controllerutil.AddFinalizer(entry, myFinalizerName)
			if err := r.Update(ctx, entry); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(entry, myFinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			if err := r.deleteExternalResources(nodeRes); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(entry, myFinalizerName)
			if err := r.Update(ctx, entry); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EntryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webappv1.Entry{}).
		Complete(r)
}
