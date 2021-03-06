package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/shzy2012/common/log"
)

var upgrader = websocket.Upgrader{} // use default options

func echo(w http.ResponseWriter, r *http.Request) {

	done := make(chan struct{})
	ch := make(chan string)
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade:%s\n", err)
		return
	}

	//定义用于请求的ws的客户端
	// u := url.URL{Scheme: "ws", Host: "103.242.175.164:18080", Path: "/asr/streaming?content-type=audio/x-raw,+layout=(string)interleaved,+rate=(int)8000,+format=(string)S16LE,+channels=(int)1"}
	cu := "ws://103.242.175.164:18080/asr/streaming" + r.RequestURI[strings.Index(r.RequestURI, "?"):]
	log.Printf("connecting to %s", cu)
	ws, _, err := websocket.DefaultDialer.Dial(cu, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer ws.Close()

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

			log.Printf("data length:%v\n", len(data))
			log.Printf("data :%s\n", data)
			//注意:返回TextMessage
			err := c.WriteMessage(websocket.TextMessage, []byte(data))
			if err != nil {
				log.Println("write:", err)
			}
		}
	}()

	//3:读取消息
	go func() {
		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv message service: %v\n", len(message))
			ch <- fmt.Sprintf("%s", message)
		}
	}()

	for {
		//1:接受消息
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		log.Printf("recv message from client: %v\n", len(message))

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
