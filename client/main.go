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

func client(cfg *websocket.Config, done chan bool, history chan Record) {
	ts := time.Now()
	conn, err := websocket.DialConfig(cfg)
	latency := time.Since(ts).Seconds()
	history <- Record{Event: HandShake, TimeStamp: ts, Latency: latency}

	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case <-done:
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
			history <- Record{Event: Message, TimeStamp: ts, Latency: latency}
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
	history := make(chan Record, 10240)
	fmt.Println("running")
	for i := 0; i < maxclients; i++ {
		go client(cfg, done, history)
		time.Sleep(10 * time.Millisecond)
		fmt.Printf("%d\n", i)
	}
	output := new([]Record)
	timer := time.After(time.Duration(duration) * time.Second)
	for {
		select {
		case <-timer:
			done <- true
			publish(output)
			return
		case v := <-history:
			*output = append(*output, v)
		}
	}
}
