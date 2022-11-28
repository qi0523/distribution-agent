package filesystem

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/qi0523/distribution-agent/constant"
	storagedriver "github.com/qi0523/distribution-agent/storage/driver"
	"github.com/qi0523/distribution-agent/storage/driver/base"
	"github.com/qi0523/distribution-agent/storage/driver/factory"
	"github.com/qi0523/distribution-agent/storage/help"
)

const (
	driverName        = "filesystem"
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
		rootDir    = constant.ContainerdRoot
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

func (d *driver) Reader(dgst string, mediaType string, offset int64) (io.ReadCloser, error) {
	file, err := os.Open(d.fullPath(subPath(dgst, mediaType)))
	if err != nil {
		file, err = os.Open(d.fullPath(help.BlobPath(dgst)))
		if err != nil {
			return nil, err
		}
	}
	seekPos, err := file.Seek(offset, io.SeekStart)
	if err != nil {
		file.Close()
		return nil, err
	} else if seekPos < offset {
		file.Close()
		return nil, storagedriver.InvalidOffsetError{Path: dgst, Offset: offset}
	}
	return file, nil
}

func (d *driver) Stat(subPath string) (int64, error) {
	fi, err := os.Stat(d.fullPath(subPath))
	if err != nil {
		if os.IsNotExist(err) {
			return 0, storagedriver.PathNotFoundError{Path: subPath}
		}
		return 0, err
	}
	return fi.Size(), nil
}

func subPath(dgst string, mediaType string) string {
	return filepath.Join(help.IngestPath(mediaType, dgst), "data")
}

func (d *driver) fullPath(subPath string) string {
	return path.Join(d.rootDir, subPath)
}
