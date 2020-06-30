package client

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/gorilla/websocket"

	"quorumengineering/quorum-report/log"
)

type message struct {
	Version string          `json:"jsonrpc,omitempty"`
	ID      string          `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Error   *msgError       `json:"error,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
}

type msgError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type subMessage struct {
	ID     string          `json:"subscription"`
	Result json.RawMessage `json:"result"`
}

func (err *msgError) Error() string {
	if err.Message == "" {
		return fmt.Sprintf("error code: %v", err.Code)
	}
	return err.Message
}

type webSocketClient struct {
	rawUrl                string
	conn                  *websocket.Conn
	connWriteMux          sync.Mutex
	idCounter             uint32
	chainHeadSubscribedId string
	chainHeadChan         chan<- *ethTypes.Header
	rpcPendingResp        map[string]chan<- *message
	rpcMux                sync.RWMutex
}

func newWebSocketClient(rawUrl string) (*webSocketClient, error) {
	client := &webSocketClient{
		rawUrl:         rawUrl,
		idCounter:      0,
		rpcPendingResp: make(map[string]chan<- *message),
	}
	if err := client.dial(rawUrl); err != nil {
		return nil, err
	}
	return client, nil
}

func (c *webSocketClient) dial(rawUrl string) error {
	conn, _, err := websocket.DefaultDialer.Dial(rawUrl, nil)
	if err != nil {
		log.Error("Dial websocket endpoint error", "error", err)
		return err
	}
	c.conn = conn
	return nil
}

// subscribe header
func (c *webSocketClient) subscribeChainHead(ch chan<- *ethTypes.Header) error {
	c.chainHeadChan = ch
	c.chainHeadSubscribedId = c.nextID()

	params, _ := json.Marshal([]interface{}{"newHeads"})

	msg := &message{
		Version: "2.0",
		ID:      c.chainHeadSubscribedId,
		Method:  "eth_subscribe",
		Params:  params,
	}

	log.Debug("Send subscribe chain head message", "msg", msg)

	c.connWriteMux.Lock()
	defer c.connWriteMux.Unlock()

	// send subscription message
	if err := c.conn.WriteJSON(msg); err != nil {
		log.Error("Subscribe chain head error", "error", err)
		return err
	}
	return nil
}

// send rpc call
func (c *webSocketClient) sendRPCMsg(ch chan<- *message, method string, args ...interface{}) error {
	msg := &message{
		Version: "2.0",
		ID:      c.nextID(),
		Method:  method,
	}
	// marshal args to params
	if args != nil {
		params, err := json.Marshal(args)
		if err != nil {
			return err
		}
		msg.Params = params
	}

	c.setPendingRPC(msg.ID, ch)
	log.Debug("Send JSON RPC message", "msg.Method", msg.Method, "args", args, "msg.ID", msg.ID)

	c.connWriteMux.Lock()
	defer c.connWriteMux.Unlock()

	if err := c.conn.WriteJSON(msg); err != nil {
		log.Error("Write JSON RPC message error", "error", err, "msg", msg)
		return err
	}
	return nil
}

// listen and handle message
func (c *webSocketClient) listen() {
	if c.conn == nil {
		err := c.dial(c.rawUrl)
		if err != nil {
			log.Error("Dialing failed")
			return
		}
	}
	defer c.conn.Close()
	// TODO: send in a stop channel to break the loop
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if err == io.EOF {
				log.Warn("WebSocket connection closed")
			} else {
				log.Error("WebSocket read message error", "error", err)
			}
			return
		}

		log.Debug("WebSocket message received", "msg", string(msg))
		var receivedMsg message
		if err = json.Unmarshal(msg, &receivedMsg); err != nil {
			log.Error("Decode message error", "error", err)
			continue
		}

		if ch := c.getPendingRPC(receivedMsg.ID); ch != nil {
			// handle rpc message
			ch <- &receivedMsg
		} else if receivedMsg.ID == c.chainHeadSubscribedId {
			// handle subscription
			c.chainHeadSubscribedId = strings.Trim(string(receivedMsg.Result), "\"")
		} else if receivedMsg.Method == "eth_subscription" {
			// handle chain head message
			var subMsg subMessage
			if err = json.Unmarshal(receivedMsg.Params, &subMsg); err != nil {
				log.Error("Decode subscription message error", "error", err)
				continue
			}
			if subMsg.ID == c.chainHeadSubscribedId {
				var chainHead *ethTypes.Header
				if err = json.Unmarshal(subMsg.Result, &chainHead); err != nil {
					log.Error("Decode chain head error", "error", err)
					continue
				}
				c.chainHeadChan <- chainHead
			} else {
				// discard unknown message
				log.Warn("Unknown subscription message")
			}
		} else {
			// discard unknown message
			log.Warn("Unknown message")
		}
	}
}

// rpc pending message map update
func (c *webSocketClient) setPendingRPC(id string, ch chan<- *message) {
	c.rpcMux.Lock()
	defer c.rpcMux.Unlock()
	c.rpcPendingResp[id] = ch
}

// get rpc channel is used one time only
func (c *webSocketClient) getPendingRPC(id string) chan<- *message {
	c.rpcMux.Lock()
	defer c.rpcMux.Unlock()
	if ch, ok := c.rpcPendingResp[id]; ok {
		delete(c.rpcPendingResp, id)
		return ch
	}
	return nil
}

func (c *webSocketClient) nextID() string {
	return strconv.Itoa(int(atomic.AddUint32(&c.idCounter, 1)))
}
