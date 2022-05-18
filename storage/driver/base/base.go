package base

import (
	"io"

	"github.com/qi0523/distribution-agent/storage/driver"
)

type Base struct {
	driver.StorageDriver
}

func (base *Base) setDriverName(e error) error {
	switch actual := e.(type) {
	case nil:
		return nil
	case driver.ErrUnsupportedMethod:
		actual.DriverName = base.StorageDriver.Name()
		return actual
	case driver.PathNotFoundError:
		actual.DriverName = base.StorageDriver.Name()
		return actual
	case driver.InvalidOffsetError:
		actual.DriverName = base.StorageDriver.Name()
		return actual
	default:
		return driver.Error{
			DriverName: base.StorageDriver.Name(),
			Enclosed:   e,
		}
	}
}

func (base *Base) Reader(path string, offset int64) (io.ReadCloser, error) {
	if offset < 0 {
		return nil, driver.InvalidOffsetError{Path: path, Offset: offset, DriverName: base.StorageDriver.Name()}
	}
	rc, e := base.StorageDriver.Reader(path, offset)
	return rc, base.setDriverName(e)
}

func (base *Base) Stat(path string) (int64, error) {
	size, e := base.StorageDriver.Stat(path)
	return size, base.setDriverName(e)
}
