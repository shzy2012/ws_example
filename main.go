package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"ws_example/service"
)

func main() {
	log.Println("[server]=> starting.")
	var port int
	flag.IntVar(&port, "p", 8000, "http 端口号  -p=8000")
	flag.Parse()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		service.Serve(ctx, port)
	}()

	go func() {
		sig := <-sigs
		log.Println(sig)
		log.Println("Shuting down server...")
		cancel()
		done <- true
	}()

	<-done
	log.Println("[server]=>stop service.")
}
