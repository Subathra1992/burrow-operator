package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BurrowSpec defines the desired state of Burrow
type BurrowSpec struct {
	// +optional
	ClientProfile struct {
		// +optional
		KafkaProfile struct {
			// +optional
			ClientID string `json:"client-id,omitempty"`
			// +optional
			KafkaVersion string `json:"kafka-version,omitempty"`
		} `json:"kafka-profile,omitempty"`
	} `json:"client-profile,omitempty"`
	// +optional
	Cluster struct {
		// +optional
		MyCluster struct {
			// +optional
			ClassName string `json:"class-name,omitempty"`
			// +optional
			ClientProfile string `json:"client-profile,omitempty"`
			// +optional
			OffsetRefresh int64 `json:"offset-refresh,omitempty"`
			// +optional
			Servers string `json:"bkservers,omitempty"`
			// +optional
			TopicRefresh int64 `json:"topic-refresh,omitempty"`
		} `json:"my-cluster,omitempty"`
	} `json:"cluster,omitempty"`
	// +optional
	Consumer struct {
		// +optional
		ConsumerKafka struct {
			// +optional
			ClassName string `json:"class-name,omitempty"`
			// +optional
			ClientProfile string `json:"client-profile,omitempty"`
			// +optional
			Cluster string `json:"cluster,omitempty"`
			// +optional
			GroupBlacklist string `json:"group-blacklist,omitempty"`
			// +optional
			GroupWhitelist string `json:"group-whitelist,omitempty"`
			// +optional
			OffsetsTopic string `json:"offsets-topic,omitempty"`
			// +optional
			Servers string `json:"bkservers,omitempty"`
			// +optional
			StartLatest bool `json:"start-latest,omitempty"`
		} `json:"consumer_kafka,omitempty"`
	} `json:"consumer,omitempty"`
	// +optional
	General struct {
		// +optional
		AccessControlAllowOrigin string `json:"access-control-allow-origi,omitemptyn"`
	} `json:"general,omitempty"`
	// +optional
	Httpserver struct {
		// +optional
		Default struct {
			// +optional
			Address string `json:"address,omitempty"`
		} `json:"default,omitempty"`
	} `json:"httpserver,omitempty"`
	// +optional
	Logging struct {
		// +optional
		Level string `json:"level,omitempty"`
	} `json:"logging,omitempty"`

	// +optional

	Zookeeper struct {
		// +optional
		Servers string `json:"zkservers,omitempty"`
	} `json:"zookeeper,omitempty"`

	// +optional
	ApiPort int32 `json:"apiport,omitempty"`
	// +optional
	MetricsPort int32 `json:"metricsport,omitempty"`
	// +optional
	BurrowImage string `json:"burrowImag,omitempty"`
	// +optional
	ExporterImage string `json:"exporterImage,omitempty"`
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

	Spec   BurrowSpec   `json:"spec,omitempty"`
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
