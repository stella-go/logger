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

package logger

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	FileSizeB = 1
	FileSizeK = 1024
	FileSizeM = 1024 * FileSizeK
	FileSizeG = 1024 * FileSizeM
)

type RotateConfig struct {
	Enable      bool
	Daily       bool
	MaxFiles    int
	MaxFileSize int64
	FilePath    string
	FileName    string
}

type RotateWriter struct {
	config *RotateConfig
	dest   *os.File
}

func (w *RotateWriter) Write(p []byte) (int, error) {
	w.tryRotate()
	return w.dest.Write(p)
}

func (w *RotateWriter) tryRotate() {
	if !w.config.Enable {
		return
	}
	if w.config.Daily {
		if fi, err := w.dest.Stat(); err == nil && fi.ModTime().Local().Format("20060102") != time.Now().Local().Format("20060102") {
			w.rotate()
		}
	}
	if w.config.MaxFileSize > 0 {
		if fi, err := w.dest.Stat(); err == nil && fi.Size() > w.config.MaxFileSize {
			w.rotate()
		}
	}
}

type sfis []os.FileInfo

func (p sfis) Len() int {
	return len(p)
}
func (p sfis) Less(i, j int) bool {
	return p[i].ModTime().After(p[j].ModTime()) // new -> ... -> old
}
func (p sfis) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (w *RotateWriter) rotate() {
	if w.dest == os.Stdout || w.dest == os.Stderr {
		return
	}
	fis, err := ioutil.ReadDir(w.config.FilePath)
	if err != nil {
		print("ERROR", "Get file list error: %v", err)
		return
	}
	sfis := sfis(fis)
	sort.Sort(sfis)
	series := make([]os.FileInfo, 0)
	for _, fileInfo := range sfis {
		if strings.HasPrefix(fileInfo.Name(), w.config.FileName) {
			series = append(series, fileInfo)
		}
	}
	names := make([]string, len(series))
	for i, s := range series {
		names[i] = s.Name()
	}
	fileInfo, err := w.dest.Stat()
	if err != nil {
		print("ERROR", "Get file stat error: %v", err)
	}
	date := fileInfo.ModTime().Local().Format("20060102")
	newName := fmt.Sprintf("%s.%s", w.config.FileName, date)
	index := 1
	for _, name := range names {
		if strings.HasPrefix(name, newName) {
			suffix := name[len(newName):]
			if len(suffix) != 0 {
				i, err := strconv.Atoi(suffix[1:])
				if err != nil {
					print("ERROR", "Parse file index error: %v", err)
				}
				if i >= index {
					index = i + 1
				}
			}
		}
	}
	newName = fmt.Sprintf("%s.%d", newName, index)
	oldPath := path.Join(w.config.FilePath, w.config.FileName)
	newPath := path.Join(w.config.FilePath, newName)
	err = os.Rename(oldPath, newPath)
	if err != nil {
		print("ERROR", "Rename file error: %v", err)
		return
	}
	fo, err := os.OpenFile(oldPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		print("ERROR", "Open file error: %v", err)
		return
	}
	w.dest = fo

	if len(names) > w.config.MaxFiles-1 {
		for _, name := range names[w.config.MaxFiles-1:] {
			p := path.Join(w.config.FilePath, name)
			err := os.Remove(p)
			if err != nil {
				print("ERROR", "Remove file error: %v", err)
			}
		}
	}
}

func NewConfigRotateWriter(config *RotateConfig) (*RotateWriter, error) {
	switch config.FileName {
	case "", "stdout":
		return &RotateWriter{
			config: config,
			dest:   os.Stdout,
		}, nil
	case "stderr":
		return &RotateWriter{
			config: config,
			dest:   os.Stderr,
		}, nil
	default:

	}
	exist, err := isExists(config.FilePath)
	if err != nil {
		return nil, err
	}
	if !exist {
		err := os.MkdirAll(config.FilePath, 0755)
		if err != nil {
			return nil, err
		}
	}
	fo, err := os.OpenFile(path.Join(config.FilePath, config.FileName), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &RotateWriter{
		config: config,
		dest:   fo,
	}, nil
}

func NewRotateWriter(filePath string, fileName string) (*RotateWriter, error) {
	config := &RotateConfig{
		Enable:      true,
		Daily:       true,
		MaxFiles:    31,
		MaxFileSize: 200 * FileSizeM,
		FilePath:    filePath,
		FileName:    fileName,
	}
	return NewConfigRotateWriter(config)
}

func isExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func print(tag string, format string, a ...interface{}) (int, error) {
	msg := fmt.Sprintf(format, a...)
	now := time.Now().Local()
	datetime := now.Format("2006/01/02 15:04:05")
	return fmt.Printf("%s [%s] RotateWriter - %s\n", datetime, tag, msg)
}
