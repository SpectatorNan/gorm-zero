package utils

import (
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

var gormSourceDir string
var gormZeroSourceDir string

func init() {
	_, file, _, _ := runtime.Caller(0)
	// compatible solution to get gorm source directory with various operating systems
	gormSourceDir = gormSourceDirPath(file)
	gormZeroSourceDir = gormZeroSourceDirPath(file)
}

func gormSourceDirPath(file string) string {
	dir := filepath.Dir(file)
	dir = filepath.Dir(dir)

	s := filepath.Dir(dir)
	if filepath.Base(s) != "gorm.io" {
		s = dir
	}
	return filepath.ToSlash(s) + "/"
}

func gormZeroSourceDirPath(file string) string {
	dir := filepath.Dir(file)
	dir = filepath.Dir(dir)

	s := filepath.Dir(dir)
	if filepath.Base(s) != "gorm-zero" {
		s = dir
	}
	return filepath.ToSlash(s) + "/"
}

// FileWithLineNum return the file name and line number of the current file
func FileWithLineNum() string {

	// the second caller usually from gorm internal, so set i start from 2
	for i := 2; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)
		inGorm := strings.Contains(file, "gorm.io")
		inGormZero := strings.HasPrefix(file, gormZeroSourceDir)
		if ok && (!(inGorm || inGormZero) || strings.HasSuffix(file, "_test.go")) {
			return file + ":" + strconv.FormatInt(int64(line), 10)
		}
	}

	return ""
}
