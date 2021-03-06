package main

import (
	"flag"
	"os/exec"
	// "vger/cocoa"
	"vger/download"
	"vger/util"
	// "subscribe"
	"vger/website"
)

var debug *bool = flag.Bool("debug", false, "debug")
var config *string = flag.String("config", "", "config file")

func main() {
	flag.Parse()

	util.ConfigPath = *config

	go download.Start()

	if *debug {
		go func() {
			server := util.ReadConfig("server")
			cmd := exec.Command("open", "http://"+server)
			cmd.Run()
		}()
		website.Run(*debug)
	} else {
		go website.Run(*debug)
		//cocoa.Start()
		<-(chan struct{})(nil)
	}
}
