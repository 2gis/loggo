package common

/* Pack of constant logstash purposed keys */
const (
	LabelDataCenter        = "dc"
	LabelPurpose           = "purpose"
	LabelContainerID       = "docker.container_id"
	LabelLogstashPrefix    = "logstash_prefix"
	LabelLogstashNamespace = "namespace"
	LabelLogType           = "type"
	LabelTime              = "time"
)

/* Filename is arbitrary cursorStorage key picked as the standard journal file ending */
const (
	FilenameJournald  = "system.journal"
	NamespaceJournald = "journald"
)

// common
const (
	// KubernetesPodName name of field
	KubernetesPodName = "kubernetes.pod_name"

	// KubernetesNamespaceName name of field
	KubernetesNamespaceName = "kubernetes.namespace_name"

	// KubernetesContainerName name of field
	KubernetesContainerName = "kubernetes.container_name"

	// KubernetesNodeHostname name of field
	KubernetesNodeHostname = "kubernetes.node_hostname"
)

// Containers Provider related
const (
	// LabelKubernetesPodNamespace store label name for pod namespace
	LabelKubernetesPodNamespace = "io.kubernetes.pod.namespace"

	// LabelKubernetesPodName store label name of pod
	LabelKubernetesPodName = "io.kubernetes.pod.name"

	// LabelKubernetesContainerName store label name of container
	LabelKubernetesContainerName = "io.kubernetes.container.name"
)
