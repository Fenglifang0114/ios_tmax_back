package util

import (
	"fmt"

	"tmaxsrv/comm"
)

type CmdID uint16

type CmdMap map[CmdID]comm.RespMsgType

var (
	CmdsMap     CmdMap
	CmdsRespMap CmdMap
)

func init() {
	CmdsMap = make(CmdMap)
	CmdsRespMap = make(CmdMap)
}

func (m CmdMap) Set(id CmdID, s comm.RespMsgType) {
	m[id] = s
}

// func GetString(id CmdID) (string, error) {
// 	s, ok := cmdsMap[id]
// 	if ok {
// 		return s.(string), nil
// 	} else {
// 		return "", fmt.Errorf("not found id: %v", id)
// 	}
// }

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
