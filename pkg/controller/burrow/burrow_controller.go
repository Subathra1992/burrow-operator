package burrow

import (
	"context"
	//"io/ioutil"
	//"os"

	//"fmt"
	//"os"

	//"io/ioutil"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"log"
	//"os"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	monitorsv1beta1 "burrow-operator/pkg/apis/monitors/v1beta1"
	"github.com/pelletier/go-toml"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	ClientProfile struct {
		KafkaProfile struct {
			ClientID     string `json:"client-id"`
			KafkaVersion string `json:"kafka-version"`
		} `json:"kafka-profile"`
	} `json:"client-profile"`
	Cluster struct {
		MyCluster struct {
			ClassName     string `json:"class-name"`
			ClientProfile string `json:"client-profile"`
			OffsetRefresh int64  `json:"offset-refresh"`
			Servers       string `json:"servers"`
			TopicRefresh  int64  `json:"topic-refresh"`
		} `json:"my-cluster"`
	} `json:"cluster"`
	Consumer struct {
		ConsumerKafka struct {
			ClassName      string `json:"class-name"`
			ClientProfile  string `json:"client-profile"`
			Cluster        string `json:"cluster"`
			GroupBlacklist string `json:"group-blacklist"`
			GroupWhitelist string `json:"group-whitelist"`
			OffsetsTopic   string `json:"offsets-topic"`
			Servers        string `json:"servers"`
			StartLatest    bool   `json:"start-latest"`
		} `json:"consumer_kafka"`
	} `json:"consumer"`
	General struct {
		AccessControlAllowOrigin string `json:"access-control-allow-origin"`
	} `json:"general"`
	Httpserver struct {
		Default struct {
			Address string `json:"address"`
		} `json:"default"`
	} `json:"httpserver"`
	Logging struct {
		Level string `json:"level"`
	} `json:"logging"`
	Zookeeper struct {
		Servers string `json:"servers"`
	} `json:"zookeeper"`
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

	configMap := &corev1.ConfigMap{}
	configMap.APIVersion = "v1"
	configMap.Kind = "ConfigMap"
	configMap.Name = "config"
	configMap.Namespace = instance.Namespace


	tl, _ := toml.LoadFile("template/burrow.toml")

	data := make(map[string]string)

	for k, v := range tl.ToMap() {
		data[k] = v.(string)
	}
	configMap.Data=data

	//ts, ok := configMap.Data["template"]

	//configMap.BinaryData= map[string][]byte{}


	//bc := BuildConfigMap()
	//fmt.Printf("err: %v\n", bc)
	/*m, err := yaml.Marshal(bc)
	if err != nil {
		fmt.Printf("err: %v\n", err)

	}

	err = configMap.Unmarshal(m)*/
	if err != nil {
		return reconcile.Result{}, err

	}
	if err := controllerutil.SetControllerReference(instance, configMap, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	found_cm := &corev1.ConfigMap{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: "config", Namespace: configMap.Namespace}, found_cm)
	if err != nil && errors.IsNotFound(err) {
		log.Printf("Creating configMap %s/%s\n", configMap.Namespace, "config")
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
			Name:      instance.Name + "-deployment",
			Namespace: instance.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"deployment": instance.Name + "-deployment"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"deployment": instance.Name + "-deployment"}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "burrow-exporter",
							Image: instance.Spec.ExporterImage,
						},
						{
							Name:  "burrow-image",
							Image: instance.Spec.BurrowImage,
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:    "BurrowConfig",
							SubPath: "v1.BurrowDeploymentSpec{}." + ".toml",
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "BurrowConfig",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{

									LocalObjectReference: corev1.LocalObjectReference{Name: DeploymentName},
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
			Name: instance.Name,

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
				"name": instance.Name,
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

/*func LoadConfig(configFile string) (*Burrow, error) {
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, ge.New("Config file does not exist.")
	} else if err != nil {
		return nil, err
	}

	var conf Burrow
	if _, err := toml.DecodeFile(configFile, &conf); err != nil {
		return nil, err
	}

	return &conf, nil
}*/

func BuildConfigMap() (*corev1.ConfigMap) {

	/*tempcm := &Burrow{}

	reflect.TypeOf(tempcm)
	cmv := reflect.ValueOf(tempcm)
	bc := monitorsv1beta1.BurrowSpec{}
	crt := reflect.TypeOf(bc)
	crv := reflect.ValueOf(bc)

	//log.Printf("", cmt, cmv, crt, crv)

	for i := 0; i < crv.NumField(); i++ {

		cmv.FieldByName(crt.Field(i).Name).MapIndex(crv.Field(i))

	}

	log.Print(cmv)
	return tempcm*/

	/*jsonFile, err := os.Open("template/burrow.json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var result map[string]interface{}
	json.Unmarshal([]byte(byteValue), &result)
	fmt.Println("configmap", result)
	cm.Data=result


*/
	/*tempcm := &Burrow{}
	if file != nil {
		if _, err := toml.DecodeReader("template/burrow-configmap.yaml", &tempcm); err != nil {
			fileErr = fmt.Errorf("decode config: %v", err)
		}
	}
*/


/*config, _ := toml.LoadFile("template/burrow.toml")

	for _,field := range splittoList(monitorsv1beta1.Burrow{}.Spec.Zookeeper.Servers, ' ') {

		fmt.Println(field)
	}
config.Set("zookeeper.servers",)
*/
return nil



}

func splittoList(tosplit string, sep rune) []string {
	var fields []string

	last := 0
	for i,c := range tosplit {
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