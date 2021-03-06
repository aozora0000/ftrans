package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// For 2.0

func serverWorker2_0(connChan <-chan *websocket.Conn, conn1 *websocket.Conn, canceller func()) {
	defer canceller()
	defer conn1.Close()

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			<-ticker.C

			if err := conn1.WriteControl(websocket.PingMessage, []byte("keep-alive"), time.Now().Add(3*time.Second)); err != nil {
				canceller()

				return
			}
		}
	}()

	var ok bool
	var conn2 *websocket.Conn
	if conn2, ok = <-connChan; !ok {
		return
	}

	defer conn2.Close()

	conn1.SetReadDeadline(time.Now().Add(10 * time.Second))
	conn2.SetReadDeadline(time.Now().Add(10 * time.Second))
	conn1.SetWriteDeadline(time.Now().Add(10 * time.Second))
	conn2.SetWriteDeadline(time.Now().Add(10 * time.Second))

	if err := conn1.WriteJSON("CONNECTED"); err != nil {
		return
	}
	if err := conn2.WriteJSON("CONNECTED"); err != nil {
		return
	}

	var wg sync.WaitGroup

	wg.Add(1)

	readWrite := func(a, b *websocket.Conn) {
		var msg Message2_0
		if err := a.ReadJSON(&msg); err != nil {
			return
		}

		b.WriteJSON(msg)
	}

	wg.Add(1)
	go func() {
		readWrite(conn1, conn2)
		wg.Done()
	}()

	readWrite(conn2, conn1)

	wg.Wait()

}

var mut2_0 sync.Mutex
var conns2_0 map[string]chan *websocket.Conn

func init2_0() {
	conns2_0 = make(map[string]chan *websocket.Conn)
}

func serverHandler2_0(rw http.ResponseWriter, req *http.Request) {
	mut := mut2_0
	conns := conns2_0

	log.Printf("connected(addr: %s, version: %s)", req.RemoteAddr, protocolVersion2_0)
	defer log.Println("closed", req.RemoteAddr)
	conn, err := upgrader.Upgrade(rw, req, nil)

	if err != nil {
		log.Println(err)

		return
	}

	conn.SetReadLimit(10 * 1024) // 10KiB

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	var hs Handshake2_0
	if err := conn.ReadJSON(&hs); err != nil {
		log.Println(err)

		return
	}
	conn.SetReadDeadline(time.Time{})

	if hs.Version != protocolVersion2_0 {
		conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
		conn.WriteJSON("Incorrect Version")
		conn.Close()

		return
	}
	pass := hs.Pass

	log.Println(req.RemoteAddr, ": ", pass)

	mut.Lock()
	ch, ok := conns[pass]

	if !ok {
		ch := make(chan *websocket.Conn, 1)
		conns[pass] = ch

		go serverWorker2_0(ch, conn, func() {
			mut.Lock()
			if c, ok := conns[pass]; ok && c == ch {
				delete(conns, pass)

				close(ch)
			}
			mut.Unlock()
		})
	} else {
		delete(conns, pass)

		ch <- conn
	}

	mut.Unlock()
}
