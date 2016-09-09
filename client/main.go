package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

type event int

const (
	HandShake event = iota
	Message
)

type Record struct {
	Event     event
	TimeStamp time.Time
	Latency   float64
}

func client(cfg *websocket.Config, done chan bool, wg *sync.WaitGroup, output chan []Record) {
	history := make([]Record, 1024)
	ts := time.Now()
	conn, err := websocket.DialConfig(cfg)
	latency := time.Since(ts).Seconds()
	history = append(history, Record{Event: HandShake, TimeStamp: ts, Latency: latency})

	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case <-done:
			output <- history
			wg.Done()
			return
		default:
			ts := time.Now()
			_, err = conn.Write(bytes.Repeat([]byte("b"), 33))
			if err != nil {
				log.Fatal(err)
			}
			var msg = make([]byte, 512)
			if _, err = conn.Read(msg); err != nil {
				log.Fatal(err)
			}
			latency := time.Since(ts).Seconds()
			history = append(history, Record{Event: Message, TimeStamp: ts, Latency: latency})
			time.Sleep(time.Second)
		}
	}
}

func publish(history *[]Record) {
	b, _ := json.Marshal(history)
	os.Stdout.Write(b)
}

func main() {
	var maxclients int
	flag.IntVar(&maxclients, "maxclients", 10000, "max number of clients")

	var duration int
	flag.IntVar(&duration, "duration", 60, "test duration (seconds)")

	var host string
	flag.StringVar(&host, "host", "127.0.0.1", "")

	var port string
	flag.StringVar(&port, "port", "8000", "")

	flag.Parse()

	target := net.JoinHostPort(host, port)
	url := url.URL{Scheme: "ws", Host: target, Path: "/"}
	cfg, _ := websocket.NewConfig(url.String(), "http://localhost")

	done := make(chan bool)
	var wg = new(sync.WaitGroup)
	output := make(chan []Record, maxclients+1)

	for i := 0; i < maxclients; i++ {
		wg.Add(1)
		go client(cfg, done, wg, output)
		time.Sleep(10 * time.Millisecond)
		fmt.Printf("%d\n", i)
	}
	time.Sleep(time.Duration(duration) * time.Second)
	close(done)
	wg.Wait()
	ret := new([]Record)
	for {
		select {
		case v := <-output:
			*ret = append(*ret, v...)
		default:
			publish(ret)
			return
		}
	}
}
