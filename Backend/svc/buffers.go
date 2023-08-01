package svc

import (
	"fmt"
)

type CmdID uint16

type CmdMap map[CmdID]RespMsgType

var cmdsMap CmdMap
var cmdsRespMap CmdMap

func (m CmdMap) Set(id CmdID, s string) {
	m[id] = RespMsgType(s)
}

func GetString(id CmdID) (string, error) {
	s, ok := cmdsMap[id]
	if ok {
		return string(s), nil
	} else {
		return "", fmt.Errorf("not found id: %v", id)
	}
}

type RingBuffers map[string]*CircularBuffer

func NewRingBuffers(cmds ...string) RingBuffers {
	ringBuffers := make(RingBuffers)
	for _, cmd := range cmds {
		ringBuffers[cmd] = NewCircularBuffer(2048)
	}
	return ringBuffers
}

func (r RingBuffers) Write(cmd string, data []byte) error {
	if _, ok := r[cmd]; !ok {
		return fmt.Errorf("unknown command type: %s", cmd)
	}
	err := r[cmd].EnqueueN(data, len(data))
	return err
}

func (r RingBuffers) Read(cmd string, n int) ([]byte, error) {
	if _, ok := r[cmd]; !ok {
		return nil, fmt.Errorf("unknown command type: %s", cmd)
	}
	return r[cmd].DequeueN(n)
}
