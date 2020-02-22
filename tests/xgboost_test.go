package tests

import (
	"fmt"
	"github.com/sudachen/go-ml/xgb"
	"testing"
)

func Test_XgboostVersion(t *testing.T) {
	v := xgb.LibVersion()
	fmt.Println(v)
}
