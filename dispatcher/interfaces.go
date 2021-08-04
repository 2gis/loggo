package dispatcher

import (
	"github.com/2gis/loggo/components/containers"
)

// ContainersProvider is a containers provider interface for dispatcher
type ContainersProvider interface {
	Containers() (containers.Containers, error)
}
