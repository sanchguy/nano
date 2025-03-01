// Copyright (c) nano Author. All Rights Reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package nano

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/websocket"
	"strings"
	"sync/atomic"
	"time"
)

var running int32

func listen(addr string, isWs bool, opts ...Option) {
	listenTLS(addr, isWs, "", "", opts...);
}

func listenTLS(addr string, isWs bool, certificate string, key string, opts ...Option) {
	// mark application running
	if atomic.AddInt32(&running, 1) != 1 {
		logger.Println("Nano has running")
		return
	}

	for _, opt := range opts {
		opt(handler.options)
	}

	// cache heartbeat data
	hbdEncode()

	// initial all components
	startupComponents()

	// create global ticker instance, timer precision could be customized
	// by SetTimerPrecision
	globalTicker = time.NewTicker(timerPrecision)

	// startup logic dispatcher
	go handler.dispatch()

	go func() {
		if isWs {
			if len(certificate) != 0 {
				listenAndServeWSTLS(addr, certificate, key)
			} else {
				listenAndServeWS(addr)
			}
		} else {
			listenAndServe(addr)
		}
	}()

	logger.Println(fmt.Sprintf("starting application %s, listen at %s", app.name, addr))
	sg := make(chan os.Signal)
	signal.Notify(sg, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM)

	// stop server
	select {
	case <-env.die:
		logger.Println("The app will shutdown in a few seconds")
	case s := <-sg:
		logger.Println("got signal", s)
	}

	logger.Println("server is stopping...")

	// shutdown all components registered by application, that
	// call by reverse order against register
	shutdownComponents()
	atomic.StoreInt32(&running, 0)
}

// Enable current server accept connection
func listenAndServe(addr string) {
	//listener, err := net.Listen("tcp", addr)
	//if err != nil {
	//	logger.Fatal(err.Error())
	//}
	//
	//defer listener.Close()
	//for {
	//	conn, err := listener.Accept()
	//	if err != nil {
	//		logger.Println(err.Error())
	//		continue
	//	}
	//
	//	go handler.handle(conn)
	//}
}

func listenAndServeWS(addr string) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		//CheckOrigin:     env.checkOrigin,
	}

	http.HandleFunc("/"+strings.TrimPrefix(env.wsPath, "/"), func(w http.ResponseWriter, r *http.Request) {
		info := r.URL.Query()
		//logger.Println("get websocket~~~~~",info)
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Println(fmt.Sprintf("Upgrade failure, URI=%s, Error=%s", r.RequestURI, err.Error()))
			return
		}
		data ,ok := info["postData"]
		if !ok {
			panic("没有postData数据")
		}
		postData := data[0]
		//ti, ok := info["timestamp"]
		//if !ok {
		//	panic("没有timestamp数据")
		//}
		//t := ti[0]
		//n, ok := info["nonce"]
		//if !ok {
		//	panic("没有nonce数据")
		//}
		//nonce := n[0]
		//s, ok := info["sign"]
		//if !ok {
		//	panic("没有sign数据")
		//}
		//sign := s[0]
		logger.Println("LoginInfo %v", info)

		handler.handleWS(conn,[]byte(postData))
	})

	if err := http.ListenAndServe(addr, nil); err != nil {
		logger.Fatal(err.Error())
	}
}

func listenAndServeWSTLS(addr string, certificate string, key string) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     env.checkOrigin,
	}

	http.HandleFunc("/"+strings.TrimPrefix(env.wsPath, "/"), func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Println(fmt.Sprintf("Upgrade failure, URI=%s, Error=%s", r.RequestURI, err.Error()))
			return
		}

		handler.handleWS(conn,[]byte{})
	})

	if err := http.ListenAndServeTLS(addr, certificate, key, nil); err != nil {
		logger.Fatal(err.Error())
	}
}
