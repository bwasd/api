package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bwasd/api"
)

func post(file string) {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	file = filepath.Join(dir, file)
	p, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	for n, line := range bytes.Split(p, []byte{'\n'}) {
		var product api.Product
		// check if the JSON is well-formed before attempting to POST it
		if err := json.Unmarshal(line, &product); err != nil {
			fmt.Errorf("ignoring malformed JSON on line: %d\n", n)
			continue
		}
		jsonReq, err := json.Marshal(product)
		fmt.Printf("%s\n", jsonReq)
		_, err = http.Post("http://localhost:8080/product/create", "application/json; charset=utf-8", bytes.NewBuffer(jsonReq))
		if err != nil {
			fmt.Errorf("API error: %s", err)
		}

	}
}

var (
	flagList = flag.Bool("list", false, "list")
	flagFile = flag.String("file", "", "the file")
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: client [-list] [-post] [file...]\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	log.SetPrefix("client: ")
	log.SetFlags(0)
	flag.Parse()

	if *flagList {
		if flag.NArg() > 0 {
			usage()
		}
		// list()
		return
	}
	if len(flag.Args()) > 1 {
		fmt.Fprintf(os.Stderr, "too many arguments\n")
		usage()
	}
	post(*flagFile)
}
