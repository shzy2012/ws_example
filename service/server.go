package service

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{} // use default options

func echo(w http.ResponseWriter, r *http.Request) {
	done := make(chan struct{})
	ch := make(chan []byte, 0)
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer func() {
		close(done)
		close(ch)
		c.Close()
	}()

	//4:回写消息
	go func() {
		for {
			data, ok := <-ch
			if !ok {
				log.Printf("close:%s\n", r.RemoteAddr)
				return
			}

			err := c.WriteMessage(websocket.BinaryMessage, data)
			if err != nil {
				log.Println("write:", err)
			}
		}
	}()

	//定义用于请求的ws的客户端
	u := url.URL{Scheme: "ws", Host: "103.242.175.164:18080", Path: "/asr/streaming"}
	log.Printf("connecting to %s", u.String())
	ws, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer ws.Close()

	//3:读取消息
	go func() {
		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
			ch <- message
		}
	}()

	for {
		//1:接受消息
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		// log.Printf("recv: %s", message)

		//2:处理消息
		err = ws.WriteMessage(websocket.BinaryMessage, message)
		if err != nil {
			log.Println("write:", err)
			continue
		}
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("i am working"))
}

//Serve 启动服务
func Serve(ctx context.Context, port int) {
	http.HandleFunc("/asr/streaming", echo)
	http.HandleFunc("/", home)

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("listen on:%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
