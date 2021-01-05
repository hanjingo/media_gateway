package util

import (
	"fmt"
	"testing"
)

// go test -v help_test.go help.go const.go -test.run TestGetSubDirs
func TestGetSubDirs(t *testing.T) {
	fmt.Println(GetSubDirs("C:\\"))
}
