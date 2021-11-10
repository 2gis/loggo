package common

// EntryMap is a shorthand for easier representation of parsed data
type EntryMap map[string]interface{}

// Extend is a shorthand for extending EntryMapString with keys and values of specified one
func (entryMap EntryMap) Extend(extends EntryMap) {
	for key, value := range extends {
		entryMap[key] = value
	}
}

func (entryMap EntryMap) Filter(keys ...string) EntryMap {
	entrymap := make(EntryMap)

	for _, key := range keys {
		if value, ok := entryMap[key]; ok {
			entrymap[key] = value
		}
	}

	return entrymap
}


// NamespaceName attempts to return field from EntryMap
func (entryMap EntryMap) NamespaceName() string {
	namespace, ok := entryMap[KubernetesNamespaceName].(string)

	if !ok {
		return ""
	}

	return namespace
}

// PodName attempts to return field from EntryMap
func (entryMap EntryMap) PodName() string {
	podName, ok := entryMap[KubernetesPodName].(string)

	if !ok {
		return ""
	}

	return podName
}

// ContainerName attempts to return field from EntryMap
func (entryMap EntryMap) ContainerName() string {
	containerName, ok := entryMap[KubernetesContainerName].(string)

	if !ok {
		return ""
	}

	return containerName
}
