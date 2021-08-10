# Rotate and Format Logger
The logger package provides a daily, max files, max file-size rotate writer and a customizable configuration format logger implementation.

## Installation
```bash
go get -u github.com/stella-go/logger
```

## Getting Started
### Simple Logging Example
```go
package main

import (
	"github.com/stella-go/logger"
)

func main() {
	logger.INFO("RootInfo")

    mainLogger := rootLogger.GetLogger("Main")
	mainLogger.WARN("MainWarning")
}
```
In the above example, the log will be printed in the log file `./log/log.txt` and the `console` at the same time. The log level can be set by the environment variable `STELLA_LOGGER_LEVEL`, the log path can be set by `STELLA_LOGGER_PATH`, and the log filename can be set by `STELLA_LOGGER_FILE`. The default maximum number of files is 31, and the maximum file size is 200MB. They cannot be modified in this example.

The following methods have the same effect.
```go
package main

import (
	"github.com/stella-go/logger"
)

func main() {
	rootLogger := logger.NewRotateRootLogger(logger.InfoLevel, "./logs", "example.log")
	rootLogger.INFO("RootInfo")

	mainLogger := rootLogger.GetLogger("Main")
	mainLogger.WARN("MainWarning")
}
```

### Custom format and rotate configuration
```go
package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/stella-go/logger"
)

const (
	logLevel = "INFO"
)

type MyLoggerFormatter struct{}

func (*MyLoggerFormatter) Format(e *logger.Entry) []byte {
	timestamp := time.Now().Format("06/01/02 15:04:05")
	msg := fmt.Sprintf("%s %s %s | %s\n", timestamp, strings.ToUpper(e.Level.String()), e.Tag, e.Message)
	return []byte(msg)
}

func main() {
	writer, err := logger.NewConfigRotateWriter(&logger.RotateConfig{
		Enable:      true,
		Daily:       true,
		MaxFiles:    31,
		MaxFileSize: 200 * logger.FileSizeM,
		FilePath:    "./logs",
		FileName:    "custom-log.txt",
	})
	if err != nil {
		panic(err)
	}
	formater := &MyLoggerFormatter{}

	rootLogger := logger.NewRootLogger(logger.Parse(logLevel), formater, writer)
	rootLogger.INFO("RootInfo")

	mainLogger := rootLogger.GetLogger("Main")
	mainLogger.WARN("MainWarning")
}
```

### Format Message
```go
package main

import (
	"errors"

	"github.com/stella-go/logger"
)

func main() {
	logger.INFO("This message is from %s", "Bob")

	err := errors.New("something wrong")
	logger.ERROR("Something went wrong when printing log: ", err)
}
```
**NOTICE**:If the last parameter is an error, do not use placeholders in the format.

### Leveled Logging
```go
func DEBUG(format string, arr ...interface{})

func INFO(format string, arr ...interface{})

func WARN(format string, arr ...interface{})

func ERROR(format string, arr ...interface{})
```