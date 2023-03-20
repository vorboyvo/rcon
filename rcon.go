package main

import (
	"errors"
	"fmt"
)

/**
 * RCON Connection
 */

// An RCONConnection provides an interface with which you can connect to a remote server using the Source RCON protocol
// documented in the following link. It should not be initialized directly; behaviour is undefined if it is. Its
// methods use the protocol to authenticate and connect with servers.
type RCONConnection struct {
	idCounter int
	client    *client
}

// NewRCONConnection authenticates with the provided server given details, and returns a pointer to an RCONConnection
// for successful connection. Returns error on
func NewRCONConnection(host string, port int, password string) (*RCONConnection, error) {
	// Checks for argument legality
	if host == "" {
		return nil, errors.New("cannot have empty hostname")
	}
	if port < 1 || port > 65535 {
		return nil, errors.New("cannot have invalid port; must be between 1 and 65535, inclusive")
	}

	client, err := newClient(host, port)
	if err != nil {
		return nil, err
	}

	// Authenticate RCON connection
	err = client.sendPacket(packet{packetId: 0, packetType: serverdataAuth, packetBody: password})
	if err != nil {
		return nil, err
	}
	// Receive empty SERVERDATA_RESPONSE_VALUE
	{
		response, err := client.receivePacket()
		if err != nil {
			return nil, err
		}
		if (response != packet{0, 0, ""}) {
			msg := fmt.Sprintf("received unexpected packet (auth ping); expected %v %v %v, got %v %v %v",
				0, serverdataResponseValue, "", response.packetId, response.packetType, response.packetBody)
			return nil, errors.New(msg)
		}
	}
	// Receive authentication response SERVERDATA_AUTH_RESPONSE
	{
		response, err := client.receivePacket()
		if err != nil {
			return nil, err
		}
		if response.packetType != serverdataAuthResponse {
			msg := fmt.Sprintf("received unexpected packet (auth response); expected type %v, received %v",
				serverdataAuthResponse, response.packetType)
			return nil, errors.New(msg)
		}
		if response.packetId != 0 {
			return nil, new(AuthenticationFailure)
		}
	}

	return &RCONConnection{client: client}, nil
}

func (conn *RCONConnection) SendCommand(cmd string) (string, error) {
	// This method implements the trick, discovered by Koraktor and documented in the following link, to guarantee that
	// all meaningful responses have been received:
	// https://developer.valvesoftware.com/wiki/Source_RCON_Protocol#Multiple-packet_Responses

	// Send request packet
	var requestId = conn.counter()
	{
		packet := packet{
			packetId:   requestId,
			packetType: serverdataExeccommand,
			packetBody: cmd,
		}
		err := conn.client.sendPacket(packet)
		if err != nil {
			return "", err
		}
	}

	// Send ping packet
	// This packet will receive TWO responses: one identical (empty body), one more RESPONSE_VALUE with body 0x01 00
	var pingId = conn.counter()
	{
		packet := packet{
			packetId:   pingId,
			packetType: serverdataResponseValue,
			packetBody: "",
		}
		err := conn.client.sendPacket(packet)
		if err != nil {
			return "", err
		}
	}

	// Receive packet
	var respBody = ""
	{
		var resp packet
		{
			var err error
			resp, err = conn.client.receivePacket()
			if err != nil {
				return "", err
			}
		}
		// Do this while the packet received has ID requestId
		{
			var err error
			for ; resp.packetId == requestId; resp, err = conn.client.receivePacket() {
				if err != nil {
					return "", err
				}
				if resp.packetType != serverdataResponseValue {
					return "", errors.New("unexpected response type")
				}
				respBody = respBody + resp.packetBody
			}
		}
	}
	// Receive ping back, then response
	{
		var resp packet
		var err error
		// Receive ping and check for expectation
		resp, err = conn.client.receivePacket()
		if err != nil {
			return "", err
		}
		if (resp != packet{pingId, serverdataResponseValue, "\x00\x01\x00\x00"}) {
			msg := fmt.Sprintf("received unexpected response (ping); expected %v %v %v, got %v %v %v",
				pingId, serverdataResponseValue, "", resp.packetId, resp.packetType, resp.packetBody)
			return "", errors.New(msg)
		}
	}
	return respBody, nil
}

func (conn *RCONConnection) counter() int {
	conn.idCounter++
	return conn.idCounter
}

func (conn *RCONConnection) Close() {
	if conn.client == nil {
		return
	}
	conn.client.close()
}

// AuthenticationFailure is an error type which indicates that an RCONConnection failed to authenticate. It is intended
// to be caught and handled as a recoverable error.
type AuthenticationFailure struct{}

func (e AuthenticationFailure) Error() string {
	return "Failed to make connection: authentication failure"
}
