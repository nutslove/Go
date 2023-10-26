package main

import (
	"fmt"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}

	// for文でlistener.Accept()を呼び出すことで、複数のクライアントからのリクエストを受け付けることができる
	for {
		conn, err := listener.Accept() // Acceptは、クライアントからのリクエストを待ち受ける。クライアントからのリクエストがあると、そのリクエストを表すnet.Connを返す。
		if err != nil {
			panic(err)
		}

		// go handler(conn)
		go handler_keep_alive(conn)
	}
}

// func handler(conn net.Conn) {
// 	buf := make([]byte, 1024) // 1KB
// 	n, err := conn.Read(buf)  // n is the number of bytes read from request
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println(n)
// 	fmt.Println(string(buf[:n])) // http requestの中身を表(http header, http body)

// 	conn.Write([]byte("Hello! from Server"))

// 	conn.Close() // このconnは、1つのクライアントとの通信を表す。 このconnを閉じることで、クライアントとの通信を終了する。
// }

// １回のTCPコネクションで、複数のリクエストを受け付ける(keep-alive)
func handler_keep_alive(conn net.Conn) {
	buf := make([]byte, 1024) // 最大1KBまでのリクエストを受け付ける

	for {
		n, err := conn.Read(buf) // n is the number of bytes read from request
		if err != nil {
			panic(err)
		}
		fmt.Printf("受信バイト数: %d\n", n)
		responseData := string(buf[:n])
		fmt.Printf("受信データ: %s\n", responseData)
		if responseData == `"close"` { // json.Marshal()で、文字列を[]byteに変換すると、文字列の前後にダブルクォーテーションが付与されるため、``で囲んで""をエスケープする必要がある。
			fmt.Println("Close connection...")
			conn.Close()
			break
		}

		conn.Write([]byte("Hello! from Server"))
	}
}
