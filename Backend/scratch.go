package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"tmaxsrv/cmd"
	"tmaxsrv/comm"
	"time"
)

func main() {
	now := time.Now().Unix()
	fmt.Printf("Current Unix Timestamp: %d\n", now)
	
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, uint32(now))
	
	composer := comm.CmdComposer{ScaleCat: comm.SCALE_TMAX}
	command, _, err := cmd.ComposeCmdTMAX(&composer, comm.CMD_SET_SCALE_TIME, comm.CmdData{Type: comm.DATA_TYPE_STR, Data: string(bytes)})
	
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	fmt.Printf("Raw bytes sent to scale (Hex):\n%s\n", hex.EncodeToString(command))
	
	// Also format it nicely:
	for i, b := range command {
		fmt.Printf("%02X ", b)
		if (i+1)%16 == 0 {
			fmt.Println()
		}
	}
	fmt.Println()
}
