package service

import (
	bfxlog "github.com/bitfinexcom/bfxfixgw/log"
	"github.com/bitfinexcom/bitfinex-api-go/v2"
	"github.com/bitfinexcom/bitfinex-api-go/v2/rest"
	"github.com/bitfinexcom/bitfinex-api-go/v2/websocket"
	"go.uber.org/zap"
	"log"
)

type ClientFactory interface {
	NewRest() *rest.Client
	NewWs() *websocket.Client
}

// Peers is an interface to create, remove, and lookup peers.
type Peers interface {
	FindPeer(id string) (*Peer, bool)
	RemovePeer(id string) bool
	AddPeer(id string)
}

// Peer represents a FIX-websocket peer user
type Peer struct {
	Ws   *websocket.Client
	Rest *rest.Client

	logger *zap.Logger

	bfxUserID    string
	fixSessionID string
}

// could be from FIX market data, or FIX order flow
type subscription struct {
	Request        *websocket.SubscriptionRequest
	SubscriptionID string
}

// NewPeer creates a peer, but does not establish a websocket connection yet
func newPeer(factory ClientFactory, fixSessionID string) *Peer {
	return &Peer{
		Ws:           factory.NewWs(),
		Rest:         factory.NewRest(),
		logger:       bfxlog.Logger,
		fixSessionID: fixSessionID,
	}
}

// Logon establishes a websocket connection and attempts to authenticate with the given apiKey and apiSecret
func (p *Peer) Logon(apiKey, apiSecret, bfxUserID string) error {
	p.Ws.Credentials(apiKey, apiSecret)
	p.bfxUserID = bfxUserID
	err := p.Ws.Connect()
	if err != nil {
		return err
	}
	go p.listen()
	return nil
}

func (p *Peer) listen() {
	for msg := range p.Ws.Listen() {
		log.Printf("peer got msg: %#v", msg)
		if msg == nil {
			p.logger.Info("upstream disconnect")
			// TODO log out peer
			return
		}
		switch m := msg.(type) {
		case *websocket.InfoEvent:
			// TODO logon? no logon--client has not yet set credentials
		case *websocket.AuthEvent:
			// TODO log off FIX session if auth error
			log.Printf("auth: %#v", m)
		case *bitfinex.BookUpdate:
			// TODO
		default:
			p.logger.Error("unhandled event: %#v", zap.Any("msg", msg))
		}
	}
}

// BfxUserID is an immutable accessor to the bitfinex user ID
func (p *Peer) BfxUserID() string {
	return p.bfxUserID
}

func (p *Peer) FIXSessionID() string {
	return p.fixSessionID
}

func (p *Peer) Close() {
	p.Ws.Close()
}
