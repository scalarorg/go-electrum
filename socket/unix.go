package socket

import (
	"encoding/json"
	"net"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/scalarorg/go-electrum/electrum/types"
)

// Add message types
type MessageType string

const (
	Ping MessageType = "ping"
	Pong MessageType = "pong"
)

type Message struct {
	Type MessageType `json:"type"`
}

type UnixSocketServer struct {
	listener    net.Listener
	vaultTxChan <-chan *types.VaultTransaction
	connections map[net.Conn]struct{}
	mu          sync.Mutex
}

func Start(path string, vaultTxChan <-chan *types.VaultTransaction) (*UnixSocketServer, error) {
	// Create a Unix domain socket and listen for incoming connections.
	log.Info().Msgf("Starting unix socket server on %s", path)
	listener, err := net.Listen("unix", path)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create unix socket")
	}

	server := UnixSocketServer{
		listener:    listener,
		vaultTxChan: vaultTxChan,
		connections: make(map[net.Conn]struct{}),
	}
	// Listen for incoming connections
	go server.acceptConnection()
	go server.handleIncommingTransaction()
	return &server, nil
}
func (s *UnixSocketServer) Close() error {
	s.mu.Lock()
	for conn := range s.connections {
		delete(s.connections, conn)
		conn.Close()
	}
	s.mu.Unlock()
	return s.listener.Close()
}
func (s *UnixSocketServer) handleIncommingTransaction() error {
	// for vaultTx := range s.vaultTxChan {
	// 	log.Info().Msgf("Received vault transaction from socket: %v", vaultTx)
	// }
	for {
		vaultTx := <-s.vaultTxChan
		// Convert VaultTransaction to bytes before writing
		txBytes, err := vaultTx.Marshal()
		if err != nil {
			return err
		}
		if len(s.connections) == 0 {
			log.Warn().Msgf("No connections to write vaultTx to")
			continue
		}
		s.mu.Lock()
		for conn := range s.connections {
			// Write in a goroutine to not block other connections
			go func(c net.Conn) {
				log.Debug().Msgf("Write vaultTx to the socket: %x", txBytes)
				if _, err := c.Write(txBytes); err != nil {
					log.Error().Err(err).Msg("failed to write to connection")
					s.removeConnection(c)
				}
			}(conn)
		}
		s.mu.Unlock()
	}
}
func (s *UnixSocketServer) acceptConnection() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return err
		}
		s.mu.Lock()
		s.connections[conn] = struct{}{}
		s.mu.Unlock()
		// Handle connection in a separate goroutine
		go s.handleConnection(conn)
	}
}

// Move handleConnection to be a method of UnixSocketServer
func (s *UnixSocketServer) handleConnection(conn net.Conn) error {
	// Create ping ticker
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Create done channel for cleanup
	done := make(chan struct{})
	defer close(done)

	// Handle pings in separate goroutine
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				msg := Message{Type: Ping}
				pingBytes, err := json.Marshal(msg)
				if err != nil {
					log.Error().Err(err).Msg("failed to marshal ping message")
					s.removeConnection(conn)
					return
				}

				// Set write deadline
				if err := conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
					log.Error().Err(err).Msg("failed to set write deadline")
					s.removeConnection(conn)
					return
				}

				if _, err := conn.Write(pingBytes); err != nil {
					log.Error().Err(err).Msg("failed to write ping message")
					s.removeConnection(conn)
					return
				}
			}
		}
	}()

	// Main connection handler
	defer func() {
		s.removeConnection(conn)
		conn.Close()
	}()

	for {
		// Set read deadline
		if err := conn.SetReadDeadline(time.Now().Add(45 * time.Second)); err != nil {
			log.Error().Err(err).Msg("failed to set read deadline")
			return err
		}

		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			log.Error().Err(err).Msg("connection read error")
			return err
		}

		// Handle received message
		var msg Message
		if err := json.Unmarshal(buf[:n], &msg); err != nil {
			log.Error().Err(err).Msg("failed to unmarshal message")
			continue
		}

		// If it's a ping, respond with pong
		if msg.Type == Ping {
			response := Message{Type: Pong}
			responseBytes, err := json.Marshal(response)
			if err != nil {
				log.Error().Err(err).Msg("failed to marshal pong message")
				continue
			}

			if err := conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
				log.Error().Err(err).Msg("failed to set write deadline")
				return err
			}

			if _, err := conn.Write(responseBytes); err != nil {
				log.Error().Err(err).Msg("failed to write pong message")
				return err
			}
		} else if msg.Type == Pong {
			log.Debug().Msg("received pong message")
		}
	}
}

// Helper method to safely remove a connection
func (s *UnixSocketServer) removeConnection(conn net.Conn) {
	s.mu.Lock()
	delete(s.connections, conn)
	s.mu.Unlock()
}
