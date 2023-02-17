package chat

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
)

type client struct {
	name  string
	conn  net.Conn
	in    chan []byte
	out   chan []byte
	token chan struct{}
}

type chat struct {
	clients      map[*client]bool
	messageQueue chan []byte
	exitQueue    chan *client
}

func NewChat() *chat {
	return &chat{
		clients:      make(map[*client]bool),
		messageQueue: make(chan []byte),
		exitQueue:    make(chan *client),
	}
}

func (c *chat) Start() {
	c.buildServer()
}

func (c *chat) buildServer() {
	server, err := net.Listen("tcp", ":8080") // localhost:8080

	if err != nil {
		log.Fatalf("could not start chat: %v", err)
	}
	go c.serve()
	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatalf("connection err: %v", err)
			continue
		}
		go c.handleConnection(conn) // start new goroutine per connection
	}
}

func (c *chat) exitGuide(client *client) {
	delete(c.clients, client)
	close(client.in)
	close(client.out)

	c.messageQueue <- []byte(fmt.Sprintf("%s left the chat\n", client.name))
	defer client.conn.Close()
}

func (c *chat) serve() {
	for {
		select {
		case msg := <-c.messageQueue:
			for client := range c.clients {
				client.out <- msg
			}
		case client := <-c.exitQueue:
			go c.exitGuide(client)
		}

	}
}

func (c *chat) handleConnection(conn net.Conn) {
	in, out := make(chan []byte), make(chan []byte)
	client := newClient(conn, in, out)
	c.clients[client] = true
	client.start()
	go c.bookRoom(client)
	<-client.token
	c.exitQueue <- client
}

func newClient(conn net.Conn, in chan []byte, out chan []byte) *client {
	return &client{in: in, out: out, conn: conn, token: make(chan struct{})}
}

func (c *chat) bookRoom(client *client) {
	for msg := range client.in {
		c.messageQueue <- []byte(fmt.Sprintf("%s says: %s\n", client.name, string(msg)))
	}
}

func (cl *client) start() {
	cl.setUpUsername()
	go cl.receiveMessages()
	go cl.sendMessages()
}

func (cl *client) receiveMessages() {
	scanner := bufio.NewScanner(cl.conn)
	for scanner.Scan() {
		if scanner.Err() != nil {
			log.Fatalf("receive messages %v: ", scanner.Err())
			break
		}
		cl.in <- scanner.Bytes()
	}
	cl.token <- struct{}{} // end connection

}

func (cl *client) sendMessages() {
	for msg := range cl.out {

		cl.conn.Write(msg)
	}
}

func (cl *client) setUpUsername() {
	io.WriteString(cl.conn, "Enter your username: ")
	scann := bufio.NewScanner(cl.conn)
	scann.Scan()
	cl.name = scann.Text()
	io.WriteString(cl.conn, fmt.Sprintf("welcome %s\n", cl.name))
}
