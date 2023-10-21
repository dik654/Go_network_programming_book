package main

import (
	"flag"
	"log"
)

var (
	addr  = flag.String("listen", "127.0.0.1:8080", "listen address")
	cert  = flag.String("cert", "", "certificate")
	pkey  = flag.String("key", "", "private key")
	files = flag.String("files", "./files", "static file directory")
)

func main() {
	flag.Parse()

	err := run(*addr, *files, *cert, *pkey)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Server gracefully shutdown")
}
