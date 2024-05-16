package main

import (
	"bufio"
	"os"
	"time"
)

func main() {
	buffer := bufio.NewWriter(os.Stdout)
	buffer.WriteString("bufio.Writer \n")

	time.Sleep(time.Second * 2)
	buffer.Flush()
}
