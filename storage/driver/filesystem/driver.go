package filesystem

import (
	"fmt"
	"io"
	"os"
	"path"

	storagedriver "github.com/qi0523/distribution-agent/storage/driver"
	"github.com/qi0523/distribution-agent/storage/driver/base"
	"github.com/qi0523/distribution-agent/storage/driver/factory"
)

const (
	driverName        = "filesystem"
	defaultRootDir    = "/var/lib/containerd/io.containerd.content.v1.content/blobs/sha256"
	defaultMaxThreads = uint64(30)
	minThreads        = uint64(10)
)

type DriverParameters struct {
	RootDir    string
	MaxThreads uint64
}

func init() {
	factory.Register(driverName, &filesystemDriverFactory{})
}

type filesystemDriverFactory struct{}

func (factory *filesystemDriverFactory) Create(parameters map[string]interface{}) (storagedriver.StorageDriver, error) {
	return FromParameters(parameters)
}

type driver struct {
	rootDir string
}

type baseEmbed struct {
	base.Base
}

type Driver struct {
	baseEmbed
}

func FromParameters(parameters map[string]interface{}) (*Driver, error) {
	params, err := FromParametersImpl(parameters)
	if err != nil || params == nil {
		return nil, err
	}
	return New(*params), nil
}

func FromParametersImpl(parameters map[string]interface{}) (*DriverParameters, error) {
	var (
		err        error
		maxThreads = defaultMaxThreads
		rootDir    = defaultRootDir
	)
	if parameters != nil {
		if rootDirectory, ok := parameters["rootdirectory"]; ok {
			rootDir = fmt.Sprint(rootDirectory)
		}
		maxThreads, err = base.GetLimitFromParamter(parameters["maxthreads"], minThreads, defaultMaxThreads)
		if err != nil {
			return nil, fmt.Errorf("maxthreads config error: %s", err.Error())
		}
	}
	params := &DriverParameters{
		RootDir:    rootDir,
		MaxThreads: maxThreads,
	}
	return params, nil
}

func New(params DriverParameters) *Driver {
	fsDriver := &driver{rootDir: params.RootDir}

	return &Driver{
		baseEmbed: baseEmbed{
			Base: base.Base{
				StorageDriver: base.NewRegulator(fsDriver, params.MaxThreads),
			},
		},
	}
}

func (d *driver) Name() string {
	return driverName
}

func (d *driver) Reader(path string, offset int64) (io.ReadCloser, error) {
	file, err := os.Open(d.fullPath(path))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, storagedriver.PathNotFoundError{Path: path}
		}
		return nil, err
	}
	seekPos, err := file.Seek(offset, io.SeekStart)
	if err != nil {
		file.Close()
		return nil, err
	} else if seekPos < offset {
		file.Close()
		return nil, storagedriver.InvalidOffsetError{Path: path, Offset: offset}
	}
	return file, nil
}

func (d *driver) Stat(subPath string) (int64, error) {
	fullPath := d.fullPath(subPath)

	fi, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, storagedriver.PathNotFoundError{Path: subPath}
		}
		return 0, err
	}
	return fi.Size(), nil
}

func (d *driver) fullPath(subPath string) string {
	return path.Join(d.rootDir, subPath)
}
