package burrow

import (
	burrowv1beta1 "burrow-operator/pkg/apis/monitors/v1beta1"
	//"burrow-operator/pkg/controller"
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"log"
)
func (c *Controller)) addResources(obj interface{}) {
	burrowspec := obj.(*burrowv1beta1.Burrow)

	DName := fmt.Sprintf("%s:%s", "Burrow", burrowspec.GetName())

	cm, err := BurrowConfigMap(DName)

	err = controllers.CreateConfigMap(c.client.GetK8sClient().CoreV1().ConfigMaps(burrowspec.GetNamespace()), cm)

	if err != nil {
		log.Print(err)
		return
	}
	deploy := burrowDeployment(DName, &v1.BurrowDeploymentSpec{})
	err = controllers.CreateDeployment(c.client.GetK8sClient().AppsV1().Deployments(burrowspec.GetNamespace()), deploy)
	if err != nil {
		log.Print(err)
	}
	svc := BurrowService(DName)
	err = controllers.CreateService(c.client.GetK8sClient().CoreV1().Services(burrowspec.GetNamespace()), svc)

	if err != nil {
		log.Print(err)
		return
	}

}

func burrowDeployment(DeploymentName string, burrowspec *v1.BurrowDeploymentSpec) *appsv1.Deployment {

	labels := map[string]string{}

	labels["name"] = DeploymentName
	labels["resource"] = "Burrow"

	i := controllers.DeploymentInput{
		Name:  labels["name"],
		Image: v1.BurrowDeploymentSpec{}.BurrowImage,

		Labels:    labels,
		Namespace: v1.BurrowDeploymentSpec{}.KafkaNamespace,
		Ports: []corev1.ContainerPort{
			corev1.ContainerPort{
				Name:          "http",
				ContainerPort: 8086,
				Protocol:      corev1.ProtocolTCP,
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
	}

	return controllers.NewDeployment(i)
}

func BurrowService(name string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,

			OwnerReferences: nil,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "burrow",
					Port:       8086,
					TargetPort: intstr.FromInt(8086),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"name": name,
			},
		},
	}
}

func BurrowConfigMap(name string) (*corev1.ConfigMap, error) {
	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(burrowconfig{}); err != nil {
		return nil, err
	}

	data := map[string]string{
		v1.BurrowDeploymentSpec{}.KafkaNamespace: buf.String(),
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Data: data,
	}, nil
}

func (c *Controller) deleteResources(obj interface{}) {
	//burrowspec := obj.(*v1.BurrowDeployment)

}
