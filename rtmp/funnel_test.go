package core

import (
	"bytes"
	"testing"
	"time"

	"git.code.oa.com/idc/vdn/v4/base"
	"git.code.oa.com/idc/vdn/v4/packet"
	"github.com/stretchr/testify/assert"
)

type testClient struct {
	*Universal
	buf    bytes.Buffer
	before bool
	after  bool
}

func newTestClient() *testClient {
	info := base.NewInfo("t_app", "t_ins", false)

	return &testClient{
		Universal: NewUniversal(info),
	}
}

func (c *testClient) Write(p *packet.Packet) error {
	_, err := c.buf.Write(p.Data)
	return err
}

func (c *testClient) Before() {
	c.before = true
}

func (c *testClient) After() {
	c.after = true
}

func TestNewStream(t *testing.T) {
	at := assert.New(t)

	c := newTestClient()
	s := NewFunnel(c)

	err := s.Write(&packet.Packet{
		Data: make([]byte, 1024),
	})
	at.Nil(err)

	time.Sleep(10 * time.Millisecond)

	at.True(c.before)
	at.False(c.after)

	s.Close()

	time.Sleep(10 * time.Millisecond)

	at.Equal(1024, c.buf.Len())

	at.True(c.before)
	at.True(c.after)

	s.Wait()
}
