package tmaxbackend

import (
	"fmt"
	"os"
	"time"
	"tmaxsrv/comm"
	"tmaxsrv/svc"
)

var Version = "1.0.0"

const (
	INSTANCE_PORT = 9292
)

// StartBackend is the entry point for Android gomobile integration
func StartBackend(dataDir string) {
	fmt.Println("TmaxService Go Backend Engine Starting on Android...")
	comm.AndroidDataDir = dataDir
	os.Setenv("TMPDIR", dataDir)

	s := svc.NewScaleMgr() 
	quitch := make(chan bool)
	h := *svc.NewSrvMgr(s, quitch)

	s.SetSrvMsg(&h)
	go s.Run()

	go h.Run()
	ws := svc.NewWsServer()
	go ws.Start(&h)

	for {
		time.Sleep(10 * time.Second)
	}
}
