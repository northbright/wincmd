package main

import (
	"log"

	"github.com/northbright/wincmd"
)

func main() {
	output, err := wincmd.Run("NET VIEW")
	log.Printf("NET VIEW output:\n%s, error: %v", output, err)
}
