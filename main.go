// Copyright 2011 Google Inc. All Rights Reserved.
// This file is available under the Apache license.

package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	_ "net/http/pprof"
)

var port *string = flag.String("port", "3903", "HTTP port to listen on.")
var logs *string = flag.String("logs", "", "List of files to monitor.")
var progs *string = flag.String("progs", "", "Dicrectory containing programs")

// Global metrics storage.
var metrics []*Metric

// CSV export
func handleCsv(w http.ResponseWriter, r *http.Request) {
	c := csv.NewWriter(w)
	for _, m := range metrics {
		record := []string{m.Name,
			fmt.Sprintf("%f", m.Value),
			fmt.Sprintf("%s", m.Time),
			fmt.Sprintf("%d", m.Type),
			m.Unit}
		for k, v := range m.Tags {
			record = append(record, fmt.Sprintf("%s=%s", k, v))
		}
		c.Write(record)
	}
	c.Flush()
}

type Foo struct {
	Name string
}

// JSON export
func handleJson(w http.ResponseWriter, r *http.Request) {
	b, err := json.Marshal(metrics)
	if err != nil {
		log.Println("error marshalling metrics into json:", err.Error())
	}
	w.Write(b)
}

// vms contains a list of virtual machines to execute when each new line is received
var vms []*vm

// RunVms receives a line from a channel and sends it to all VMs.
func RunVms(lines chan string) {
	for {
		select {
		case line := <-lines:
			for _, v := range vms {
				v.Run(line)
			}
		}
	}

}

func main() {
	flag.Parse()
	w := NewWatcher()
	t := NewTailer(w)

	fis, err := ioutil.ReadDir(*progs)
	if err != nil {
		log.Fatalf("Failure reading progs from %q: %s", *progs, err)
	}
	for _, fi := range fis {
		f, err := os.Open(fmt.Sprintf("%s/%s", *progs, fi.Name()))
		if err != nil {
			log.Printf("Failed to open %s: %s\n", fi.Name(), err)
			continue
		}
		defer f.Close()
		vm, errors := Compile(fi.Name(), f)
		if errors != nil {
			for _, e := range errors {
				log.Printf(e)
			}
			continue
		}
		vms = append(vms, vm)
	}

	go RunVms(t.Line)
	go t.start()
	go w.start()

	for _, pathname := range strings.Split(*logs, ",") {
		t.Tail(pathname)
	}

	http.HandleFunc("/json", handleJson)
	http.HandleFunc("/csv", handleCsv)
	log.Fatal(http.ListenAndServe(":"+*port, nil))

}