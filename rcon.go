package main

import (
	"errors"
	"fmt"
	"os"
)

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
	err = client.sendPacket(Packet{packetId: 0, packetType: serverdataAuth, packetBody: password})
	if err != nil {
		panic(err)
	}
	// Receive empty SERVERDATA_RESPONSE_VALUE
	{
		response, err := client.receivePacket()
		if err != nil {
			panic(err)
		}
		if (response != Packet{0, 0, ""}) {
			panic(errors.New("received unexpected packet"))
		}
	}
	// Receive authentication response SERVERDATA_AUTH_RESPONSE
	{
		response, err := client.receivePacket()
		if err != nil {
			panic(err)
		}
		if response.packetType != serverdataAuthResponse {
			panic(errors.New("received unexpected packet"))
		}
		if response.packetId != 0 {
			_, err := fmt.Fprintln(os.Stderr, "Failed to make connection: authentication failure")
			if err != nil {
				return nil
			}
			client.close()
			os.Exit(2)
		}
	}

	return &RCONConnection{client: client}
}

func (conn *RCONConnection) sendCommand(cmd string) string {
	// This method implements the trick, discovered by Koraktor and documented in the following link, to guarantee that
	// all meaningful responses have been received:
	// https://developer.valvesoftware.com/wiki/Source_RCON_Protocol#Multiple-packet_Responses

	// Send request packet
	var requestId = conn.counter()
	{
		packet := Packet{
			packetId:   requestId,
			packetType: serverdataExeccommand,
			packetBody: cmd,
		}
		err := conn.client.sendPacket(packet)
		if err != nil {
			panic(err)
		}
	}

	// Send ping packet
	// This packet will receive TWO responses: one identical (empty body), one more RESPONSE_VALUE with body 0x01 00
	var pingId = conn.counter()
	{
		packet := Packet{
			packetId:   pingId,
			packetType: serverdataResponseValue,
			packetBody: "",
		}
		err := conn.client.sendPacket(packet)
		if err != nil {
			panic(err)
		}
	}

	// Receive packet
	var respBody string = ""
	{
		var resp Packet
		{
			var err error
			resp, err = conn.client.receivePacket()
			if err != nil {
				panic(err)
			}
		}
		// Do this while the packet received has ID requestId
		{
			var err error
			for ; resp.packetId == requestId; resp, err = conn.client.receivePacket() {
				if err != nil {
					panic(err)
				}
				if resp.packetType != serverdataResponseValue {
					panic(errors.New("unexpected response type"))
				}
				respBody = respBody + resp.packetBody
			}
		}
	}
	// Receive ping back, then response
	{
		var resp Packet
		var err error
		// Receive ping and check for expectation
		resp, err = conn.client.receivePacket()
		if err != nil {
			panic(err)
		}
		if (resp != Packet{pingId, serverdataResponseValue, "\x00\x01\x00\x00"}) {
			msg := fmt.Sprintf("received unexpected response (ping); expected %v %v %v, got %v %v %v",
				pingId, serverdataResponseValue, "", resp.packetId, resp.packetType, resp.packetBody)
			panic(errors.New(msg))
		}
	}
	return respBody
}

func (conn *RCONConnection) counter() int {
	conn.idCounter++
	return conn.idCounter
}

func (conn *RCONConnection) close() {
	conn.client.close()
}
