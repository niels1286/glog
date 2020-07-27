// @Title
// @Description
// @Author  Niels  2020/7/22
package glog

import (
	"testing"
)

func TestGetLogger(t *testing.T) {
	GetLogger("").Info.Println("asdfasdfasdf")
}
