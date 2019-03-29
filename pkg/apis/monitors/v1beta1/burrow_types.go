package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BurrowSpec defines the desired state of Burrow

type BurrowConfigSpec struct {

	// +optional
	Logging struct {
		// +optional
		Level string `json:"level,omitempty"`
	} `json:"logging,omitempty"`
	// +optional
	Zookeeper struct {
		Servers string `json:"zkservers"`
	} `json:"zookeeper,omitempty"`
	// +optional
	ClientProfileKafkaProfile struct {
		// +optional
		KafkaVersion string `json:"kafka-version,omitempty"`
		// +optional
		ClientID string `json:"client-id,omitempty"`
	} `json:"client-profile,omitempty"`
	// +optional
	ClusterMyCluster struct {
		// +optional
		ClassName string `json:"class-name,omitempty"`
		// +optional
		ClientProfile string `json:"client-profile,omitempty"`

		Servers string `json:"bkservers"`
		// +optional
		TopicRefresh int `json:"topic-refresh,omitempty"`
		// +optional
		OffsetRefresh int `json:"offset-refresh,omitempty"`
	} `json:"cluster,omitempty"`
	// +optional
	ConsumerConsumerKafka struct {
		// +optional
		ClassName string `json:"class-name,omitempty"`
		// +optional
		Cluster string `json:"cluster,omitempty"`

		Servers string `json:"bkservers"`
		// +optional
		ClientProfile string `json:"client-profile,omitempty"`
		// +optional
		StartLatest bool `json:"start-latest,omitempty"`
		// +optional
		OffsetsTopic string `json:"offsets-topic,omitempty"`
		// +optional
		GroupWhitelist string `json:"group-whitelist,omitempty"`
		// +optional
		GroupBlacklist string `json:"group-blacklist,omitempty"`
	} `json:"consumer,omitempty"`
	// +optional
	HttpserverDefault struct {
		// +optional
		Address string `json:"address,omitempty"`
	} `json:"httpserver,omitempty"`
	// +optional
	ApiPort int32 `json:"apiport,omitempty"`
	// +optional
	MetricsPort int32 `json:"metricsport,omitempty"`
	// +optional
	BurrowImage string `json:"burrowImag,omitempty"`
	// +optional
	ExporterImage string `json:"exporterImage,omitempty"`
	// +optional
	ApiVersion int32 `json:"api_version,omitempty"`
	// +optional
	Interval int32 `json:"interval,omitempty"`
}

/*type Zookeeper struct {
	Servers []string `json:"servers,omitempty"`

}

type Logging struct {
	Level string `json:"level,omitempty"`
}*/

// BurrowStatus defines the observed state of Burrow
type BurrowStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Burrowcreated corev1.ConditionStatus `json:"status,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Burrow is the Schema for the burrows API
// +k8s:openapi-gen=true
type Burrow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec BurrowConfigSpec `json:"spec,omitempty"`

	Status BurrowStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BurrowList contains a list of Burrow
type BurrowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Burrow `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Burrow{}, &BurrowList{})
}
