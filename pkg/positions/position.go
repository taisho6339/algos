package positions

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

type Position struct {
	// this is unnecessary because only a goroutine accesses this position file and no shared this among multiple ones
	lock *sync.RWMutex
	path string
}

func NewPosition(path string) *Position {
	return &Position{
		lock: &sync.RWMutex{},
		path: path,
	}
}

func (p *Position) ReadLastID() string {
	p.lock.RLock()
	defer p.lock.RUnlock()
	target := filepath.Clean(p.path)
	buf, err := ioutil.ReadFile(target)
	if err != nil {
		return ""
	}
	return string(buf)
}

func (p *Position) Save(id string) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	target := filepath.Clean(p.path)
	tmp := fmt.Sprintf("%s-tmp", target)
	file, err := os.Create(tmp)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write([]byte(id))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(tmp, []byte(id), fs.FileMode(0600))
	if err != nil {
		return err
	}
	return os.Rename(tmp, target)
}
