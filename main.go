package main

import (
	"os"
	"fmt"
	"strconv"
)

func main() {
	var from, to int
	var dest string
	var err error

	from, err = strconv.Atoi(os.Args[1])
	if err != nil {
		os.Exit(1)
	}
	to, err = strconv.Atoi(os.Args[2])
	if err != nil {
		os.Exit(1)
	}
	dest = os.Args[3]
	fmt.Printf("foward port from %d to %d, destination: %s\n", from, to, dest)
	for port := from; port <= to; port ++ {
		var local = fmt.Sprintf("0.0.0.0:%d", port)
		var remote = fmt.Sprintf("%s:%d", dest, port)
		NewServer(local, remote).Start()
	}
	select {}
}
