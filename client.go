package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"time"
)

//goland:noinspection SpellCheckingInspection
const (
	serverdataAuth          = 3
	serverdataAuthResponse  = 2
	serverdataExeccommand   = 2
	serverdataResponseValue = 0
)

/**
 * RCON TCP Packet
 */

type Packet struct {
	packetId   int
	packetType int
	packetBody string
}

func (p Packet) size() int {
	var packetSize = 0                  // Do not count the size field
	packetSize += 4                     // 4 bytes for ID field
	packetSize += 4                     // 4 bytes for Type field
	packetSize += len(p.packetBody) + 1 // Length of Body field, plus 1 for null terminator
	return packetSize + 1               // 1 byte for null terminator
}

func (p Packet) serializePacket() []byte {
	body := append([]byte(p.packetBody), 0) // Zero-terminate body string
	size := p.size()                        // Avoid repeat calls
	// Construct data slice
	bytes := make([]byte, size+4)
	binary.PutUvarint(bytes[0:4], uint64(size))
	binary.PutUvarint(bytes[4:8], uint64(p.packetId))
	binary.PutUvarint(bytes[8:12], uint64(p.packetType))
	copy(bytes[12:], body)
	return bytes
}

func deserializePacket(bytes []byte) (Packet, error) {
	// Handle data too short
	if len(bytes) < 10 {
		return Packet{}, errors.New("invalid data - too short")
	}
	// Handle data or body not zero terminated
	if bytes[len(bytes)-1] != 0 {
		return Packet{}, errors.New("invalid data - not zero-terminated")
	}
	if bytes[len(bytes)-2] != 0 {
		return Packet{}, errors.New("invalid data - body not zero-terminated")
	}
	// Read ID
	var packetId int
	{
		packetId64, num := binary.Uvarint(bytes[0:4])
		if num > 4 {
			return Packet{}, errors.New("failed to read id while deserializing data")
		}
		packetId = int(packetId64)
	}
	// Read type
	var packetType int
	{
		packetType64, num := binary.Uvarint(bytes[4:8])
		if num > 4 {
			return Packet{}, errors.New("failed to read type while deserializing data")
		}
		packetType = int(packetType64)
	}
	var packetBody = string(bytes[8 : len(bytes)-2]) // -2 for null terminators on body and whole packet
	fmt.Printf("Packet type: %v, Packet ID: %v, Packet Body: %v\n", packetType, packetId, packetBody)
	return Packet{packetId: packetId, packetType: packetType, packetBody: packetBody}, nil
}

/**
 * TCP client
 */

type client struct {
	con *net.Conn
}

func newClient(host string, port int) (*client, error) {
	con, err := net.DialTimeout("tcp", host+":"+strconv.Itoa(port), 10*time.Second)
	if err != nil {
		log.Println("Failed to make connection")
		log.Fatalln(err)
	}
	return &client{
		&con,
	}, nil
}

func (c *client) sendAndReceive(p Packet) (Packet, error) {
	bytes := p.serializePacket()
	// Check format
	if bytes[len(bytes)-2] != 0 {
		return Packet{}, errors.New("request body not a null terminated string")
	} else if bytes[len(bytes)-1] != 0 {
		return Packet{}, errors.New("request not null terminated")
	}
	// Send packet
	{
		num, err := (*c.con).Write(bytes)
		if err != nil {
			return Packet{}, err
		}
		if num != len(bytes) {
			return Packet{}, errors.New("failed to send full packet")
		}
	}
	// Read response
	var response Packet
	for response.packetType == 0 {
		// read size
		var size int
		{
			buf := make([]byte, 4)
			num, err := io.ReadFull(*c.con, buf)
			if err != nil {
				return Packet{}, err
			} else if num < 4 {
				return Packet{}, errors.New("failed to read size of packet; could not read first word")
			}
			fmt.Println(buf[:num])
			size64, num := binary.Uvarint(buf)
			if num == 0 {
				return Packet{}, errors.New("failed to read size of packet; buffer too small")
			} else if num < 0 {
				return Packet{}, errors.New("failed to read size of packet; buffer too small")
			}
			size = int(size64)
		}
		// read size number of bytes
		{
			var err error
			buf := make([]byte, size)
			num, err := io.ReadFull(*c.con, buf)
			if err != nil {
				return Packet{}, err
			} else if num < size {
				return Packet{}, errors.New("failed to read packet up to full size")
			}
			response, err = deserializePacket(buf)
			if err != nil {
				return Packet{}, err
			}
			fmt.Println(response)
		}
	}
	return response, nil
}

func (c *client) close() {
	err := (*c.con).Close()
	if err != nil {
		return
	}
}
