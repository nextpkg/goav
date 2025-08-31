package control

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewControl(t *testing.T) {
	at := assert.New(t)

	c := NewControl(false)
	at.NotNil(c)

	at.False(c.IsDone())

	c.Cancel()
	at.True(c.isDone.Load())

	c.Restart()
	at.False(c.IsDone())
}

func TestNewControlAtDone(t *testing.T) {
	at := assert.New(t)

	c := NewControl(true)
	at.NotNil(c)

	at.True(c.isDone.Load())
}
