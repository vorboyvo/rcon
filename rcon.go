package main

import "fmt"

/**
 * RCON Connection
 */

type RCONConnection struct {
	idCounter int
	client    *client
}

func NewRCONConnection(host string, port int, password string) *RCONConnection {
	// Checks for argument legality
	if host == "" {
		panic("cannot have empty host!")
	}
	if port < 1 || port > 65535 {
		panic("cannot have invalid port!")
	}

	client, err := newClient(host, port)
	if err != nil {
		panic(err)
	}

	// Authenticate RCON connection
	response, err := client.sendAndReceive(Packet{packetId: 0, packetType: serverdataAuth, packetBody: password})
	if err != nil {
		panic(err)
	}
	fmt.Println(response)
	return &RCONConnection{client: client}
}

func (conn *RCONConnection) close() {
	conn.client.close()
}
