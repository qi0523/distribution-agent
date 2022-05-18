package driver

import (
	"fmt"
	"io"
)

type StorageDriver interface {
	Name() string

	Reader(path string, offset int64) (io.ReadCloser, error)

	Stat(path string) (int64, error)
}

type ErrUnsupportedMethod struct {
	DriverName string
}

func (err ErrUnsupportedMethod) Error() string {
	return fmt.Sprintf("%s: unsupported method", err.DriverName)
}

type PathNotFoundError struct {
	Path       string
	DriverName string
}

func (err PathNotFoundError) Error() string {
	return fmt.Sprintf("%s: Path not found: %s", err.DriverName, err.Path)
}

type InvalidOffsetError struct {
	Path       string
	Offset     int64
	DriverName string
}

func (err InvalidOffsetError) Error() string {
	return fmt.Sprintf("%s: invalid offset: %d for path: %s", err.DriverName, err.Offset, err.Path)
}

type Error struct {
	DriverName string
	Enclosed   error
}

func (err Error) Error() string {
	return fmt.Sprintf("%s: %s", err.DriverName, err.Enclosed)
}
