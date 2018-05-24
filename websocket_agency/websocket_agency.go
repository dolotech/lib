package main

import (
	"io"
	"log"
	"net"
	"os"
	"fmt"
	"bufio"
	"bytes"
	"strconv"
	"strings"
	"crypto/sha1"
	"runtime"
	"encoding/base64"
	"encoding/binary"
	"net/http"
)

type WebSocket struct {
	Listener net.Listener
	Clients  []*Client
}

var num int
var new_addr string

type Client struct {
	Conn          net.Conn
	Nickname      string
	Shook         bool
	Server        *WebSocket
	Id            int
	TcpConn       net.Conn
	WebsocketType int
	Num           int
}

type Msg struct {
	Data string
	Num  int
}

func (self *Client) Release() {
	// release all connect
	self.TcpConn.Close()
	self.Conn.Close()
}

func (self *Client) Handle() {
	defer self.Release()
	if !self.Handshake() {
		// handshak err , del this conn
		return
	}

	// connect to another server for tcp
	if !self.ConnTcpServer() {
		// can not connect to the other server , release
		return
	}
	num = num + 1
	log.Print("now connect num : ", num)
	self.Num = num
	go self.Read()
	self.ReadTcp()
}

func (self *Client) Read() {
	var (
		buf     []byte
		err     error
		rsv     byte
		opcode  byte
		mask    byte
		mKey    []byte
		length  uint64
		l       uint16
		payload byte
	)
	for {
		buf = make([]byte, 2)
		_, err = io.ReadFull(self.Conn, buf)
		if err != nil {
			self.Release()
			break
		}
		//fin = buf[0] >> 7
		//if fin == 0 {
		//}

		rsv = (buf[0] >> 4) & 0x7
		// which must be 0
		if rsv != 0 {
			log.Print("Client send err msg:", rsv, ", disconnect it")
			self.Release()
			break
		}

		opcode = buf[0] & 0xf
		// opcode   if 8 then disconnect
		if opcode == 8 {
			log.Print("CLient want close Connection")
			self.Release()
			break
		}

		// should save the opcode
		// if client send by binary should return binary (especially for Egret)
		self.WebsocketType = int(opcode)

		mask = buf[1] >> 7
		// the translate may have mask

		payload = buf[1] & 0x7f
		// if length < 126 then payload mean the length
		// if length == 126 then the next 8bit mean the length
		// if length == 127 then the next 64bit mean the length
		switch {
		case payload < 126:
			length = uint64(payload)

		case payload == 126:
			buf = make([]byte, 2)
			io.ReadFull(self.Conn, buf)
			binary.Read(bytes.NewReader(buf), binary.BigEndian, &l)
			length = uint64(l)

		case payload == 127:
			buf = make([]byte, 8)
			io.ReadFull(self.Conn, buf)
			binary.Read(bytes.NewReader(buf), binary.BigEndian, &length)
		}
		if mask == 1 {
			mKey = make([]byte, 4)
			io.ReadFull(self.Conn, mKey)
		}
		buf = make([]byte, length)
		io.ReadFull(self.Conn, buf)
		if mask == 1 {
			for i, v := range buf {
				buf[i] = v ^ mKey[i%4]
			}
			//fmt.Print("mask", mKey)
		}
		log.Print("rec from the client(", self.Num, ")", string(buf))
		self.TcpConn.Write(buf)
	}
}

// read from other tcp
func (self *Client) ReadTcp() {
	var (
		buf []byte
	)
	buf = make([]byte, 1024)

	for {
		length, err := self.TcpConn.Read(buf)

		if err != nil {
			self.Release()
			num = num - 1
			// only
			log.Print("other tcp connect err", err)
			log.Print("disconnect client :", self.Num)
			log.Print("now have:", num)
			break
		}
		log.Print("recv from other tcp : ", string(buf[:length]))
		self.Write(buf[:length])
		//Write to websocket
	}
}

// write to websocket
func (self *Client) Write(data []byte) bool {
	data_binary := new(bytes.Buffer) //which

	//should be binary or string
	frame := []byte{129} //string
	length := len(data)
	// 10000001
	if self.WebsocketType == 2 {
		frame = []byte{130}
		// 10000010
		err := binary.Write(data_binary, binary.LittleEndian, data)
		if err != nil {
			log.Print(" translate to binary err", err)
		}
		length = len(data_binary.Bytes())
	}
	switch {
	case length < 126:
		frame = append(frame, byte(length))
	case length <= 0xffff:
		buf := make([]byte, 2)
		binary.BigEndian.PutUint16(buf, uint16(length))
		frame = append(frame, byte(126))
		frame = append(frame, buf...)
	case uint64(length) <= 0xffffffffffffffff:
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, uint64(length))
		frame = append(frame, byte(127))
		frame = append(frame, buf...)
	default:
		log.Print("Data too large")
		return false
	}
	if self.WebsocketType == 2 {
		frame = append(frame, data_binary.Bytes()...)
	} else {
		frame = append(frame, data...)
	}
	self.Conn.Write(frame)
	frame = []byte{0}
	return true
}

func (self *Client) ConnTcpServer() bool {

	conn, err := net.Dial("tcp", new_addr)

	if (err != nil) {
		log.Print("connect other tcp server error")
		return false
	}

	self.TcpConn = conn
	return true
}

func (self *Client) Handshake() bool {
	if self.Shook {
		return true
	}
	reader := bufio.NewReader(self.Conn)
	key := ""
	str := ""
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			log.Print("Handshake err:", err)
			return false
		}
		if len(line) == 0 {
			break
		}
		str = string(line)
		if strings.HasPrefix(str, "Sec-WebSocket-Key") {
			if len(line) >= 43 {
				key = str[19:43]
			}
		}
	}
	if key == "" {
		return false
	}
	sha := sha1.New()
	io.WriteString(sha, key+"258EAFA5-E914-47DA-95CA-C5AB0DC85B11")
	key = base64.StdEncoding.EncodeToString(sha.Sum(nil))
	header := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Version: 13\r\n" +
		"Sec-WebSocket-Accept: " + key + "\r\n" +
		"Upgrade: websocket\r\n\r\n"
	self.Conn.Write([]byte(header))
	self.Shook = true
	self.Server.Clients = append(self.Server.Clients, self)
	return true
}

func NewWebSocket(addr string) *WebSocket {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
		// if cannot listen then return
	}
	return &WebSocket{l, make([]*Client, 0)}
}

func (self *WebSocket) Loop() {
	for {
		conn, err := self.Listener.Accept()
		if err != nil {
			log.Print("client conn err:", err)
			continue
		}
		s := conn.RemoteAddr().String()
		i, _ := strconv.Atoi(strings.Split(s, ":")[1])
		client := &Client{conn, "", false, self, i, conn, 1, num}
		go client.Handle()
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	// show num of goroutine
	w.Header().Set("Content-Type", "text/plain")
	num := strconv.FormatInt(int64(runtime.NumGoroutine()), 10)
	w.Write([]byte(num))
}

func main() {
	arg_num := len(os.Args)
	if arg_num < 2 {
		fmt.Println(arg_num)
		fmt.Print("Wrong Arguments\nxxxx xxx.xxx.xxx.xxx:xxxx\nport ip:port(for tcp)")
		os.Exit(0)
	}
	num = 0
	conn, err := net.Dial("tcp", string(os.Args[2]))
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	new_addr = os.Args[2]
	fmt.Println("Check Arguments Ok")
	conn.Close()
	port := os.Args[1]
	ip_port := "0.0.0.0:" + string(port)
	ws := NewWebSocket(ip_port)
	// listen 9051
	go ws.Loop()
	fmt.Println("Start Listen")
	// listen 11181 to show num of goroutine
	http.HandleFunc("/", handler)
	http.ListenAndServe(":11181", nil)
}
