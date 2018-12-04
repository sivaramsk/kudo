/*

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

package instance

import (
	"context"
	"fmt"
	"log"

	maestrov1alpha1 "github.com/kubernetes-sigs/kubebuilder-maestro/pkg/apis/maestro/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const basePath = "/kustomize"

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Instance Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
// USER ACTION REQUIRED: update cmd/manager/main.go to call this maestro.Add(mgr) to install this Controller
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileInstance{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("instance-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	p := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {

			return e.ObjectOld != e.ObjectNew
		},
		//New Instances should have Deploy called
		CreateFunc: func(e event.CreateEvent) bool {
			fmt.Printf("Recieved create event for %v\n", e.Meta)
			gvk, _ := apiutil.GVKForObject(e.Object, mgr.GetScheme())

			// Create a new ref
			ref := corev1.ObjectReference{
				Kind:      gvk.Kind,
				Name:      e.Meta.GetName(),
				Namespace: e.Meta.GetNamespace(),
			}

			deploy := maestrov1alpha1.PlanExecution{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%v-deploy", e.Meta.GetName()),
					Namespace: e.Meta.GetNamespace(),
				},
				Spec: maestrov1alpha1.PlanExecutionSpec{
					Instance: ref,
					PlanName: "deploy",
				},
			}
			//Make this instance the owner of the PlanExecution
			controllerutil.SetControllerReference(e.Meta, &deploy, mgr.GetScheme())
			//new!
			err := mgr.GetClient().Create(context.TODO(), &deploy)
			if err != nil {
				fmt.Printf("Error creating deploy object for %v: %v\n", e.Meta.GetName(), err)
			}
			return err == nil
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return true
		},
	}

	// Watch for changes to Instance
	err = c.Watch(&source.Kind{Type: &maestrov1alpha1.Instance{}}, &handler.EnqueueRequestForObject{}, p)
	if err != nil {
		return err
	}
	return nil
}

var _ reconcile.Reconciler = &ReconcileInstance{}

// ReconcileInstance reconciles a Instance object
type ReconcileInstance struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Instance object and makes changes based on the state read
// and what is in the Instance.Spec
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=maestro.k8s.io,resources=instances,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileInstance) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the Instance instance
	instance := &maestrov1alpha1.Instance{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	log.Printf("InstanceController: Recieved Reconcile request for %v\n", request.Name)

	//defer call from above should apply the status changes to the object
	return reconcile.Result{}, nil
}