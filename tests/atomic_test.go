package tests

import (
	"github.com/sudachen/go-ml/fu"
	"gotest.tools/assert"
	"testing"
)

func Test_Atomic1(t *testing.T) {
	f := fu.AtomicFlag{1}
	assert.Assert(t, f.State() == true)
	f.Clear()
	assert.Assert(t, f.State() == false)
	f.Set()
	assert.Assert(t, f.State() == true)
	f.Clear()
	assert.Assert(t, f.State() == false)

	f = fu.AtomicFlag{0}
	assert.Assert(t, f.State() == false)
	f.Clear()
	assert.Assert(t, f.State() == false)
	f.Set()
	assert.Assert(t, f.State() == true)
}
