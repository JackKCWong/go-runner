package util

import (
	"github.com/stretchr/testify/require"
	"testing"
)

type Expect struct {
	t *testing.T
}

func NewExpect(t *testing.T) *Expect {
	return &Expect{t}
}

func (e Expect) Nil(v interface{}) {
	require.Nil(e.t, v)
}

func (e Expect) Nilf(v interface{}, msg string, args ...interface{}) {
	require.Nilf(e.t, v, msg, args)
}

func (e Expect) True(cond bool, args ...interface{}) {
	require.True(e.t, cond, args)
}

func (e Expect) Equal(exp interface{}, actual interface{}) {
	require.Equal(e.t, exp, actual)
}
