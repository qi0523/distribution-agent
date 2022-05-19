package factory

import (
	"fmt"

	storagedriver "github.com/qi0523/distribution-agent/storage/driver"
)

var driverFactories = make(map[string]StorageDriverFactory)

type StorageDriverFactory interface {
	Create(parameters map[string]interface{}) (storagedriver.StorageDriver, error)
}

func Register(name string, factory StorageDriverFactory) {
	if factory == nil {
		panic("Must not provide nil StorageDriverFactory")
	}
	_, registered := driverFactories[name]
	if registered {
		panic(fmt.Sprintf("StorageDriverFactory named %s already registered", name))
	}
	driverFactories[name] = factory
}

func Create(name string, parameters map[string]interface{}) (storagedriver.StorageDriver, error) {
	driverFactory, ok := driverFactories[name]
	if !ok {
		return nil, InvalidStorageDriverError{name}
	}
	return driverFactory.Create(parameters)
}

type InvalidStorageDriverError struct {
	Name string
}

func (err InvalidStorageDriverError) Error() string {
	return fmt.Sprintf("StorageDriver not registered: %s", err.Name)
}
