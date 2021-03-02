package main

import (
	// "log"
	"io/ioutil"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shzy2012/common/log"
)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	//39.96.21.121
	url := "ws://192.168.55.1:8001/asr/streaming?content-type=audio/x-raw,+layout=(string)interleaved,+rate=(int)8000,+format=(string)S16LE,+channels=(int)1"
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	go func() {
		// for {

		bytes, err := ioutil.ReadFile("16k_16bit.wav")
		if err != nil {
			log.Println("write:", err)
			return
		}
		err = c.WriteMessage(websocket.TextMessage, bytes)
		if err != nil {
			log.Println("write:", err)
			return
		}
		// time.Sleep(time.Second * 1)
		// }
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
