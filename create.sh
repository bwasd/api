#!/bin/bash

curl --header "Content-Type: application/json" \
	--request POST \
	--data '{"sku":"9a982384y932", "attrs":{"mass":"10g"}}' \
	http://localhost:8080/product/create

