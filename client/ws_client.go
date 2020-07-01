package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

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
	rawUrl                      string
	conn                        *websocket.Conn
	connMux                     sync.Mutex
	connWriteMux                sync.Mutex
	idCounter                   uint32
	chainHeadSubscriptionId     string
	chainHeadSubscriptionCallId string
	chainHeadChan               chan<- *ethTypes.Header
	rpcPendingResp              map[string]chan<- *message
	rpcMux                      sync.RWMutex
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
	c.connMux.Lock()
	defer c.connMux.Unlock()

	conn, _, err := websocket.DefaultDialer.Dial(rawUrl, nil)
	if err != nil {
		log.Error("Dial WebSocket endpoint error", "error", err)
		return err
	}
	log.Info("Dial to WebSocket endpoint success", "rawUrl", rawUrl)
	c.conn = conn

	return nil
}

// subscribe header
func (c *webSocketClient) subscribeChainHead(ch chan<- *ethTypes.Header) error {
	c.connMux.Lock()
	defer c.connMux.Unlock()
	if c.conn == nil {
		return errors.New("no WebSocket connection")
	}

	c.chainHeadChan = ch
	c.chainHeadSubscriptionCallId = c.nextID()

	params, _ := json.Marshal([]interface{}{"newHeads"})

	msg := &message{
		Version: "2.0",
		ID:      c.chainHeadSubscriptionCallId,
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
	c.connMux.Lock()
	defer c.connMux.Unlock()
	if c.conn == nil {
		return errors.New("no WebSocket connection")
	}

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
func (c *webSocketClient) listen(shutdownChan <-chan struct{}) {
	for {
		// check shutdown channel
		select {
		case <-shutdownChan:
			log.Debug("WebSocket listener stopped")
			return
		default:
		}

		// TODO: we may potentially need c.conn protected with lock.
		// Currently, listen function is running in a single go routine and all resetConn function calls are initiated
		// from here. Therefore, it does not require lock protection.
		if c.conn == nil {
			if err := c.dial(c.rawUrl); err != nil {
				log.Error("Dialing failed", "error", err)
				log.Debug("Retry connection in 1 second")
				// retry connection in one second
				ticker := time.NewTicker(time.Second)
				<-ticker.C
				ticker.Stop()
				continue
			}
			if c.chainHeadSubscriptionId != "" {
				if err := c.subscribeChainHead(c.chainHeadChan); err != nil {
					log.Debug("Reconnect subscribe to chain head failed")
					c.resetConn()
					continue
				}
			}
		}

		// read message
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			log.Error("WebSocket read message error", "error", err)
			if strings.Contains(err.Error(), "EOF") {
				c.resetConn()
			}
			continue
		}
		log.Debug("WebSocket message received", "msg", string(msg))
		var receivedMsg message
		if err = json.Unmarshal(msg, &receivedMsg); err != nil {
			log.Error("Decode message error", "error", err)
			continue
		}

		// handle websocket message
		if ch := c.getPendingRPC(receivedMsg.ID); ch != nil {
			// handle rpc message
			ch <- &receivedMsg
		} else if c.chainHeadSubscriptionCallId != "" && receivedMsg.ID == c.chainHeadSubscriptionCallId {
			// handle subscription
			c.chainHeadSubscriptionCallId = ""
			c.chainHeadSubscriptionId = strings.Trim(string(receivedMsg.Result), "\"")
		} else if receivedMsg.Method == "eth_subscription" {
			// handle chain head message
			var subMsg subMessage
			if err = json.Unmarshal(receivedMsg.Params, &subMsg); err != nil {
				log.Error("Decode subscription message error", "error", err)
				continue
			}
			if c.chainHeadSubscriptionId != "" && subMsg.ID == c.chainHeadSubscriptionId {
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

func (c *webSocketClient) resetConn() {
	log.Debug("Reset WebSocket connection")
	// reset connection
	c.connMux.Lock()
	c.conn.Close()
	c.conn = nil
	c.connMux.Unlock()
}
