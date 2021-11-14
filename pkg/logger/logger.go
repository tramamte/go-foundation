package logger

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

// Author: Patrick Yu

// Version history
// v0.1: initial implementation
// v0.2: support any type of logging object
// v0.3: add truncation mode to file adapter
// v0.4: add log date/time format (short and long)
// v0.5: singleton, enhance thread safety

const version = "0.5.8"

// GetVersion returns version string.
func GetVersion() string {
	return version
}

// singleton
var lgr *Logger

var lineFeed = "\n"

func init() {
	if runtime.GOOS == "windows" {
		lineFeed = "\r\n"
	}

	lgr = New("Default", false)
	lgr.Attach(NewConsoleAdapterConfig())
}

// GetLogger returns global logger instance
func GetLogger() *Logger {
	return lgr
}

// Substitute replaces global logger instance
func Substitute(l *Logger) {
	if l != nil {
		lgr = l
	}
}

///////////////////////////////////////////////////////////////////////
// Logging Functions
///////////////////////////////////////////////////////////////////////

// Debug for singleton
func Debug(obj interface{}) { lgr.Debug(obj) }

// Debug outputs "debug" level normal string log.
func (logger *Logger) Debug(obj interface{}) {
	logger.lock.RLock()
	defer logger.lock.RUnlock()

	if LevelDebug >= logger.level {
		logger.write(LevelDebug, obj)
	}
}

// Debugf for singleton
func Debugf(format string, arg ...interface{}) { lgr.Debugf(format, arg...) }

// Debugf outputs "debug" level formatted string log.
func (logger *Logger) Debugf(format string, arg ...interface{}) {
	logger.lock.RLock()
	defer logger.lock.RUnlock()

	if LevelDebug >= logger.level && len(format) > 0 {
		log := fmt.Sprintf(format, arg...)
		logger.write(LevelDebug, log)
	}
}

// Verbose for singleton
func Verbose(obj interface{}) { lgr.Verbose(obj) }

// Verbose outputs "verbose" level normal string log.
func (logger *Logger) Verbose(obj interface{}) {
	logger.lock.RLock()
	defer logger.lock.RUnlock()

	if LevelVerbose >= logger.level {
		logger.write(LevelVerbose, obj)
	}
}

// Verbosef for singleton
func Verbosef(format string, arg ...interface{}) { lgr.Verbosef(format, arg...) }

// Verbosef outputs "verbose" level formatted string log.
func (logger *Logger) Verbosef(format string, arg ...interface{}) {
	logger.lock.RLock()
	defer logger.lock.RUnlock()

	if LevelVerbose >= logger.level && len(format) > 0 {
		log := fmt.Sprintf(format, arg...)
		logger.write(LevelVerbose, log)
	}
}

// Information for singleton
func Information(obj interface{}) { lgr.Information(obj) }

// Information outputs "information" level normal string log.
func (logger *Logger) Information(obj interface{}) {
	logger.lock.RLock()
	defer logger.lock.RUnlock()

	if LevelInformation >= logger.level {
		logger.write(LevelInformation, obj)
	}
}

// Informationf for singleton
func Informationf(format string, arg ...interface{}) { lgr.Informationf(format, arg...) }

// Informationf outputs "information" level formatted string log.
func (logger *Logger) Informationf(format string, arg ...interface{}) {
	logger.lock.RLock()
	defer logger.lock.RUnlock()

	if LevelInformation >= logger.level && len(format) > 0 {
		log := fmt.Sprintf(format, arg...)
		logger.write(LevelInformation, log)
	}
}

// Warning for singleton
func Warning(obj interface{}) { lgr.Warning(obj) }

// Warning outputs "warninig" level normal string log.
func (logger *Logger) Warning(obj interface{}) {
	logger.lock.RLock()
	defer logger.lock.RUnlock()

	if LevelWarning >= logger.level {
		logger.write(LevelWarning, obj)
	}
}

// Warningf for singleton
func Warningf(format string, arg ...interface{}) { lgr.Warningf(format, arg...) }

// Warningf outputs "warning" level formatted string log.
func (logger *Logger) Warningf(format string, arg ...interface{}) {
	logger.lock.RLock()
	defer logger.lock.RUnlock()

	if LevelWarning >= logger.level && len(format) > 0 {
		log := fmt.Sprintf(format, arg...)
		logger.write(LevelWarning, log)
	}
}

// Error for singleton
func Error(obj interface{}) { lgr.Error(obj) }

// Error outputs "error" level normal string log.
func (logger *Logger) Error(obj interface{}) {
	logger.lock.RLock()
	defer logger.lock.RUnlock()

	if LevelError >= logger.level {
		logger.write(LevelError, obj)
	}
}

// Errorf for singleton
func Errorf(format string, arg ...interface{}) { lgr.Errorf(format, arg...) }

// Errorf outputs "error" level formatted string log.
func (logger *Logger) Errorf(format string, arg ...interface{}) {
	logger.lock.RLock()
	defer logger.lock.RUnlock()

	if LevelError >= logger.level && len(format) > 0 {
		log := fmt.Sprintf(format, arg...)
		logger.write(LevelError, log)
	}
}

// Panic for singleton
func Panic(obj interface{}) { lgr.Panic(obj) }

// Panic outputs "panic" level normal string log
// when the logger's level is set to less equal LevelPanic
// and is follwed by a call to panic(log).
func (logger *Logger) Panic(obj interface{}) {
	logger.lock.RLock()
	defer logger.lock.RUnlock()

	if LevelPanic >= logger.level {
		logger.write(LevelPanic, obj)
	}
	logger.Flush()
	panic(obj)
}

// Panicf for singleton
func Panicf(format string, arg ...interface{}) { lgr.Panicf(format, arg...) }

// Panicf outputs "panic" level formatted string log
// when the logger's level is set to less equal LevelPanic
// and is followed by a call to panic(...).
func (logger *Logger) Panicf(format string, arg ...interface{}) {
	logger.lock.RLock()
	defer logger.lock.RUnlock()

	log := fmt.Sprintf(format, arg...)
	if LevelPanic >= logger.level && len(format) > 0 {
		logger.write(LevelPanic, log)
	}
	logger.Flush()
	panic(log)
}

// Fatal for singleton
func Fatal(obj interface{}) { lgr.Fatal(obj) }

// Fatal outputs "fatal" level normal string log
// and is followed by a call to os.Exit(1).
func (logger *Logger) Fatal(obj interface{}) {
	logger.lock.RLock()
	defer logger.lock.RUnlock()

	logger.write(LevelFatal, obj)
	logger.Flush()
	os.Exit(1)
}

// Fatalf for singleton
func Fatalf(format string, arg ...interface{}) { lgr.Fatalf(format, arg...) }

// Fatalf outputs "fatal" level formatted string log
// and is followed by a call to os.Exit(1).
func (logger *Logger) Fatalf(format string, arg ...interface{}) {
	logger.lock.RLock()
	defer logger.lock.RUnlock()

	if len(format) > 0 {
		log := fmt.Sprintf(format, arg...)
		logger.write(LevelFatal, log)
	}
	logger.Flush()
	os.Exit(1)
}

// Stack for singleton
func Stack(l Level, bufLen int) { lgr.Stack(l, bufLen) }

// Stack outputs current goroutines's execution stack.
func (logger *Logger) Stack(l Level, bufLen int) {
	logger.lock.RLock()
	defer logger.lock.RUnlock()

	if l < LevelDebug || l > LevelFatal {
		return
	}

	// minimum 1024 bytes
	if bufLen < 1024 {
		bufLen = 1024
	}

	buf := make([]byte, bufLen)
	len := runtime.Stack(buf, false)
	logger.write(l, string(buf[:len]))
}

// Flush for singleton
func Flush() { lgr.Flush() }

// Flush flushes all output adapters
// and flushes message channel for async. logger.
func (logger *Logger) Flush() {
	logger.lock.RLock()
	defer logger.lock.RUnlock()

	logger.flush()
}

///////////////////////////////////////////////////////////////////////

// errors
var (
	ErrNilConfig     = errors.New("config is nil")
	ErrAlreadyExist  = errors.New("adapter already exist")
	ErrInvalidConfig = errors.New("invalid config")
	ErrInvalidLevel  = errors.New("invalid level")
)

// AdapterID is log adapter ID
type AdapterID int

// adapter ID
const (
	AdapterConsole AdapterID = iota
	AdapterFile
)

// Level is log level
type Level int

// log levels
const (
	LevelDebug Level = iota
	LevelVerbose
	LevelInformation
	LevelWarning
	LevelError
	LevelPanic
	LevelFatal
)

var levelString = []string{
	"DBG",
	"VBS",
	"INF",
	"WRN",
	"ERR",
	"PNC",
	"FTL",
}

// Str2Level convert string to Level
func Str2Level(str string) (l Level) {
	lower := strings.ToLower(str)
	switch lower {
	case "verbose", "vbs":
		l = LevelVerbose
	case "information", "inf":
		l = LevelInformation
	case "warning", "wrn":
		l = LevelWarning
	case "error", "err":
		l = LevelError
	case "panic", "pnc":
		l = LevelPanic
	case "fatal", "ftl":
		l = LevelFatal
	default:
		l = LevelDebug
	}
	return
}

// DefaultFormat is default log string format.
const DefaultFormat = "$ltime [$slevel] $msg ($file:$line)"

// Logger structure
type Logger struct {
	name     string
	level    Level
	lock     sync.RWMutex
	adapters []adapter
	async    bool
	msgChan  chan *message
	wait     sync.WaitGroup
}

// AdapterConfig is an interface of configuration for a log output.
type AdapterConfig interface {
	id() AdapterID
}

type adapter interface {
	id() AdapterID
	init(c AdapterConfig) error
	uninit()
	write(m *message)
	flush()
}

// New makes a new Logger instance.
func New(name string, async bool) (logger *Logger) {
	logger = &Logger{
		name:  name,
		level: LevelDebug,
	}
	if len(name) == 0 {
		logger.name = "No Name"
	}

	if async {
		// use core count for channel buffer
		logger.msgChan = make(chan *message, runtime.GOMAXPROCS(0))
		go logger.asyncProc()
		logger.async = true
	}

	return
}

// Attach for singleton
func Attach(config AdapterConfig) error { return lgr.Attach(config) }

// Attach attaches an output adapter.
func (logger *Logger) Attach(config AdapterConfig) error {
	if config == nil {
		return ErrNilConfig
	}

	logger.lock.Lock()
	defer logger.lock.Unlock()

	for _, a := range logger.adapters {
		if a.id() == config.id() {
			return ErrAlreadyExist
		}
	}

	var ctor func() adapter
	switch config.id() {
	case AdapterConsole:
		ctor = newConsoleAdapter
	case AdapterFile:
		ctor = newFileAdapter
	}

	ad := ctor()
	if err := ad.init(config); err != nil {
		return err
	}

	logger.adapters = append(logger.adapters, ad)
	return nil
}

// Detach for singleton
func Detach(id AdapterID) { lgr.Detach(id) }

// Detach detaches an output adapter.
func (logger *Logger) Detach(id AdapterID) {
	if id < AdapterConsole || id > AdapterFile {
		return
	}

	logger.lock.Lock()
	defer logger.lock.Unlock()

	logger.flush()

	var adapters []adapter
	for _, a := range logger.adapters {
		if a.id() == id {
			a.uninit()
			continue
		}
		adapters = append(adapters, a)
	}
	logger.adapters = adapters
}

// SetLevel for singleton
func SetLevel(l Level) error { return lgr.SetLevel(l) }

// SetLevel sets level of logger.
func (logger *Logger) SetLevel(l Level) error {
	if l < LevelDebug || l > LevelFatal {
		return ErrInvalidLevel
	}
	logger.lock.Lock()
	defer logger.lock.Unlock()
	logger.level = l
	return nil
}

///////////////////////////////////////////////////////////////////////

func (logger *Logger) flush() {
	if logger.async {
		logger.wait.Wait()
	}
	for _, a := range logger.adapters {
		a.flush()
	}
}

type message struct {
	name     string
	time     time.Time
	level    Level
	function string
	file     string
	line     int
	msg      interface{}
}

var messageCache = sync.Pool{
	New: func() interface{} {
		return &message{}
	},
}

func (logger *Logger) write(v Level, o interface{}) {
	skip := 2
	if logger == lgr {
		skip = 3
	}

	funcName := "null"
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		file = "null"
		line = 0
	} else {
		funcName = runtime.FuncForPC(pc).Name()
	}
	_, fileName := path.Split(file)

	msg := messageCache.Get().(*message)
	msg.name = logger.name
	msg.time = time.Now()
	msg.level = v
	msg.function = funcName
	msg.file = fileName
	msg.line = line
	msg.msg = o

	if logger.async {
		logger.wait.Add(1)
		logger.msgChan <- msg
	} else {
		logger.writeToOutputs(msg)
	}
}

func (logger *Logger) asyncProc() {
	for m := range logger.msgChan {
		logger.writeToOutputs(m)
		logger.wait.Done()
	}
}

func (logger *Logger) writeToOutputs(msg *message) {
	for _, a := range logger.adapters {
		a.write(msg)
	}
	messageCache.Put(msg)
}

///////////////////////////////////////////////////////////////////////
// log writer
///////////////////////////////////////////////////////////////////////

type logWriter func(w io.Writer, msg *message, prefix, suffix interface{})

func makeWriter(format string, maxMsgLen uint32, prefixHolder, suffixHolder string) logWriter {
	if len(format) == 0 {
		return func(w io.Writer, msg *message, prefix, suffix interface{}) {
		}
	}

	var (
		namePos     = strings.Index(format, "$name")
		ltimePos    = strings.Index(format, "$ltime")
		stimePos    = strings.Index(format, "$stime")
		tsPos       = strings.Index(format, "$ts")
		ilevelPos   = strings.Index(format, "$ilevel")
		slevelPos   = strings.Index(format, "$slevel")
		functionPos = strings.Index(format, "$function")
		filePos     = strings.Index(format, "$file")
		linePos     = strings.Index(format, "$line")
		msgPos      = strings.Index(format, "$msg")
	)

	var argMap = make(map[int]string)

	if namePos != -1 {
		argMap[namePos] = "$name"
		format = strings.Replace(format, "$name", "%s", 1)
	}
	if ltimePos != -1 {
		argMap[ltimePos] = "$ltime"
		format = strings.Replace(format, "$ltime", "%s", 1)
	}
	if stimePos != -1 {
		argMap[stimePos] = "$stime"
		format = strings.Replace(format, "$stime", "%s", 1)
	}
	if tsPos != -1 {
		argMap[tsPos] = "$ts"
		format = strings.Replace(format, "$ts", "%d", 1)
	}
	if ilevelPos != -1 {
		argMap[ilevelPos] = "$ilevel"
		format = strings.Replace(format, "$ilevel", "%d", 1)
	}
	if slevelPos != -1 {
		argMap[slevelPos] = "$slevel"
		format = strings.Replace(format, "$slevel", "%s", 1)
	}
	if functionPos != -1 {
		argMap[functionPos] = "$function"
		format = strings.Replace(format, "$function", "%s", 1)
	}
	if filePos != -1 {
		argMap[filePos] = "$file"
		format = strings.Replace(format, "$file", "%s", 1)
	}
	if linePos != -1 {
		argMap[linePos] = "$line"
		format = strings.Replace(format, "$line", "%d", 1)
	}
	if msgPos != -1 {
		argMap[msgPos] = "$msg"
		if maxMsgLen == 0 {
			format = strings.Replace(format, "$msg", "%+v", 1)
		} else {
			format = strings.Replace(format, "$msg", "%s", 1)
		}
	}

	if len(prefixHolder) > 0 {
		argMap[-1] = "$prefix"
		format = prefixHolder + format
	}
	if len(suffixHolder) > 0 {
		argMap[math.MaxInt32] = "$suffix"
		format = format + suffixHolder
	}

	format += lineFeed

	var argKeys []int
	for k := range argMap {
		argKeys = append(argKeys, k)
	}
	sort.Ints(argKeys)
	argLen := len(argKeys)

	prefixArgPicker := func(msg *message, prefix, suffix interface{}) interface{} {
		return prefix
	}
	suffixArgPicker := func(msg *message, prefix, suffix interface{}) interface{} {
		return suffix
	}
	nameArgPicker := func(msg *message, prefix, suffix interface{}) interface{} {
		return msg.name
	}
	ltimeArgPicker := func(msg *message, prefix, suffix interface{}) interface{} {
		return msg.time.Format("2006-01-02 15:04:05.000")
	}
	stimeArgPicker := func(msg *message, prefix, suffix interface{}) interface{} {
		return msg.time.Format("15:04:05.000")
	}
	tsArgPicker := func(msg *message, prefix, suffix interface{}) interface{} {
		return msg.time.UnixNano() / 1e3
	}
	ilevelArgPicker := func(msg *message, prefix, suffix interface{}) interface{} {
		return int(msg.level)
	}
	slevelArgPicker := func(msg *message, prefix, suffix interface{}) interface{} {
		return levelString[msg.level]
	}
	functionArgPicker := func(msg *message, prefix, suffix interface{}) interface{} {
		return msg.function
	}
	fileArgPicker := func(msg *message, prefix, suffix interface{}) interface{} {
		return msg.file
	}
	lineArgPicker := func(msg *message, prefix, suffix interface{}) interface{} {
		return msg.line
	}
	msgArgPicker := func(msg *message, prefix, suffix interface{}) interface{} {
		if maxMsgLen == 0 {
			return msg.msg
		}

		str := fmt.Sprintf("%+v", msg.msg)
		if uint32(len(str)) > maxMsgLen {
			str = str[:maxMsgLen] + " ..."
		}
		return str
	}

	pickerList := make([]func(msg *message, prefix, suffix interface{}) interface{}, argLen)
	for i, k := range argKeys {
		kStr := argMap[k]
		switch kStr {
		case "$prefix":
			pickerList[i] = prefixArgPicker
		case "$suffix":
			pickerList[i] = suffixArgPicker
		case "$name":
			pickerList[i] = nameArgPicker
		case "$ltime":
			pickerList[i] = ltimeArgPicker
		case "$stime":
			pickerList[i] = stimeArgPicker
		case "$ts":
			pickerList[i] = tsArgPicker
		case "$ilevel":
			pickerList[i] = ilevelArgPicker
		case "$slevel":
			pickerList[i] = slevelArgPicker
		case "$function":
			pickerList[i] = functionArgPicker
		case "$file":
			pickerList[i] = fileArgPicker
		case "$line":
			pickerList[i] = lineArgPicker
		case "$msg":
			pickerList[i] = msgArgPicker
		}
	}

	argComposer := func(msg *message, prefix, suffix interface{}) []interface{} {
		args := make([]interface{}, argLen)
		for i, p := range pickerList {
			a := p(msg, prefix, suffix)
			args[i] = a
		}
		return args
	}

	return func(w io.Writer, msg *message, prefix, suffix interface{}) {
		args := argComposer(msg, prefix, suffix)
		fmt.Fprintf(w, format, args...)
	}
}
