package main

import (
	"encoding/json"

	"time"

	"log"

	"bytes"

	"github.com/gorilla/websocket"
)

const (
	userID   string = "verysecretID"
	username string = "[Bot] bobot"
	apiURL   string = "ws://botws.generals.io/socket.io/?EIO=3&transport=websocket"
)

type ApiEvent string

// Event string constants
const (
	// queue events
	setUsername   ApiEvent = "set_username"
	join1v1       ApiEvent = "join_1v1"
	joinPrivate   ApiEvent = "join_private"
	joinFFA       ApiEvent = "play"
	setCustomTeam ApiEvent = "set_custom_team"
	forceStart    ApiEvent = "set_force_start"
	leave         ApiEvent = "leave_game"

	// game state events
	StartEvent    ApiEvent = "game_start"
	LostEvent     ApiEvent = "game_lost"
	WonEvent      ApiEvent = "game_won"
	QueueEvent    ApiEvent = "queue_update"
	PreStartEvent ApiEvent = "pre_game_start"

	// gameplay
	attack      ApiEvent = "attack"
	UpdateEvent ApiEvent = "game_update"
)

// Client handles the websocket communication using SocketIO protocol
type Client struct {
	conn *websocket.Conn

	userID   string
	username string

	ended    bool
	sendChan chan []byte
	Handlers map[ApiEvent]func(json.RawMessage)
}

func NewClient() (*Client, error) {

	dialer := &websocket.Dialer{}
	dialer.EnableCompression = false

	c, _, err := dialer.Dial(apiURL, nil)
	if err != nil {
		return nil, err
	}

	return &Client{
			c,
			userID, username, false,
			make(chan []byte), map[ApiEvent]func(json.RawMessage){}},
		nil
}

func (c *Client) Connect() error {
	go func() {
		for range time.Tick(10 * time.Second) {
			if c.ended {
				return
			}
			c.sendChan <- []byte("2") // socket.io ping
		}
	}()

	// Ping the socket
	go func() {
		for data := range c.sendChan {
			err := c.conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Println("Connection failed")
				c.Stop()
				return
			}
		}
	}()

	// main loop reading from socket
	for !c.ended {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return err
		}

		decoder := json.NewDecoder(bytes.NewBuffer(msg))
		var msgType int
		decoder.Decode(&msgType)

		if msgType != 42 {
			continue
		}

		var raw json.RawMessage
		decoder.Decode(&raw)
		eventName := ""
		data := []interface{}{&eventName}

		json.Unmarshal(raw, &data)
		f, ok := c.Handlers[ApiEvent(eventName)]
		if ok {
			f(raw)
		}
	}
	return nil
}

func (c *Client) Stop() {
	c.conn.Close()
	c.ended = true
}

func (c *Client) sendMsg(data ...interface{}) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return
	}
	buf := []byte("42" + string(dataBytes)) // 42 is socketIO text message
	c.sendChan <- buf
}

func (c *Client) SetUsername() {
	c.sendMsg(setUsername, c.userID, c.username)
}

func (c *Client) JoinPrivate(gameID string) *Game {
	g := NewGame(c, gameID)
	c.sendMsg(joinPrivate, gameID, c.userID)
	return g
}

func (c *Client) Join1v1() *Game {
	g := NewGame(c, "dunno")
	c.sendMsg(join1v1, c.userID)
	return g
}

func (c *Client) ForceStart(gameID string) {
	c.sendMsg(forceStart, gameID, true)
}

func (c *Client) Attack(from, to int, half bool) {
	c.sendMsg(attack, from, to, half)
}
