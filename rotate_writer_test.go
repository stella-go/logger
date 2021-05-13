// Copyright 2010-2021 the original author or authors.

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
	"testing"

	"github.com/stella-go/logger"
)

func TestNewRotateWriter(t *testing.T) {
	writer, _ := logger.NewRotateWriter("./logs", "stella-go.log")
	writer.Write([]byte("abc"))
}

func TestNewConfigRotateWriter(t *testing.T) {
	config := &logger.RotateConfig{
		Enable:      true,
		Daily:       true,
		MaxFiles:    5,
		MaxFileSize: 10 * logger.FileSizeB,
		FilePath:    "./logs",
		FileName:    "stella-go-10b.log",
	}
	writer, _ := logger.NewConfigRotateWriter(config)
	for i := 0; i < 100; i++ {
		writer.Write([]byte("1234567890"))
	}
}

func TestNewConfigRotateWriter2(t *testing.T) {
	config := &logger.RotateConfig{
		FileName: "stdout",
	}
	writer, _ := logger.NewConfigRotateWriter(config)
	for i := 0; i < 100; i++ {
		writer.Write([]byte("1234567890"))
	}
}

func TestNewConfigRotateWriter3(t *testing.T) {
	config := &logger.RotateConfig{
		FileName: "stderr",
	}
	writer, _ := logger.NewConfigRotateWriter(config)
	for i := 0; i < 100; i++ {
		writer.Write([]byte("1234567890"))
	}
}
