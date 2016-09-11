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
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

type event int

const (
	HandShake event = iota
	Message
	Error
)

type Record struct {
	Event     event
	TimeStamp int64
	Latency   float64
	Tid       int
}

func client(idx int, cfg *websocket.Config, done chan bool, wg *sync.WaitGroup, output chan []Record, interval time.Duration) {
	defer wg.Done()
	history := make([]Record, 0, 1024)
	ts := time.Now()
	conn, err := websocket.DialConfig(cfg)

	if err != nil {
		log.Printf("Error dialing: %v\n", err)
		history = append(history, Record{Event: Error, TimeStamp: ts.UnixNano(), Latency: 0})
		output <- history
		return
	}
	latency := time.Since(ts).Seconds()
	fmt.Printf("%d: Connected\n", idx)
	history = append(history, Record{Event: HandShake, TimeStamp: ts.UnixNano(), Latency: latency})

loop:
	for {
		select {
		case <-done:
			break loop
		default:
			ts := time.Now()
			_, err = conn.Write(bytes.Repeat([]byte(" "), 33))
			if err != nil {
				log.Printf("Error writing: %v\n", err)
				history = append(history, Record{Event: Error, TimeStamp: ts.UnixNano(), Latency: 0})
				break loop
			}
			var msg = make([]byte, 512)
			var respLength int
			respLength, err = conn.Read(msg)
			if err != nil {
				log.Printf("Error reading: %v\n", err)
				history = append(history, Record{Event: Error, TimeStamp: ts.UnixNano(), Latency: 0})
				break loop
			}
			tid, err := strconv.Atoi(strings.TrimSpace(string(msg[:respLength])))
			if err != nil {
				tid = 0
			}
			latency := time.Since(ts).Seconds()
			history = append(history, Record{Event: Message, TimeStamp: ts.UnixNano(), Latency: latency, Tid: tid})
			time.Sleep(interval)
		}
	}
	output <- history
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

	var i float64
	flag.Float64Var(&i, "interval", 1., "sleep interval (seconds)")

	flag.Parse()

	target := net.JoinHostPort(host, port)
	url := url.URL{Scheme: "ws", Host: target, Path: "/"}
	cfg, _ := websocket.NewConfig(url.String(), "http://localhost")
	interval := time.Duration(int64(i*1000)) * time.Millisecond

	done := make(chan bool)
	var wg = new(sync.WaitGroup)
	output := make(chan []Record, maxclients+1)

	for i := 0; i < maxclients; i++ {
		wg.Add(1)
		go client(i, cfg, done, wg, output, interval)
		time.Sleep(10 * time.Millisecond)
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
