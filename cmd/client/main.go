package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/bwasd/api"
)

func main() {
	p, err := ioutil.ReadFile("products.json")
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
