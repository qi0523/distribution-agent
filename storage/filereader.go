package storage

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	storagedriver "github.com/qi0523/distribution-agent/storage/driver"
)

const (
	fileReaderBufferSize = 4 << 20
)

type fileReader struct {
	driver storagedriver.StorageDriver
	path   string
	size   int64
	rc     io.ReadCloser
	brd    *bufio.Reader
	offset int64
	err    error
}

func newFileReader(driver storagedriver.StorageDriver, path string, size int64) (*fileReader, error) {
	return &fileReader{
		driver: driver,
		path:   path,
		size:   size,
	}, nil
}

func (fr *fileReader) Read(p []byte) (n int, err error) {
	if fr.err != nil {
		return 0, fr.err
	}
	rd, err := fr.reader()
	if err != nil {
		return 0, err
	}
	n, err = rd.Read(p)
	fr.offset += int64(n)
	if err == nil && fr.offset >= fr.size {
		err = io.EOF
	}
	return n, err
}

func (fr *fileReader) Seek(offset int64, whence int) (int64, error) {
	if fr.err != nil {
		return 0, fr.err
	}
	var err error
	newOffset := fr.offset
	switch whence {
	case io.SeekCurrent:
		newOffset += offset
	case io.SeekEnd:
		newOffset = fr.size + offset
	case io.SeekStart:
		newOffset = offset
	}
	if newOffset < 0 {
		err = fmt.Errorf("cannot seek to negative position")
	} else {
		if fr.offset != newOffset {
			fr.reset()
		}
		fr.offset = newOffset
	}
	return fr.offset, err
}

func (fr *fileReader) Close() error {
	return fr.closeWithErr(fmt.Errorf("fileReader: closed"))
}

func (fr *fileReader) reader() (io.Reader, error) {
	if fr.err != nil {
		return nil, fr.err
	}
	if fr.rc != nil {
		return fr.brd, nil
	}

	rc, err := fr.driver.Reader(fr.path, fr.offset)
	if err != nil {
		switch err := err.(type) {
		case storagedriver.PathNotFoundError:
			return ioutil.NopCloser(bytes.NewReader([]byte{})), nil
		default:
			return nil, err
		}
	}
	fr.rc = rc
	if fr.brd == nil {
		fr.brd = bufio.NewReaderSize(fr.rc, fileReaderBufferSize)
	} else {
		fr.brd.Reset(fr.rc)
	}
	return fr.brd, nil
}

func (fr *fileReader) reset() {
	if fr.err != nil {
		return
	}
	if fr.rc != nil {
		fr.rc.Close()
		fr.rc = nil
	}
}

func (fr *fileReader) closeWithErr(err error) error {
	if fr.err != nil {
		return fr.err
	}
	fr.err = err
	if fr.rc != nil {
		fr.rc.Close()
	}
	fr.rc = nil
	fr.brd = nil
	return fr.err
}
