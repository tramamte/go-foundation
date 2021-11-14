package logger

import (
	"io"
	"os"
	"runtime"
	"sync"
)

// ConsoleAdapterConfig sturcture
type ConsoleAdapterConfig struct {
	Level     Level
	Color     bool
	Format    string
	MaxLength uint32
}

// NewConsoleAdapterConfig returns a new ConsoleAdapterConfig instance.
func NewConsoleAdapterConfig() *ConsoleAdapterConfig {
	color := true
	if runtime.GOOS == "windows" {
		// ANSI color code is not supported in windows' default terminal
		color = false
	}

	return &ConsoleAdapterConfig{
		Level:     LevelDebug,
		Color:     color,
		Format:    DefaultFormat,
		MaxLength: 0,
	}
}

func (c *ConsoleAdapterConfig) id() AdapterID {
	return AdapterConsole
}

///////////////////////////////////////////////////////////////////////

type consoleAdapter struct {
	lock   sync.RWMutex
	writer io.Writer
	config ConsoleAdapterConfig
	w      logWriter
}

func newConsoleAdapter() adapter {
	return &consoleAdapter{}
}

func (a *consoleAdapter) id() AdapterID {
	return AdapterConsole
}

func (a *consoleAdapter) init(c AdapterConfig) error {
	cc, ok := c.(*ConsoleAdapterConfig)
	if !ok {
		return ErrInvalidConfig
	}

	if cc.Level < LevelDebug || cc.Level > LevelFatal {
		return ErrInvalidLevel
	}

	a.writer = os.Stdout
	a.config = *cc // deep copy

	var prefix, suffix string
	if cc.Color {
		prefix = "%s"
		suffix = "%s"
	}
	a.w = makeWriter(cc.Format, cc.MaxLength, prefix, suffix)
	return nil
}

func (a *consoleAdapter) uninit() {
	a.lock.Lock()
	defer a.lock.Unlock()

	a.writer = nil
	a.w = nil
}

func (a *consoleAdapter) write(msg *message) {
	a.lock.RLock()
	defer a.lock.RUnlock()

	if a.writer == nil {
		return
	}

	if a.config.Level > msg.level {
		return
	}

	var prefix, suffix interface{}
	if a.config.Color {
		switch msg.level {
		case LevelDebug:
			prefix = prefixCyan
		case LevelVerbose:
			prefix = ""
		case LevelInformation:
			prefix = prefixLGreen
		case LevelWarning:
			prefix = prefixLYellow
		case LevelError:
			prefix = prefixLRed
		case LevelPanic:
			prefix = prefixBPurble
		case LevelFatal:
			prefix = prefixBRed
		}
		suffix = suffixReset
	}
	a.w(a.writer, msg, prefix, suffix)
}

func (a *consoleAdapter) flush() {
}
