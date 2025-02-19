package tracker

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
)

type Handler interface {
	Handle(conn *net.UDPConn, addr *net.UDPAddr, request []byte, transactionID uint32)
}

type Router struct {
	handlers map[uint32]Handler
	mutex    sync.Mutex
}

// NewRouter creates a new Router instance
func NewRouter() *Router {
	return &Router{
		handlers: make(map[uint32]Handler),
	}
}

// RegisterHandler registers a handler for a specific action type (e.g., Connect, Announce)
func (r *Router) RegisterHandler(action uint32, handler Handler) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.handlers[action] = handler
}

// Route processes incoming requests by delegating them to the appropriate handler
func (r *Router) Route(conn *net.UDPConn, addr *net.UDPAddr, request []byte) {
	if len(request) < 16 {
		fmt.Println("Received too short packet")
		return
	}

	// Read the action (next 4 bytes)
	action := binary.BigEndian.Uint32(request[8:12])
	// Read the transaction ID (next 4 bytes)
	transactionID := binary.BigEndian.Uint32(request[12:16])

	// Lookup the handler based on the action
	r.mutex.Lock()
	handler, found := r.handlers[action]
	r.mutex.Unlock()

	if found {
		// Delegate the request to the appropriate handler
		handler.Handle(conn, addr, request, transactionID)
	} else {
		fmt.Printf("No handler for action: %d\n", action)
	}
}
