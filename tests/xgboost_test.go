package tests

import (
	"fmt"
	"github.com/sudachen/go-ml/xgboost"
	"testing"
)

func Test_XgboostVersion(t *testing.T) {
	v := xgboost.LibVersion()
	fmt.Println(v)
}
