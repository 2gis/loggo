package common

// EntryMapString is shorthand for easier representation of parsed data
type EntryMapString map[string]string

// Filter filters keys from map except ones specified
func (entryMap EntryMapString) Filter(keys ...string) EntryMapString {
	entrymap := make(EntryMapString)

	for _, key := range keys {
		if value, ok := entryMap[key]; ok {
			entrymap[key] = value
		}
	}
	return entrymap
}

// Extend is shorthand for extending EntryMapString with keys and values of specified one
func (entryMap EntryMapString) Extend(extends EntryMapString) {
	for key, value := range extends {
		entryMap[key] = value
	}
}
