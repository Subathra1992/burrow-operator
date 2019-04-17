package burrow

import (
	"bytes"
	"encoding/json"
	"github.com/BurntSushi/toml"
	monitorsv1beta1 "github.com/subravi92/burrow-operator/pkg/apis/monitors/v1beta1"
	"io/ioutil"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"log"
	"path/filepath"
	"strings"
	//	"text/template"
)

const (
	serviceName       = "burrow"
	deploymentName    = "burrow"
	configMapName     = "burrow-config"
	configMapMetaName = "config"
)

type burrowconfig struct {
	Keys map[string]map[string]string
}

//move to the utils.go

type BurrowDeployment struct {
	interval    int
	api_version int
}
type Config struct {
	OperatorNamespace      string
	OperatorDeploymentName string
	OperatorEnvConfigName  string
	NamespaceExclusion     []string
	EntitledNamespace      []string
	LabelFilter            string
	burrowconfiglist       []string
}

func populateConfig() *Config {
	c := new(Config)
	c.NamespaceExclusion = []string{"kube-system", "kube-public", "default"}
	return c

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
	ClientProfile struct {
		KafkaProfile struct {
			KafkaVersion string `json:"kafka-version"`
			ClientID     string `json:"client-id"`
		} `json:"kafka-profile"`
	} `json:"client-profile"`
	Cluster struct {
		MyCluster struct {
			ClassName     string   `json:"class-name"`
			ClientProfile string   `json:"client-profile"`
			Servers       []string `json:"servers"`
			TopicRefresh  int      `json:"topic-refresh"`
			OffsetRefresh int      `json:"offset-refresh"`
		} `json:"my-cluster"`
	} `json:"cluster"`
	Consumer struct {
		ConsumerKafka struct {
			ClassName      string   `json:"class-name"`
			Cluster        string   `json:"cluster"`
			Servers        []string `json:"servers"`
			ClientProfile  string   `json:"client-profile"`
			StartLatest    bool     `json:"start-latest"`
			OffsetsTopic   string   `json:"offsets-topic"`
			GroupWhitelist string   `json:"group-whitelist"`
			GroupBlacklist string   `json:"group-blacklist"`
		} `json:"consumer_kafka"`
	} `json:"consumer"`
	Httpserver struct {
		Default struct {
			Address string `json:"address"`
		} `json:"default"`
	} `json:"httpserver"`
}

type DeploymentInput struct {
	Name          string
	burrowImage   string
	exporterImage string

	//ImagePullPolicy v1.PullPolicy
	Labels       map[string]string
	Selector     *meta_v1.LabelSelector
	Replicas     *int32
	Namespace    string
	Ports        []v1.ContainerPort
	VolumeMounts []v1.VolumeMount
	Resources    v1.ResourceRequirements
	Volumes      []v1.Volume
	Envs         []v1.EnvVar
	ApiPort      *int32
	Interval     string
	ApiVersion   string
}

func NewDeployment(instance monitorsv1beta1.Burrow) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "burrow",
			Namespace: instance.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &meta_v1.LabelSelector{
				MatchLabels: map[string]string{"deployment": deploymentName},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: meta_v1.ObjectMeta{Labels: map[string]string{"deployment": deploymentName}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "burrow-image",
							Image: instance.Spec.BurrowImage,
							VolumeMounts: []corev1.VolumeMount{{
								Name:      configMapMetaName,
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
							Name: configMapMetaName,
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{

									LocalObjectReference: corev1.LocalObjectReference{Name: configMapName},
								},
							},
						},
					},
				},
			},
		},
	}
}

func NewConfigMap(instance monitorsv1beta1.Burrow) *corev1.ConfigMap {

	var burrowconfigmap Burrow
	filename, _ := filepath.Abs("config/template/burrow.json")
	jsonFile, err := ioutil.ReadFile(filename)

	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(jsonFile, &burrowconfigmap)
	if err != nil {
		panic(err)
	}

	//b, err := ioutil.ReadFile("template/burrow.json")

	log.Printf(":%s", burrowconfigmap.Zookeeper)

	//data,err :=  YamlToStruct("tmpldata.yaml")

	configMap := &corev1.ConfigMap{}
	configMap.APIVersion = "v1"
	configMap.Kind = "ConfigMap"
	configMap.Name = configMapName
	configMap.Namespace = instance.Namespace
	output := buildInmemoryConfigMap(burrowconfigmap, instance.Spec, *configMap)

	//var cm Burrow
	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(output); err != nil {
		panic(err)
	}

	configMap.Data = map[string]string{
		"burrow.toml": strings.ToLower(buf.String()),
	}

	/*t, err := template.New("cm").ParseFiles("template/burrow-tem.yaml")
	if err != nil {
		panic(err)
	}

	var tpl bytes.Buffer

	err = t.Execute(&tpl, cm)
	if err != nil {
		panic(err)
	}

	result := tpl.String()*/

	return configMap
}

func NewService(instance monitorsv1beta1.Burrow) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:            serviceName,
			Namespace:       instance.Namespace,
			OwnerReferences: nil,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       serviceName,
					Port:       instance.Spec.ApiPort,
					TargetPort: intstr.FromInt(int(instance.Spec.ApiPort)),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"name": serviceName,
			},
		},
	}
}

func buildInmemoryConfigMap(input Burrow, spec monitorsv1beta1.BurrowConfigSpec, configMap corev1.ConfigMap) Burrow {

	if spec.ClusterMyCluster.OffsetRefresh != 0 {
		input.Cluster.MyCluster.OffsetRefresh = spec.ClusterMyCluster.OffsetRefresh
	}

	if spec.ClusterMyCluster.ClientProfile != "" {
		input.Cluster.MyCluster.ClientProfile = spec.ClusterMyCluster.ClientProfile
	}

	if spec.ClusterMyCluster.TopicRefresh != 0 {
		input.Cluster.MyCluster.TopicRefresh = spec.ClusterMyCluster.TopicRefresh
	}
	if spec.ClusterMyCluster.Servers != "" {
		serverList := strings.Split(spec.ClusterMyCluster.Servers, ",")
		input.Cluster.MyCluster.Servers = serverList
	}

	if spec.ConsumerConsumerKafka.Servers != "" {
		serverList := strings.Split(spec.ConsumerConsumerKafka.Servers, ",")
		input.Consumer.ConsumerKafka.Servers = serverList
	}

	if spec.ConsumerConsumerKafka.Cluster != "" {
		input.Consumer.ConsumerKafka.Cluster = spec.ConsumerConsumerKafka.Cluster
	}

	if spec.ConsumerConsumerKafka.OffsetsTopic != "" {
		input.Consumer.ConsumerKafka.Cluster = spec.ConsumerConsumerKafka.Cluster
	}
	if spec.ConsumerConsumerKafka.GroupBlacklist != "" {
		input.Consumer.ConsumerKafka.GroupBlacklist = spec.ConsumerConsumerKafka.GroupBlacklist
	}
	if spec.ConsumerConsumerKafka.GroupWhitelist != "" {
		input.Consumer.ConsumerKafka.GroupWhitelist = spec.ConsumerConsumerKafka.GroupWhitelist
	}

	if spec.ConsumerConsumerKafka.StartLatest != false {
		input.Consumer.ConsumerKafka.StartLatest = spec.ConsumerConsumerKafka.StartLatest
	}
	if spec.Zookeeper.Servers != "" {
		serverList := strings.Split(spec.Zookeeper.Servers, ",")
		input.Zookeeper.Servers = serverList

	}
	if spec.ClientProfileKafkaProfile.KafkaVersion != "" {
		input.ClientProfile.KafkaProfile.KafkaVersion = spec.ClientProfileKafkaProfile.KafkaVersion
	}
	if spec.ClientProfileKafkaProfile.ClientID != "" {
		input.ClientProfile.KafkaProfile.ClientID = spec.ClientProfileKafkaProfile.ClientID
	}

	if spec.Logging.Level != "" {
		input.Logging.Level = spec.Logging.Level
	}

	input.Consumer.ConsumerKafka.ClassName = "kafka"
	log.Printf("%+v", input)
	return input

}

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

func isValidNamespace(namespace string) bool {

	switch namespace {
	case "kube-system", "kube-public", "default":

		log.Printf("You are not allowed create in %s Namespace", namespace)
		return false
	}
	return true
}
