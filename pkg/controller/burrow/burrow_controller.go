package burrow

import (
	"context"
	"fmt"
	monitorsv1beta1 "github.com/subravi92/burrow-operator/pkg/apis/monitors/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("burrow-controller")

// +kubebuilder:controller:group=monitors,version=v1beta1,kind=Burrow,resource=burrows

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Burrow Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
// USER ACTION REQUIRED: update cmd/manager/main.go to call this monitors.Add(mgr) to install this Controller
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileBurrow{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("burrow-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Burrow
	err = c.Watch(&source.Kind{Type: &monitorsv1beta1.Burrow{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create
	// Uncomment watch a Deployment created by Burrow - change this for objects you create
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &monitorsv1beta1.Burrow{},
	})
	if err != nil {
		return err
	}

	//watch  for configmap changes created by Burrow
	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForOwner{
		IsController: false,
		OwnerType:    &monitorsv1beta1.Burrow{},
	})
	if err != nil {
		return err
	}

	//watch for service changes created by burrow
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: false,
		OwnerType:    &monitorsv1beta1.Burrow{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileBurrow{}

// ReconcileBurrow reconciles a Burrow object
type ReconcileBurrow struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Burrow object and makes changes based on the state read
// and what is in the Burrow.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitors.aims.cisco.com,resources=burrows,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitors.aims.cisco.com,resources=burrows/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=monitors.aims.cisco.com,resources=burrows/finalizers,verbs=get;update;patch;delete;list;watch
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;watch;list
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;watch;list;create;update;delete
func (r *ReconcileBurrow) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the Burrow instance
	instance := &monitorsv1beta1.Burrow{}

	burrowerror := &burrowerror{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	configMap := NewConfigMap(*instance)

	if err != nil {
		return reconcile.Result{}, err

	}
	if err := controllerutil.SetControllerReference(instance, configMap, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	//delete this service and configmap
	found_cm := &corev1.ConfigMap{}

	validnamespace := isValidNamespace(instance.Namespace)

	if validnamespace {
		err = r.Get(context.TODO(), types.NamespacedName{Name: "burrow-config", Namespace: instance.Namespace}, found_cm)
		if err != nil && errors.IsNotFound(err) {
			log.Info("Creating Burrow Configmap", "namespace", configMap.Namespace, "name", configMap.Name)
			err = r.Create(context.TODO(), configMap)
			if err != nil {
				return reconcile.Result{}, err
			}
		} else if err != nil {
			return reconcile.Result{}, err

		}
		if !reflect.DeepEqual(configMap, found_cm) {
			found_cm = configMap
			log.Info("Updating Burrow Configmap", "namespace", configMap.Namespace, "name", configMap.Name)
			err = r.Update(context.TODO(), found_cm)
			if err != nil {
				return reconcile.Result{}, err
			}
		}

		//when configmap data changes

		if !reflect.DeepEqual(configMap.Data, found_cm.Data) {
			found_cm = configMap
			log.Info("Updating Burrow Configmap data", "namespace", configMap.Namespace, "name", configMap.Name)
			err = r.Update(context.TODO(), found_cm)
			if err != nil {
				return reconcile.Result{}, err
			}
		}

		deploy := NewDeployment(*instance)

		if err := controllerutil.SetControllerReference(instance, deploy, r.scheme); err != nil {
			return reconcile.Result{}, err
		}

		found := &appsv1.Deployment{}
		err = r.Get(context.TODO(), types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			log.Info("Creating Burrow Deployment", "namespace", deploy.Namespace, "name", deploy.Name)
			err = r.Create(context.TODO(), deploy)
			if err != nil {
				return reconcile.Result{}, err

			}
		} else if err != nil {
			return reconcile.Result{}, err

		}

		if !reflect.DeepEqual(deploy.Spec, found.Spec) {
			found.Spec = deploy.Spec
			log.Info("Updating Burrow Deployment ", "namespace", deploy.Namespace, "name", deploy.Name)
			err = r.Update(context.TODO(), found)

			if err != nil {
				return reconcile.Result{}, err
			}
		}

		//update when fields are changed via CRD
		//when burrow image changes
		if !reflect.DeepEqual(found.Spec.Template.Spec.Containers[0].Image, instance.Spec.BurrowImage) {

			found.Spec = deploy.Spec
			log.Info("Updating Burrow Deployment Image", "namespace", deploy.Namespace, "name", deploy.Name)
			err = r.Update(context.TODO(), found)
			if err != nil {
				return reconcile.Result{}, err
			}
		}

		//when exporter image changes
		if !reflect.DeepEqual(found.Spec.Template.Spec.Containers[1].Image, instance.Spec.ExporterImage) {

			found.Spec = deploy.Spec
			log.Info("Updating Burrow Exporter Deployment Image", "namespace", deploy.Namespace, "name", deploy.Name)
			err = r.Update(context.TODO(), found)
			if err != nil {
				return reconcile.Result{}, err
			}
		}

		//when apiport number changes

		if !reflect.DeepEqual(found.Spec.Template.Spec.Containers[1].Env[0].Value, "http://localhost:"+fmt.Sprint(instance.Spec.ApiPort)) {

			found.Spec = deploy.Spec
			log.Info("Updating Burrow Api Port ", "namespace", deploy.Namespace, "name", deploy.Name)
			err = r.Update(context.TODO(), found)
			if err != nil {
				return reconcile.Result{}, err
			}
		}

		//when metrics port changes

		if !reflect.DeepEqual(found.Spec.Template.Spec.Containers[1].Env[1].Value, "0.0.0.0:"+fmt.Sprint(instance.Spec.MetricsPort)) {

			found.Spec = deploy.Spec
			log.Info("Updating Burrow Metrics Port ", "namespace", deploy.Namespace, "name", deploy.Name)
			err = r.Update(context.TODO(), found)
			if err != nil {
				return reconcile.Result{}, err
			}
		}

		//when interval changes

		if !reflect.DeepEqual(found.Spec.Template.Spec.Containers[1].Env[2].Value, fmt.Sprint(instance.Spec.Interval)) {

			found.Spec = deploy.Spec
			log.Info("Updating interval in the deployment", "namespace", deploy.Namespace, "name", deploy.Name)
			err = r.Update(context.TODO(), found)
			if err != nil {
				return reconcile.Result{}, err
			}
		}

		//when apiversion changes

		if !reflect.DeepEqual(found.Spec.Template.Spec.Containers[1].Env[3].Value, fmt.Sprint(instance.Spec.ApiVersion)) {

			found.Spec = deploy.Spec
			log.Info("Updating api version in the deployment ", "namespace", deploy.Namespace, "name", deploy.Name)
			err = r.Update(context.TODO(), found)
			if err != nil {
				return reconcile.Result{}, err
			}
		}

		service := NewService(*instance)

		if err := controllerutil.SetControllerReference(instance, service, r.scheme); err != nil {
			return reconcile.Result{}, err
		}

		found_service := &corev1.Service{}

		err = r.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, found_service)
		if err != nil && errors.IsNotFound(err) {
			log.Info("creating Burrow Service", "namespace", service.Namespace, "name", service.Name)
			err = r.Create(context.TODO(), service)
			if err != nil {
				return reconcile.Result{}, err
			}
		} else if err != nil {
			return reconcile.Result{}, err

		}

		if !reflect.DeepEqual(service.Spec, found_service.Spec) {
			found_service.Spec = service.Spec
			log.Info("Updating Burrow Service", "namespace", service.Namespace, "name", service.Name)
			err = r.Update(context.TODO(), found)
			if err != nil {
				return reconcile.Result{}, err
			}
		}

		//when api port changes
		if !reflect.DeepEqual(service.Spec.Ports[0].Port, instance.Spec.ApiPort) {
			found_service.Spec = service.Spec
			log.Info("Api Port Changed...Updating Burrow Service", "namespace", service.Namespace, "name", service.Name)
			err = r.Update(context.TODO(), found)
			if err != nil {
				return reconcile.Result{}, err
			}
		}

		return reconcile.Result{}, nil

	}

	burrowerror.error = instance.Namespace

	burrowerror.InvalidNamespaceError()

	return reconcile.Result{}, nil

}
