package stat

import (
	"fmt"
	"io"
	"os"
	"sync"
)

var (
	DefaultBlockLen = 36
)

type Binlog struct {
	f           *os.File
	blockLength int

	wrMutex sync.Mutex
}

func NewWriteBinLog(fpath string) (*Binlog, error) {
	bl := new(Binlog)

	var e error
	bl.f, e = os.OpenFile(fpath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	if e != nil {
		return nil, e
	}

	bl.SetBlockLen(DefaultBlockLen)

	return bl, nil
}

func (bl *Binlog) SetBlockLen(l int) {
	bl.blockLength = l
}

func (bl *Binlog) Append(bts []byte) error {
	bl.wrMutex.Lock()
	defer bl.wrMutex.Unlock()
	if len(bts) != bl.blockLength {
		return fmt.Errorf("data size not %d, %d", bl.blockLength, len(bts))
	}
	_, e := bl.f.Write(bts)
	if e != nil {
		return e
	}
	return nil
}

func (bl *Binlog) Archive(path string) error {
	bl.wrMutex.Lock()
	defer bl.wrMutex.Unlock()

	e := bl.f.Close()
	if e != nil {
		return e
	}

	srcFile, e := os.OpenFile(bl.f.Name(), os.O_RDONLY, 0777)
	if e != nil {
		return e
	}

	newFile, e := os.Create(path)
	if e != nil {
		return e
	}
	_, e = io.Copy(newFile, srcFile)
	if e != nil {
		return e
	}
	e = srcFile.Close()
	if e != nil {
		return e
	}
	e = os.Truncate(srcFile.Name(), 0)

	bl.f, e = os.OpenFile(srcFile.Name(), os.O_WRONLY|os.O_APPEND, 0777)
	if e != nil {
		return e
	}

	return nil
}

func (bl *Binlog) Release() error {
	return bl.f.Close()
}
