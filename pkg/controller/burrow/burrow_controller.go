package burrow

import (
	//"bytes"
	"context"
	"fmt"
	"io/ioutil"

	//"io/ioutil"
	//"os"

	//"fmt"
	//"os"

	//"text/template"
	monitorsv1beta1 "burrow-operator/pkg/apis/monitors/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"log"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	//"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

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
// +kubebuilder:rbac:groups=monitors.aims.cisco.com,resources=burrows,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileBurrow) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the Burrow instance
	instance := &monitorsv1beta1.Burrow{}

	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	b, err := ioutil.ReadFile("template/burrow.toml")
	if err != nil {
		fmt.Print(err)
	}
	log.Printf("Checking status of resource %s", instance.Name)
	log.Printf("zkserver:%s", instance.Spec.Zookeeper.Servers)

	log.Printf("BurrowImage:%s", instance.Spec.BurrowImage)
	log.Printf("BurrowImage:%s", instance.Spec.ExporterImage)

	//con:=new(Config)
	//con:=populateConfig()

	log.Printf(":%s", b)

	configMap := NewConfigMap(*instance)

	configMap.Data = map[string]string{
		"burrow.toml": string(b),
	}
	if err != nil {
		return reconcile.Result{}, err

	}
	if err := controllerutil.SetControllerReference(instance, configMap, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	found_cm := &corev1.ConfigMap{}

	/*	switch instance.Namespace {
		case
			"kube-system",
			"kube-public",
			"default":
			log.Printf("You are not allowed create in %s Namespace", instance.Namespace)
			panic("Namespace error")

		}*/

	err = r.Get(context.TODO(), types.NamespacedName{Name: "burrow-config", Namespace: configMap.Namespace}, found_cm)
	if err != nil && errors.IsNotFound(err) {
		log.Printf("Creating configMap %s/%s\n", configMap.Namespace, "burrow-config")
		err = r.Create(context.TODO(), configMap)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err

	}
	if !reflect.DeepEqual(configMap, found_cm) {
		found_cm = configMap
		log.Printf("Updating configMap %s/%s\n", configMap.Namespace, "config")
		err = r.Update(context.TODO(), found_cm)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	deploy := NewDeployment(*instance)

	if err := controllerutil.SetControllerReference(instance, deploy, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	//log.Printf("BurrowImage:%s", deploy.Namespace)

	found := &appsv1.Deployment{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Printf("Creating Deployment %s/%s\n", deploy.Namespace, deploy.Name)
		err = r.Create(context.TODO(), deploy)
		if err != nil {
			return reconcile.Result{}, err

		}
	} else if err != nil {
		return reconcile.Result{}, err

	}

	if !reflect.DeepEqual(deploy.Spec, found.Spec) {
		found.Spec = deploy.Spec
		log.Printf("Updating Deployment %s/%s\n", deploy.Namespace, deploy.Name)
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
		log.Printf("Creating Service %s/%s\n", service.Namespace, service.Name)
		err = r.Create(context.TODO(), service)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err

	}

	if !reflect.DeepEqual(service.Spec, found_service.Spec) {
		found_service.Spec = service.Spec
		log.Printf("Updating Service %s/%s\n", service.Namespace, service.Name)
		err = r.Update(context.TODO(), found)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil

}
