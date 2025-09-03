package main

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

type ConnectionPool struct {
	cons []*Connection
}

func (c *ConnectionPool) SendEvent(event string) error {
	for _, v := range c.cons {
		err := v.SendEvent(event)

		// FIXME
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *ConnectionPool) SetConnection(connection *Connection) {
	c.cons = append(c.cons, connection)
}

type Connection struct {
	sessions []*Sink
}

func (c *Connection) SendEvent(event string) error {
	for _, v := range c.sessions {
		err := v.SendEvent(event)

		// FIXME
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Connection) Subscribe(s *Sink) {
	c.sessions = append(c.sessions, s)
}

type Sink struct {
	id     string
	socket *websocket.Conn

	// something like filter here
}

func (s *Sink) SendEvent(event string) error {
	content := []string{
		"EVENT",
		s.id,
		event,
	}
	bytes, err := json.Marshal(content)
	if err != nil {
		return err
	}
	// should consider about filter
	return s.socket.WriteMessage(websocket.TextMessage, bytes)
}

func (s *Sink) SendEOSE() error {
	content := []string{
		"EOSE",
		s.id,
	}
	bytes, err := json.Marshal(content)
	if err != nil {
		return err
	}
	return s.socket.WriteMessage(websocket.TextMessage, bytes)
}

var aggregator = ConnectionPool{}