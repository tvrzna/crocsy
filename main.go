package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	conf, err := InitConfig(os.Args)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("-- starting crocsy")
	for _, server := range conf.Servers {
		startServer(&server)
	}
	handleStop()
}

// Handles stop of web server on signal.
func handleStop() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	<-ch

	log.Print("-- stopping crocsy")
}
