package main

import (
	"encoding/json"
	"fmt"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080") // Dialは、net.Listen()と違い、クライアントからのリクエストを待ち受けるのではなく、クライアントからのリクエストを送信する。
	if err != nil {
		panic(err)
	}

	requestAndresponse(conn, "request from client")
	requestAndresponse(conn, "request from client2")
	requestAndresponse(conn, "close")
}

func requestAndresponse(conn net.Conn, requestData string) {
	requestByteData, err := json.Marshal(requestData) // json.Marshal() returns []byte
	if err != nil {
		panic(err)
	}

	conn.Write(requestByteData)

	reponseData := make([]byte, 1024)
	n, err := conn.Read(reponseData) // EOFは、サーバー側でconn.Close()を呼び出すことで、クライアント側に送信される。
	if err != nil {
		fmt.Println("err message: ", err)
		return
	}

	fmt.Println(string(reponseData[:n]))
}
