package network

import (
	"chat/types"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = &websocket.Upgrader{ReadBufferSize: types.SocketBufferSize, WriteBufferSize: types.MessageBufferSize, CheckOrigin: func(r *http.Request) bool { return true }}

type message struct {
	Name    string
	Message string
	Time    int64
}

type Room struct {
	Forward chan *message    // 受信されるメッセージを保管する値（入ってくるメッセージを他のクライアントたちに転送）
	Join    chan *client     // Socketが接続される場合に動作
	Leave   chan *client     // Socketが切断される場合に動作
	Clients map[*client]bool // 現在RoomにいるClientの情報を保存
}

type client struct {
	Send   chan *message
	Room   *Room
	Name   string
	Socket *websocket.Conn
}

func NewRoom() *Room {
	return &Room{
		Forward: make(chan *message),
		Join:    make(chan *client),
		Leave:   make(chan *client),
		Clients: make(map[*client]bool),
	}
}

func (c *client) Read() {
	defer c.Socket.Close()
	for {
		var msg *message
		err := c.Socket.ReadJSON(&msg)
		if err != nil {
			panic(err)
		} else {
			log.Println("READ: ", msg, "client: ", c.Name)
			log.Println()
			msg.Time = time.Now().Unix()
			msg.Name = c.Name

			c.Room.Forward <- msg
		}
	}
}

func (c *client) Write() {
	defer c.Socket.Close()
	for msg := range c.Send {
		log.Println("WRITE: ", msg, "client: ", c.Name)
		log.Println()
		err := c.Socket.WriteJSON(msg)
		if err != nil {
			panic(err)
		}
	}
}

func (r *Room) RunInit() {
	for {
		select {
		case client := <-r.Join:
			r.Clients[client] = true
		case client := <-r.Leave:
			r.Clients[client] = false
			delete(r.Clients, client)
			close(client.Send)
		case msg := <-r.Forward:
			for client := range r.Clients {
				client.Send <- msg
			}
		}
	}
}

func (r *Room) SocketServe(c *gin.Context) {
	socket, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		panic(err)
	}

	userCookie, err := c.Request.Cookie("auth")
	if err != nil {
		panic(err)
	}

	client := &client{
		Socket: socket,
		Send:   make(chan *message, types.MessageBufferSize),
		Room:   r,
		Name:   userCookie.Value,
	}

	r.Join <- client

	defer func() { r.Leave <- client }()

	go client.Write()

	client.Read()
}
