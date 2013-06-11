package iproto

import (
	"bytes"
	"encoding/binary"
	"net"
)

type IProto struct {
	addr        string
	connection  *net.TCPConn
	requestID   int32
	requests    map[int32] chan *Response
	writeChan   chan *bytes.Buffer
}

type Response struct {
	requestType  int32
	bodyLength   int32
	requestID    int32
	responseBody *bytes.Buffer
}

func Connect(addr string) (connection *IProto, err error) {
	raddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return
	}

	conn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		return
	}
	connection = &IProto{ addr, conn, 0, make(map[int32] chan *Response), make(chan *bytes.Buffer) }

	go connection.read()
	go connection.write()

	return
}

func (conn *IProto) Request(requestType int32, body *bytes.Buffer) (response *Response, err error) {
	// Prepare packet
	packet, err := conn.pack(requestType, body)
	if err != nil {
		return
	}

	ch := make(chan *Response)
	// Store request
	conn.requests[conn.requestID] = ch

	// Send it
	conn.writeChan <- packet

	response = <- ch
	return
}

func (conn *IProto) pack(requestType int32, body *bytes.Buffer) (packet *bytes.Buffer, err error) {	
	// Each request should have uniq RequestID
	conn.requestID++
	// And it should not get out of 32 bits
	if conn.requestID >= 1<<31 - 1 {
		conn.requestID = 1
	}
	bodyLength := int32(body.Len())

	// <header> ::= <type><body_length><request_id>
	header := [] int32 { requestType, bodyLength, conn.requestID }

	// Put integers into packet
	packet = new(bytes.Buffer)
	err = binary.Write(packet, binary.LittleEndian, header)
	if err != nil {
		return
	}

	// Put body into packet
	_, err = packet.Write(body.Bytes())
	return
}

func (conn *IProto) read() {
	res := make([]int32, 3)
	for {
		headerBuf := make([]byte, 12)
		// Read header (12 bytes)
		_, err := conn.connection.Read(headerBuf)
		if err != nil {
			panic(err)
		}

		// Unpack data
		err = binary.Read(bytes.NewBuffer(headerBuf), binary.LittleEndian, &res)
		if err != nil {
			panic(err)
		}

		requestType := res[0]
		bodyLength  := res[1]
		requestID   := res[2]

		// Read body
		bodyRest := bodyLength
		bodyBuf  := make([]byte, bodyRest)
		if bodyRest > 0 {
			_, err = conn.connection.Read(bodyBuf)
			if err != nil {
				panic(err)
			}
		}

		response := &Response{ requestType, bodyLength, requestID, bytes.NewBuffer(bodyBuf) }

		conn.requests[requestID] <- response
	}
}

func (conn *IProto) write() {
	var (
		packet *bytes.Buffer
		err error
	)

	for {
		packet = <- conn.writeChan

		_, err = conn.connection.Write(packet.Bytes())
		if err != nil {
			panic(err)
		}
	}
}