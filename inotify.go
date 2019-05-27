package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type WatcherMux struct {
	Root      string
	UrlPrefix string
	watcher   *fsnotify.Watcher
	pubsub    PubSub
}

func NewWatcherMux(root, urlPrefix string) *WatcherMux {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	pubsub := PubSub{}
	wm := &WatcherMux{
		root,
		urlPrefix,
		watcher,
		pubsub,
	}
	go wm.Start()
	return wm
}

func (p *WatcherMux) Start() {
	log.Println("fsnotify start")
	for {
		select {
		case event, ok := <-p.watcher.Events:
			if !ok {
				return
			}
			log.Println("fsnotify event:", event)

			topic, err := filepath.Rel(p.Root, event.Name)
			MustNil(err)
			published := p.pubsub.Pub(topic, event, false)
			if !published {
				log.Println("fsnotify pub fail:", topic)
			}
		case err, ok := <-p.watcher.Errors:
			if !ok {
				return
			}
			log.Println("fsnotify error:", err)
		}
	}
}

func (p *WatcherMux) Close() {
	p.watcher.Close()
}

func (p *WatcherMux) Watch(path string) {
	if !ExistFile(path) {
		log.Printf("file does not exist %v", path)
	}
	err := p.watcher.Add(path)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("watched file %v", path)
}

func (p *WatcherMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("file as websocket:", r.URL)

	// par url path and file path
	pathAsTopic, err := filepath.Rel(p.UrlPrefix, r.URL.Path)
	PanicErr(err)
	pathFile := filepath.Join(p.Root, pathAsTopic)
	if !ExistFile(pathFile) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("not found %s", pathAsTopic)))
		return
	}

	// upgrade to websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		msg := fmt.Sprintf("unknow error when upgrade protocol: %s", err)
		w.Write([]byte(msg))
		return
	}

	// watch file
	p.Watch(pathFile)

	done := make(chan bool)
	fn := func(msg interface{}) error {
		log.Println(msg)
		conn.WriteMessage(websocket.TextMessage, []byte(msg.(fsnotify.Event).Name))
		return nil
	}
	subKey := Sha256(fmt.Sprintf("%v", conn))
	t := p.pubsub.Subscribe(pathAsTopic, subKey, fn)

	closeCleanup := func() {
		t.Unsubscribe(subKey)
		if len(t.Subs) == 0 {
			if err := p.watcher.Remove(pathFile); err != nil {
				log.Printf("failed remove watching file %s: %v", pathAsTopic, err)
			} else {
				log.Printf("removed watching file %s", pathAsTopic)
			}
		}
	}
	conn.SetCloseHandler(func(code int, text string) error {
		closeCleanup()
		return nil
	})

	// keep connection and controled by done
	for {
		var data map[string]interface{}
		select {
		case <-done:
			data = map[string]interface{}{
				"ok":  false,
				"msg": "closed by server",
			}
		default:
			messageType, clientMsg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			switch messageType {
			case websocket.TextMessage:
				data = map[string]interface{}{
					"ok":  false,
					"msg": string(clientMsg),
				}
			case websocket.BinaryMessage:
				data = map[string]interface{}{
					"ok":  false,
					"msg": "binary message is not supported",
				}
			}
		}
		// send back
		if err := conn.WriteMessage(websocket.TextMessage, MarshalMust(data)); err != nil {
			log.Println(err)
			return
		}
	}
}
