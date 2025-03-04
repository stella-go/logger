// Copyright 2010-2025 the original author or authors.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// 	http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Level int

func (level Level) String() string {
	switch level {
	case TraceLevel:
		return "TRACE"
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO "
	case WarnLevel:
		return "WARN "
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	case PanicLevel:
		return "PANIC"
	default:
		return "LEVEL"
	}
}

func Parse(slevel string) Level {
	slevel = strings.TrimSpace(strings.ToUpper(slevel))
	switch slevel {
	case "TRACE":
		return TraceLevel
	case "DEBUG":
		return DebugLevel
	case "INFO":
		return InfoLevel
	case "WARN":
		return WarnLevel
	case "ERROR":
		return ErrorLevel
	case "FATAL":
		return FatalLevel
	case "PANIC":
		return PanicLevel
	default:
		return InfoLevel
	}
}

const (
	TraceLevel Level = iota
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
	PanicLevel
)

type Entry struct {
	Tag     string
	Level   Level
	Message string
}

type LogFormatter interface {
	Format(e *Entry) []byte
}

type DefaultFormatter struct{}

func (*DefaultFormatter) Format(e *Entry) []byte {
	ts := time.Now().Format("06-01-02.15:04:05.000")
	ts = ts + "000000000000000000000"
	timestamp := ts[:21]
	goroutine := gid()
	msg := fmt.Sprintf("%s [%s] %s %s - %s\n", timestamp, goroutine, e.Level.String(), e.Tag, e.Message)
	return []byte(msg)
}

func gid() string {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return fmt.Sprintf("%s-%-4d", "goroutine", n)
}
func stack(depth int) (string, int) {
	_, file, line, ok := runtime.Caller(depth)
	if !ok {
		return "", 0
	}
	return file, line
}

type PatternFormatter struct {
	Pattern string
}

// Format formats the log entry according to the specified pattern.
// The pattern can contain the following placeholders:
//
//	%d - Date and time of the log entry
//	%p - Log level (e.g., INFO, DEBUG, ERROR)
//	%c - Tag associated with the log entry
//	%m - Log message
//	%g - Goroutine ID
//	%L - File name and line number of the log call
//	%l - Short file name and line number of the log call
//	%% - Literal percent sign
//
// The formatted log entry is returned as a byte slice.
func (p *PatternFormatter) Format(e *Entry) []byte {
	msg := make([]byte, 0, len(p.Pattern)+len(e.Message)+50) // Preallocate buffer
	pattern := p.Pattern
	for i := 0; i < len(pattern); i++ {
		if pattern[i] == '%' && i+1 < len(pattern) {
			switch pattern[i+1] {
			case 'd':
				msg, i = p.appendDate(msg, pattern, i)
			case 'p':
				msg = append(msg, e.Level.String()...)
				i++
			case 'c':
				msg = append(msg, e.Tag...)
				i++
			case 'm':
				msg = append(msg, e.Message...)
				i++
			case 'g':
				msg = append(msg, gid()...)
				i++
			case 'L':
				msg, i = p.appendFileLine(msg, pattern, i, false)
			case 'l':
				msg, i = p.appendFileLine(msg, pattern, i, true)
			case '%':
				msg = append(msg, '%')
				i++
			default:
				msg = append(msg, pattern[i])
			}
		} else {
			msg = append(msg, pattern[i])
		}
	}
	msg = append(msg, '\n')
	return msg
}

func (p *PatternFormatter) appendDate(msg []byte, pattern string, i int) ([]byte, int) {
	start := i + 2
	if pattern[start] == '{' {
		end := strings.Index(pattern[start:], "}")
		if end != -1 {
			end += start
			dateFormat := pattern[start+1 : end]
			msg = append(msg, time.Now().Format(dateFormat)...)
			return msg, end
		}
	}
	ts := time.Now().Format("06-01-02.15:04:05.000")
	ts = ts + "000000000000000000000"
	msg = append(msg, ts[:21]...)
	return msg, i + 1
}

func (p *PatternFormatter) appendFileLine(msg []byte, pattern string, i int, short bool) ([]byte, int) {
	start := i + 2
	if pattern[start] == '{' {
		end := strings.Index(pattern[start:], "}")
		if end != -1 {
			end += start
			sdepth := pattern[start+1 : end]
			depth, err := strconv.Atoi(sdepth)
			if err != nil {
				depth = 4
			}
			file, line := stack(depth)
			if short {
				file = file[strings.LastIndex(file, "/")+1:]
			}
			msg = append(msg, []byte(fmt.Sprintf("%s:%d", file, line))...)
			return msg, end
		}
	}
	file, line := stack(4)
	if short {
		file = file[strings.LastIndex(file, "/")+1:]
	}
	msg = append(msg, []byte(fmt.Sprintf("%s:%d", file, line))...)
	return msg, i + 1
}

type InternalLogger struct {
	level     Level
	formatter LogFormatter
	writer    io.Writer
	lock      sync.Mutex
}

func (l *InternalLogger) formatWrite(e *Entry) (int, error) {
	if e.Level < l.level {
		return 0, nil
	}
	l.lock.Lock()
	defer l.lock.Unlock()
	p := l.formatter.Format(e)
	return l.writer.Write(p)
}

func (l *InternalLogger) write(p []byte) (int, error) {
	l.lock.Lock()
	defer l.lock.Unlock()
	return l.writer.Write(p)
}

type Logger struct {
	tag            string
	internalLogger *InternalLogger
}

func (l *Logger) Printf(format string, arr ...interface{}) (int, error) {
	msg := fmt.Sprintf(format, arr...)
	return l.internalLogger.write([]byte(msg))
}

func (l *Logger) Write(p []byte) (n int, err error) {
	return l.internalLogger.write(p)
}

func (l *Logger) DEBUG(format string, arr ...interface{}) {
	arr, err := splitError(arr...)
	msg := fmt.Sprintf(format, arr...)
	if err != nil {
		msg = fmt.Sprintf("%s %v", msg, err)
	}
	entry := &Entry{
		Tag:     l.tag,
		Level:   DebugLevel,
		Message: msg,
	}
	l.internalLogger.formatWrite(entry)
}

func (l *Logger) INFO(format string, arr ...interface{}) {
	arr, err := splitError(arr...)
	msg := fmt.Sprintf(format, arr...)
	if err != nil {
		msg = fmt.Sprintf("%s %v", msg, err)
	}
	entry := &Entry{
		Tag:     l.tag,
		Level:   InfoLevel,
		Message: msg,
	}
	l.internalLogger.formatWrite(entry)
}

func (l *Logger) WARN(format string, arr ...interface{}) {
	arr, err := splitError(arr...)
	msg := fmt.Sprintf(format, arr...)
	if err != nil {
		msg = fmt.Sprintf("%s %v", msg, err)
	}
	entry := &Entry{
		Tag:     l.tag,
		Level:   WarnLevel,
		Message: msg,
	}
	l.internalLogger.formatWrite(entry)
}

func (l *Logger) ERROR(format string, arr ...interface{}) {
	arr, err := splitError(arr...)
	msg := fmt.Sprintf(format, arr...)
	if err != nil {
		msg = fmt.Sprintf("%s %v", msg, err)
	}
	entry := &Entry{
		Tag:     l.tag,
		Level:   ErrorLevel,
		Message: msg,
	}
	l.internalLogger.formatWrite(entry)
}

func (l *Logger) Level() Level {
	return l.internalLogger.level
}

func (l *Logger) Tag() string {
	return l.tag
}

func (l *Logger) GetLogger(tag string) *Logger {
	return &Logger{
		tag:            tag,
		internalLogger: l.internalLogger,
	}
}

func splitError(arr ...interface{}) ([]interface{}, error) {
	var err error
	if len(arr) > 0 {
		last := arr[len(arr)-1]
		switch last := last.(type) {
		case error:
			arr = arr[:len(arr)-1]
			err = last
		}
	}
	return arr, err
}

func NewRotateRootLogger(level Level, filePath string, fileName string) *Logger {
	rotateWriter, _ := NewRotateWriter(filePath, fileName)
	logger := &InternalLogger{
		level:     level,
		formatter: &DefaultFormatter{},
		writer:    rotateWriter,
		lock:      sync.Mutex{},
	}
	return &Logger{
		tag:            "ROOT",
		internalLogger: logger,
	}
}

func NewRootLogger(level Level, formatter LogFormatter, writer io.Writer) *Logger {
	logger := &InternalLogger{
		level:     level,
		formatter: formatter,
		writer:    writer,
		lock:      sync.Mutex{},
	}
	return &Logger{
		tag:            "ROOT",
		internalLogger: logger,
	}
}

var defaultRootLogger *Logger
var defaultRootLoggerOnce sync.Once

func xInit() {
	defaultRootLoggerOnce.Do(func() {
		slevel := os.Getenv("STELLA_LOGGER_LEVEL")
		if slevel == "" {
			slevel = "INFO"
		}
		spath := os.Getenv("STELLA_LOGGER_PATH")
		if spath == "" {
			spath = "./logs"
		}
		sfile := os.Getenv("STELLA_LOGGER_FILE")
		if sfile == "" {
			sfile = "log.txt"
		}
		level := Parse(slevel)
		rotateWriter, _ := NewRotateWriter(spath, sfile)
		writer := io.MultiWriter(os.Stdout, rotateWriter)

		defaultRootLogger = NewRootLogger(level, &DefaultFormatter{}, writer)
	})
}

func DEBUG(format string, arr ...interface{}) {
	xInit()
	defaultRootLogger.DEBUG(format, arr...)
}

func INFO(format string, arr ...interface{}) {
	xInit()
	defaultRootLogger.INFO(format, arr...)
}

func WARN(format string, arr ...interface{}) {
	xInit()
	defaultRootLogger.WARN(format, arr...)
}

func ERROR(format string, arr ...interface{}) {
	xInit()
	defaultRootLogger.ERROR(format, arr...)
}

func GetLogger(name string) *Logger {
	xInit()
	return defaultRootLogger.GetLogger(name)
}
