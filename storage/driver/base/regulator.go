package base

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"sync"

	storagedriver "github.com/qi0523/distribution-agent/storage/driver"
)

type regulator struct {
	storagedriver.StorageDriver
	*sync.Cond

	available uint64
}

func GetLimitFromParamter(param interface{}, min, def uint64) (uint64, error) {
	limit := def
	switch v := param.(type) {
	case string:
		var err error
		if limit, err = strconv.ParseUint(v, 0, 64); err != nil {
			return limit, fmt.Errorf("parameter must be an integer, '%v' invalid", param)
		}
	case uint64:
		limit = v
	case int, int32, int64:
		val := reflect.ValueOf(v).Convert(reflect.TypeOf(param)).Int()
		if val > 0 {
			limit = uint64(val)
		} else {
			limit = min
		}
	case uint, uint32:
		limit = reflect.ValueOf(v).Convert(reflect.TypeOf(param)).Uint()
	case nil:
	default:
		return 0, fmt.Errorf("invalid value '%#v'", param)
	}
	if limit < min {
		return min, nil
	}
	return limit, nil
}

func NewRegulator(driver storagedriver.StorageDriver, limit uint64) storagedriver.StorageDriver {
	return &regulator{
		StorageDriver: driver,
		Cond:          sync.NewCond(&sync.Mutex{}),
		available:     limit,
	}
}

func (r *regulator) enter() {
	r.L.Lock()
	for r.available == 0 {
		r.Wait()
	}
	r.available--
	r.L.Unlock()
}

func (r *regulator) exit() {
	r.L.Lock()
	r.Signal()
	r.available++
	r.L.Unlock()
}

func (r *regulator) Name() string {
	r.enter()
	defer r.exit()
	return r.StorageDriver.Name()
}

func (r *regulator) Reader(path string, offset int64) (io.ReadCloser, error) {
	r.enter()
	defer r.exit()
	return r.StorageDriver.Reader(path, offset)
}

func (r *regulator) Stat(path string) (int64, error) {
	r.enter()
	defer r.exit()
	return r.StorageDriver.Stat(path)
}
