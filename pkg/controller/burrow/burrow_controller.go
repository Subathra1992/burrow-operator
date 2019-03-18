package burrow

import (
	"bytes"
	"context"
	"io/ioutil"
	"path/filepath"
	"strings"

	//"io/ioutil"
	//"os"

	//"fmt"
	//"os"

	monitorsv1beta1 "burrow-operator/pkg/apis/monitors/v1beta1"
	"encoding/json"
	"github.com/BurntSushi/toml"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
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

type burrowconfig struct {
	Keys map[string]map[string]string
}

type Burrow struct {
	General struct {
		AccessControlAllowOrigin string `json:"access-control-allow-origin"`
	} `json:"general"`
	Logging struct {
		Level string `json:"level"`
	} `json:"logging"`
	Zookeeper struct {
		Servers []string `json:"servers"`
	} `json:"zookeeper"`
	ClientProfileKafkaProfile struct {
		KafkaVersion string `json:"kafka-version"`
		ClientID     string `json:"client-id"`
	} `json:"client-profile:kafka-profile"`
	ClusterMyCluster struct {
		ClassName     string   `json:"class-name"`
		ClientProfile string   `json:"client-profile"`
		Servers       []string `json:"servers"`
		TopicRefresh  int      `json:"topic-refresh"`
		OffsetRefresh int      `json:"offset-refresh"`
	} `json:"cluster:my-cluster"`
	ConsumerConsumerKafka struct {
		ClassName      string   `json:"class-name"`
		Cluster        string   `json:"cluster"`
		Servers        []string `json:"servers"`
		ClientProfile  string   `json:"client-profile"`
		StartLatest    bool     `json:"start-latest"`
		OffsetsTopic   string   `json:"offsets-topic"`
		GroupWhitelist string   `json:"group-whitelist"`
		GroupBlacklist string   `json:"group-blacklist"`
	} `json:"consumer:consumer_kafka"`
	HttpserverDefault struct {
		Address string `json:"address"`
	} `json:"httpserver:default"`
}

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
// +kubebuilder:rbac:groups=monitors.cisco.com,resources=burrows,verbs=get;list;watch;create;update;patch;delete
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
	log.Printf("Checking status of resource %s", instance.Name)
	log.Printf("zkserver:%s", instance.Spec.Zookeeper.Servers)

	log.Printf("BurrowImage:%s", instance.Spec.BurrowImage)
	log.Printf("BurrowImage:%s", instance.Spec.ExporterImage)

	var burrowconfigmap Burrow
	filename, _ := filepath.Abs("template/burrow.json")
	yamlFile, err := ioutil.ReadFile(filename)

	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(yamlFile, &burrowconfigmap)
	if err != nil {
		panic(err)
	}

	configMap := &corev1.ConfigMap{}
	configMap.APIVersion = "v1"
	configMap.Kind = "ConfigMap"
	configMap.Name = "burrow-config"
	configMap.Namespace = instance.Namespace

	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(burrowconfigmap); err != nil {
		panic(err)
	}

	var s string
	buf.WriteString(s)
	strings.ToLower(s)
	//log.Print("%s", s)
	configMap.Data = map[string]string{
		"burrow.toml": strings.ToLower(buf.String()),
	}



	if err != nil {
		return reconcile.Result{}, err

	}
	if err := controllerutil.SetControllerReference(instance, configMap, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	found_cm := &corev1.ConfigMap{}

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

	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "burrow",
			Namespace: instance.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"deployment": "burrow"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"deployment": "burrow"}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "burrow-image",
							Image: instance.Spec.BurrowImage,
							VolumeMounts: []corev1.VolumeMount{{
								Name:      "config",
								MountPath: "/etc/burrow/config",
							},
							},
						},
						{
							Name:  "burrow-exporter",
							Image: instance.Spec.ExporterImage,
							Env: []corev1.EnvVar{{
								Name:  "BURROW_ADDR",
								Value: "http://localhost:8000",
							},
								{
									Name:  "METRICS_ADDR",
									Value: "0.0.0.0:8080",
								},
								{
									Name:  "INTERVAL",
									Value: "15",
								},

								{
									Name:  "API_VERSION",
									Value: "3",
								},
							},
						},
					},

					Volumes: []corev1.Volume{
						{
							Name: "config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{

									LocalObjectReference: corev1.LocalObjectReference{Name: "burrow-config"},
								},
							},
						},
					},
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(instance, deploy, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	log.Printf("BurrowImage:%s", deploy.Namespace)

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

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "burrow",

			OwnerReferences: nil,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "burrow",
					Port:       instance.Spec.ApiPort,
					TargetPort: intstr.FromInt(8080),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"name": "burrow",
			},
		},
	}
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

/*func buildInmemoryConfigMap(input Burrow ,spec monitorsv1beta1.BurrowSpec,configMap corev1.ConfigMap) {


	op:= reflect.ValueOf(spec)
	op_values := make([]interface{}, op.NumField()):q


	in:=reflect.ValueOf(input)




	for i := 0; i < op.NumField(); i++ {


		//op_values[i] = op.Field(i).Interface()

		s := in.Field(i).FieldByName().String()
		i2 := op.FieldByName(s).String()

	}


	return



}*/
func splittoList(tosplit string, sep rune) []string {
	var fields []string
	last := 0
	for i, c := range tosplit {
		if c == sep {
			// Found the separator, append a slice
			fields = append(fields, string(tosplit[last:i]))
			last = i + 1
		}
	}

	// Don't forget the last field
	fields = append(fields, string(tosplit[last:]))

	return fields
}
