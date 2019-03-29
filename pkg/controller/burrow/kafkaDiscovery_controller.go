package burrow

/*import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"time"
	"github.com/Sirupsen/logrus"

)

const maxRetries = 5

var serverStartTime time.Time
var action string
var deletedNamespace string


type Controller struct {
	logger       *log.Entry
	clientset    kubernetes.Interface
	queue        workqueue.RateLimitingInterface
	informer     cache.SharedIndexInformer
	config       *config.Config



}



func GetClient() kubernetes.Interface {
	config, err := rest.InClusterConfig()
	if err != nil {
		logrus.Fatalf("Can not get kubernetes config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logrus.Fatalf("Can not create kubernetes client: %v", err)
	}

	return clientset
}

// TODO(user): add KafkaDiscovery




func execute {


	var k8Client kubernetes.Interface
	_, err := rest.InClusterConfig()

	if err != nil {
		k8Client = utils.GetClientOutOfCluster()
	} else {
		k8Client = utils.GetClient()
	}

	cfg := config.NewConfig()
	controller.Start(k8Client, cfg)


}
*/
