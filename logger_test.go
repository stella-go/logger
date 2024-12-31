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

package logger_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stella-go/logger"
)

func TestDefaultRootLogger(t *testing.T) {
	os.Setenv("STELLA_LOGGER_LEVEL", "INFO")
	logger.DEBUG("12345678901234567890123456789012")
	logger.INFO("12345678901234567890123456789012")
	logger.WARN("12345678901234567890123456789012")
	logger.ERROR("12345678901234567890123456789012", fmt.Errorf("this is an error"))
}

func TestRotateRootLogger(t *testing.T) {
	logger := logger.NewRotateRootLogger(logger.InfoLevel, "./logs", "stella-go.log")
	logger.DEBUG("12345678901234567890123456789012")
	logger.INFO("12345678901234567890123456789012")
	logger.WARN("12345678901234567890123456789012")
	logger.ERROR("12345678901234567890123456789012", fmt.Errorf("this is an error"))
}

func TestNewDefaultInternalLogger(t *testing.T) {
	rootLogger := logger.NewRotateRootLogger(logger.InfoLevel, "./logs", "stella-go.log")
	logger := rootLogger.GetLogger("Test")
	logger.DEBUG("12345678901234567890123456789012")
	logger.INFO("12345678901234567890123456789012")
	logger.WARN("12345678901234567890123456789012")
	logger.ERROR("12345678901234567890123456789012", fmt.Errorf("this is an error"))
}

func TestPrintf(t *testing.T) {
	rootLogger := logger.NewRotateRootLogger(logger.InfoLevel, "./logs", "stella-go.log")
	rootLogger.Printf("1234567890123%s4567890123456789012", "*****")
}

func TestWrite(t *testing.T) {
	rootLogger := logger.NewRotateRootLogger(logger.InfoLevel, "./logs", "stella-go.log")
	rootLogger.Write([]byte("1234567890123%s4567890123456789012"))
}

type NopFormatter struct{}

func (*NopFormatter) Format(e *logger.Entry) []byte {
	return []byte(e.Message)
}

func TestNewInternalLogger(t *testing.T) {
	rootLogger := logger.NewRootLogger(logger.InfoLevel, &NopFormatter{}, os.Stdout)
	logger := rootLogger.GetLogger("Bench")
	logger.DEBUG("12345678901234567890123456789012")
	logger.INFO("12345678901234567890123456789012")
	logger.WARN("12345678901234567890123456789012")
	logger.ERROR("12345678901234567890123456789012", fmt.Errorf("this is an error"))
}

func BenchmarkLogger(b *testing.B) {
	b.ReportAllocs()
	rootLogger := logger.NewRotateRootLogger(logger.DebugLevel, "./logs", "stella-go.log")
	logger := rootLogger.GetLogger("Bench")
	for i := 0; i < b.N; i++ {
		logger.INFO("12345678901234567890123456789012")
	}
}
func TestPatternFormatter(t *testing.T) {
	pattern := "%d{2006-01-02 15:04:05} [%g] %p %c - %m"
	formatter := &logger.PatternFormatter{Pattern: pattern}
	entry := &logger.Entry{
		Tag:     "Test",
		Level:   logger.InfoLevel,
		Message: "This is a test message",
	}
	formatted := formatter.Format(entry)
	fmt.Println(string(formatted))
}

func TestPatternFormatter2(t *testing.T) {
	pattern := "%d{2006-01-02 15:04:05 [%g] %p %c - %m"
	formatter := &logger.PatternFormatter{Pattern: pattern}
	entry := &logger.Entry{
		Tag:     "Test",
		Level:   logger.InfoLevel,
		Message: "This is a test message",
	}
	formatted := formatter.Format(entry)
	fmt.Println(string(formatted))
}

func TestPatternFormatter3(t *testing.T) {
	pattern := "%d2006-01-02 15:04:05 [%g] %p %c - %m"
	formatter := &logger.PatternFormatter{Pattern: pattern}
	entry := &logger.Entry{
		Tag:     "Test",
		Level:   logger.InfoLevel,
		Message: "This is a test message",
	}
	formatted := formatter.Format(entry)
	fmt.Println(string(formatted))
}

func TestPatternFormatterWithInvalidPattern(t *testing.T) {
	pattern := "%d{invalid} %p %c - %m"
	formatter := &logger.PatternFormatter{Pattern: pattern}
	entry := &logger.Entry{
		Tag:     "Test",
		Level:   logger.ErrorLevel,
		Message: "Error message",
	}
	formatted := formatter.Format(entry)
	fmt.Println(string(formatted))
}

func TestNewPatternFormatterLogger(t *testing.T) {
	rootLogger := logger.NewRootLogger(logger.InfoLevel, &logger.PatternFormatter{Pattern: "%d{06-01-02.15:04:05.000} [%g] %p %c - %m"}, os.Stdout)
	logger := rootLogger.GetLogger("Bench")
	logger.DEBUG("12345678901234567890123456789012")
	logger.INFO("12345678901234567890123456789012")
	logger.WARN("12345678901234567890123456789012")
	logger.ERROR("12345678901234567890123456789012", fmt.Errorf("this is an error"))
}
