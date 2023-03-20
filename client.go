package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
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
 * As defined here: https://developer.valvesoftware.com/wiki/Source_RCON_Protocol
 */

type packet struct {
	packetId   int
	packetType int
	packetBody string
}

func (p packet) size() int {
	var packetSize = 0                  // Do not count the size field
	packetSize += 4                     // 4 bytes for ID field
	packetSize += 4                     // 4 bytes for Type field
	packetSize += len(p.packetBody) + 1 // Length of Body field, plus 1 for null terminator
	return packetSize + 1               // 1 byte for null terminator
}

func (p packet) serializePacket() []byte {
	body := append([]byte(p.packetBody), 0) // Zero-terminate body string
	size := p.size()                        // Avoid repeat calls
	// Construct data slice
	bytes := make([]byte, size+4)
	binary.LittleEndian.PutUint32(bytes[0:4], uint32(size))
	binary.LittleEndian.PutUint32(bytes[4:8], uint32(p.packetId))
	binary.LittleEndian.PutUint32(bytes[8:12], uint32(p.packetType))
	copy(bytes[12:], body)
	return bytes
}

func deserializePacket(bytes []byte) (packet, error) {
	// Handle data too short
	if len(bytes) < 10 {
		return packet{}, errors.New("invalid data - too short")
	}
	// Handle data or body not zero terminated
	if bytes[len(bytes)-1] != 0 {
		return packet{}, errors.New("invalid data - not zero-terminated")
	}
	if bytes[len(bytes)-2] != 0 {
		return packet{}, errors.New("invalid data - body not zero-terminated")
	}
	// Read ID
	var packetId int
	{
		packetId32 := binary.LittleEndian.Uint32(bytes[0:4])
		packetId = int(packetId32)
	}
	// Read type
	var packetType int
	{
		packetType32 := binary.LittleEndian.Uint32(bytes[4:8])
		packetType = int(packetType32)
	}
	var packetBody = string(bytes[8 : len(bytes)-2]) // -2 for null terminators on body and whole packet
	//fmt.Printf("Packet type: %v, Packet ID: %v, Packet Body: '%v'\n", packetType, packetId, packetBody)
	return packet{packetId: packetId, packetType: packetType, packetBody: packetBody}, nil
}

/**
 * TCP client
 */

type client struct {
	con *net.Conn
}

func newClient(host string, port int) (*client, error) {
	con, err := net.DialTimeout("tcp", host+":"+strconv.Itoa(port), 4*time.Second)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Failed to make connection:", err)
		os.Exit(3)
	}
	return &client{
		&con,
	}, nil
}

func (c *client) sendPacket(p packet) error {
	bytes := p.serializePacket()
	// Check format
	if bytes[len(bytes)-2] != 0 {
		return errors.New("request body not a null terminated string")
	} else if bytes[len(bytes)-1] != 0 {
		return errors.New("request not null terminated")
	}
	// Send packet
	{
		num, err := (*c.con).Write(bytes)
		if err != nil {
			return err
		}
		if num != len(bytes) {
			return errors.New("failed to send full packet")
		}
	}
	return nil
}

func (c *client) receivePacket() (packet, error) {
	// Read response
	var response packet
	// read size
	var size int
	{
		buf := make([]byte, 4)
		num, err := io.ReadFull(*c.con, buf)
		if err != nil {
			return packet{}, err
		} else if num < 4 {
			return packet{}, errors.New("failed to read size of packet; could not read first word")
		}
		size32 := binary.LittleEndian.Uint32(buf)
		if num == 0 {
			return packet{}, errors.New("failed to read size of packet; buffer too small")
		} else if num < 0 {
			return packet{}, errors.New("failed to read size of packet; buffer too small")
		}
		size = int(size32)
	}
	// read size number of bytes
	{
		var err error
		buf := make([]byte, size)
		num, err := io.ReadFull(*c.con, buf)
		if err != nil {
			return packet{}, err
		} else if num < size {
			return packet{}, errors.New("failed to read packet up to full size")
		}
		response, err = deserializePacket(buf)
		if err != nil {
			return packet{}, err
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
