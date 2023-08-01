package main

import (
	"fmt"
	"net"
	_ "net/http/pprof"
	"os"
	"os/exec"
	"runtime/pprof"
	"strings"
	"time"

	"tmaxsrv/build"
	"tmaxsrv/log"
	"tmaxsrv/svc"
)

var Version = "1.0.0"

const (
	INSTANCE_PORT = 9292
)

func main() {
	fmt.Println("Version:\t", Version)
	fmt.Println("build.Time:\t", build.Time)
	fmt.Println("build.User:\t", build.User)

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", INSTANCE_PORT))
	if err != nil {
		if strings.Index(err.Error(), "in use") != -1 {
			//optionally send command line arguments to the other instance
			fmt.Fprintln(os.Stderr, "Already running.")
			return
		} else {
			panic(err)
		}
	}
	defer listener.Close()

	app := exec.Command("./ui/_ui.exe")
	go app.Run()

	runMode := os.Getenv("RUN_MODE")
	if runMode == "DEV" {
		cpuf, err := os.Create("cpu_profile")
		if err != nil {
			log.Log.Error(err)
		}
		pprof.StartCPUProfile(cpuf)
		defer pprof.StopCPUProfile()
	}

	s := svc.NewScaleMgr() // srvMgr should be set via SetScaleMgr
	// channel to inform application to quit
	quitch := make(chan bool)
	h := *svc.NewSrvMgr(s, quitch)

	s.SetSrvMsg(&h)
	go s.Run()

	go h.Run()
	ws := svc.NewWsServer()
	go ws.Start(&h)
	// <-quitch // wait for user to quit this application
	// Wait for the process to complete
	time.Sleep(10 * time.Second)

	err = app.Wait()
	if err != nil {
		log.Log.Errorf("Error: %v\n", err)
	}
	log.Log.Info("Process completed.")

	if runMode == "DEV" {
		memprof, _ := os.Create("mem.pprof")
		pprof.WriteHeapProfile(memprof)
		memprof.Close()
	}

	return
}
