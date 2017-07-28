package muxproto

import (
	"github.com/tinylib/msgp/msgp"
)

// ProtoType defines the type of proxied traffic
type ProtoType int

const (
	// TCP - connection delivers TCP traffic
	TCP ProtoType = iota + 1
	// UDP - connection delivers UDP traffic
	UDP
)

// Control describes a channel in a multiplexed connection
type Control struct {
	Protocol ProtoType
	Endpoint string
}

// EncodeMsg satisfes msgp.Encodable
func (c Control) EncodeMsg(w *msgp.Writer) error {
	var err error
	if err = w.WriteArrayHeader(2); err != nil {
		return err
	}
	if err = w.WriteInt(int(c.Protocol)); err != nil {
		return err
	}
	if err = w.WriteString(c.Endpoint); err != nil {
		return err
	}
	return nil
}

// DecodeMsg satisfies msgp.Decodable
func (c *Control) DecodeMsg(r *msgp.Reader) error {
	proto, err := r.ReadInt()
	if err != nil {
		return err
	}
	c.Protocol = ProtoType(proto)
	c.Endpoint, err = r.ReadString()
	if err != nil {
		return err
	}
	return nil
}

var _ msgp.Encodable = Control{}
var _ msgp.Decodable = &Control{}
