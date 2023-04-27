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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	k8sfrrv1alpha1 "github.com/metallb/frrk8s/api/v1alpha1"
	"github.com/metallb/frrk8s/internal/frr"
)

// FRRConfigurationReconciler reconciles a FRRConfiguration object
type FRRConfigurationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	FRR    *frr.FRR
}

//+kubebuilder:rbac:groups=k8sfrr.frr.metallb.io,resources=frrconfigurations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=k8sfrr.frr.metallb.io,resources=frrconfigurations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=k8sfrr.frr.metallb.io,resources=frrconfigurations/finalizers,verbs=update

func (r *FRRConfigurationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	configs := &k8sfrrv1alpha1.FRRConfigurationList{}
	err := r.Client.List(ctx, configs)
	if err != nil {
		return ctrl.Result{}, err
	}
	if len(configs.Items) == 0 {
		// TODO delete config
	}
	// TODO Implement merge
	if len(configs.Items) > 1 {
		logger.Info("More than one FRRConfiguration found, only one is supported")
		return ctrl.Result{}, nil
	}
	for _, config := range configs.Items {
		err = r.FRR.ApplyConfig(config)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *FRRConfigurationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&k8sfrrv1alpha1.FRRConfiguration{}).
		Complete(r)
}
