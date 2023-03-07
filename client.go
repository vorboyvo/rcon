package main

import (
	"encoding/binary"
	"errors"
	"net"
	"strconv"
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
	binary.PutVarint(bytes[0:4], int64(size))
	binary.PutVarint(bytes[4:8], int64(p.packetId))
	binary.PutVarint(bytes[8:12], int64(p.packetType))
	copy(bytes[12:], body)
	return bytes
}

func deserializePacket(bytes []byte) (Packet, error) {
	// Handle data too short
	if len(bytes) < 14 {
		return Packet{}, errors.New("invalid data - too short")
	}
	// Handle data or body not zero terminated
	if bytes[len(bytes)-1] != 0 {
		return Packet{}, errors.New("invalid data - not zero-terminated")
	}
	if bytes[len(bytes)-2] != 0 {
		return Packet{}, errors.New("invalid data - body not zero-terminated")
	}
	// Read size, handle size mismatch
	var size int
	{
		size64, num := binary.Varint(bytes[0:4])
		if num > 4 {
			return Packet{}, errors.New("failed to read size while deserializing data, size received is " + strconv.Itoa(num))
		}
		size = int(size64)
		if size != len(bytes)-4 {
			return Packet{}, errors.New("size in data does not match length of data")
		}
	}
	// Read ID
	var packetId int
	{
		packetId64, num := binary.Varint(bytes[4:8])
		if num > 4 {
			return Packet{}, errors.New("failed to read id while deserializing data")
		}
		packetId = int(packetId64)
	}
	// Read type
	var packetType int
	{
		packetType64, num := binary.Varint(bytes[8:12])
		if num > 4 {
			return Packet{}, errors.New("failed to read type while deserializing data")
		}
		packetType = int(packetType64)
	}
	// Read body
	var packetBody = string(bytes[12 : len(bytes)-2])
	return Packet{packetId: packetId, packetType: packetType, packetBody: packetBody}, nil
}

/**
 * TCP client
 */

type client struct {
	con *net.Conn
}

func newClient(host string, port int) (*client, error) {
	con, err := net.Dial("tcp", host+":"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}

	return &client{
		&con,
	}, nil
}

func (c *client) sendAndReceive(p Packet) (string, error) {
	bytes := p.serializePacket()
	// Check format
	if bytes[len(bytes)-2] != 0 {
		return "", errors.New("request body not a null terminated string")
	} else if bytes[len(bytes)-1] != 0 {
		return "", errors.New("request not null terminated")
	}
	// Send packet
	{
		num, err := (*c.con).Write(bytes)
		if err != nil {
			return "", err
		}
		if num != len(bytes) {
			return "", errors.New("failed to send full packet")
		}
	}
	// Read response
	var respString string
	{
		buf := make([]byte, 4096)
		num, err := (*c.con).Read(buf)
		if err != nil {
			return "", err
		}
		respString = string(buf[:num])
	}
	return respString, nil
}

func (c *client) close() {
	err := (*c.con).Close()
	if err != nil {
		return
	}
}
