package logger

import (
	"bufio"
	"os"
	"path/filepath"
	"sync"
)

// FileAdapterConfig sturcture
type FileAdapterConfig struct {
	Level     Level
	Filename  string
	Truncate  bool
	Rotate    bool
	AutoFlush bool
	Format    string
	MaxLength uint32
}

// NewFileAdapterConfig returns a new FileAdapterConfig instance.
func NewFileAdapterConfig() *FileAdapterConfig {
	return &FileAdapterConfig{
		Level:     LevelDebug,
		Truncate:  false,
		Rotate:    false,
		AutoFlush: false,
		Format:    DefaultFormat,
		MaxLength: 0,
	}
}

func (c *FileAdapterConfig) id() AdapterID {
	return AdapterFile
}

///////////////////////////////////////////////////////////////////////

type fileAdapter struct {
	lock   sync.Mutex // file should be protected
	last   string
	file   *os.File
	writer *bufio.Writer
	config FileAdapterConfig
	w      logWriter
}

func newFileAdapter() adapter {
	return &fileAdapter{}
}

func (a *fileAdapter) id() AdapterID {
	return AdapterFile
}

func (a *fileAdapter) init(c AdapterConfig) error {
	cc, ok := c.(*FileAdapterConfig)
	if !ok {
		return ErrInvalidConfig
	}

	if cc.Level < LevelDebug || cc.Level > LevelFatal {
		return ErrInvalidLevel
	}

	a.config = *cc // deep copy
	a.w = makeWriter(cc.Format, cc.MaxLength, "", "")
	if err := a.openFile(); err != nil {
		return err
	}

	return nil
}

func (a *fileAdapter) rotateFile() {
	a.closeFile()
	newName := a.config.Filename + "." + a.last
	os.Rename(a.config.Filename, newName)
	a.openFile()
}

func (a *fileAdapter) openFile() error {
	// path check and create
	dir, _ := filepath.Split(a.config.Filename)
	if len(dir) > 0 {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				return err
			}
		}
	}

	flags := os.O_WRONLY | os.O_CREATE
	if a.config.Truncate {
		flags |= os.O_TRUNC
	} else {
		flags |= os.O_APPEND
	}

	f, err := os.OpenFile(a.config.Filename, flags, 0644)
	if err != nil {
		return err
	}

	a.file = f
	a.writer = bufio.NewWriter(f)
	return nil
}

func (a *fileAdapter) closeFile() {
	if a.writer != nil {
		a.writer.Flush()
	}
	if a.file != nil {
		a.file.Close()
	}
	a.writer = nil
	a.file = nil
}

func (a *fileAdapter) uninit() {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.closeFile()
	a.last = ""
	a.w = nil
}

func (a *fileAdapter) write(msg *message) {
	a.lock.Lock()
	defer a.lock.Unlock()

	if a.writer == nil {
		return
	}

	if a.config.Level > msg.level {
		return
	}

	if a.config.Rotate {
		if dayNow := msg.time.Format("20060102"); dayNow != a.last {
			if len(a.last) > 0 {
				a.rotateFile()
			}
			a.last = dayNow
		}
	}

	a.w(a.writer, msg, nil, nil)
	if a.config.AutoFlush {
		a.writer.Flush()
	}
}

func (a *fileAdapter) flush() {
	a.lock.Lock()
	defer a.lock.Unlock()

	if a.writer != nil {
		a.writer.Flush()
	}
}
