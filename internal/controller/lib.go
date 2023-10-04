package controller

import (
	"context"
	webappv1 "demo1/api/v1"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *EntryReconciler) reconcileNodeSelector(ctx context.Context, parentResource *webappv1.Entry, l logr.Logger) (v1.Pod, error) {
	podName := parentResource.Name + "-pod"
	nodeOperator := v1.NodeSelectorOpIn
	podRes := &v1.Pod{}
	errInPodFinding := r.Get(ctx, types.NamespacedName{Name: podName, Namespace: parentResource.Namespace}, podRes)
	if errInPodFinding == nil {
		l.Info("Requested Node Selector already exists")
	}
	if !errors.IsNotFound(errInPodFinding) {
		return *podRes, errInPodFinding
	}
	podRes = &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: parentResource.Namespace,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "my-container",
					Image: "nginx:latest",
				},
			},
			Affinity: &v1.Affinity{
				NodeAffinity: &v1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
						NodeSelectorTerms: []v1.NodeSelectorTerm{
							{
								MatchExpressions: []v1.NodeSelectorRequirement{
									{
										Key:      "environment",
										Operator: nodeOperator,
										Values:   []string{"production"},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	l.Info("Creating node-selector...", "node-selector name", podRes.Name, "node-selector namespace", podRes.Namespace)

	if err := ctrl.SetControllerReference(parentResource, podRes, r.Scheme); err != nil {
		return *podRes, err
	}
	podResErr := r.Create(ctx, podRes)
	return *podRes, podResErr
}
