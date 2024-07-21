package main

import (
	"chat/network"
)

func main() {
	n := network.NewServer()
	n.StartServer()
}
