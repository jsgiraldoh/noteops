package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/johansgiraldo/noteops/backend/internal/repository"
	"github.com/johansgiraldo/noteops/backend/internal/service"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // en producción validar origen
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// client representa una conexión WebSocket activa
type client struct {
	conn      *websocket.Conn
	sessionID uuid.UUID
	send      chan []byte
}

// Hub gestiona todas las conexiones WebSocket agrupadas por session_id
type Hub struct {
	mu       sync.RWMutex
	sessions map[uuid.UUID]map[*client]bool
	repo     *repository.Repository
	svc      *service.Service
}

func NewHub(repo *repository.Repository, svc *service.Service) *Hub {
	h := &Hub{
		sessions: make(map[uuid.UUID]map[*client]bool),
		repo:     repo,
		svc:      svc,
	}
	go h.tickAll()
	return h
}

// tickAll emite el estado del reloj a todos los clientes cada segundo
func (h *Hub) tickAll() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		h.mu.RLock()
		for sessionID, clients := range h.sessions {
			if len(clients) == 0 {
				continue
			}

			session, err := h.repo.GetSessionByID(context.Background(), sessionID)
			if err != nil {
				continue
			}

			tick := h.svc.ComputeSessionTick(session)
			payload, _ := json.Marshal(tick)

			for c := range clients {
				select {
				case c.send <- payload:
				default:
					// canal lleno — cliente lento, desconectar
					close(c.send)
					delete(clients, c)
				}
			}
		}
		h.mu.RUnlock()
	}
}

func (h *Hub) register(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.sessions[c.sessionID] == nil {
		h.sessions[c.sessionID] = make(map[*client]bool)
	}
	h.sessions[c.sessionID][c] = true
}

func (h *Hub) unregister(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if clients, ok := h.sessions[c.sessionID]; ok {
		delete(clients, c)
	}
}

// ServeWS maneja la conexión WebSocket de un cliente
// GET /ws/session/:id
func (h *Hub) ServeWS(c *gin.Context) {
	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("ws upgrade error: %v", err)
		return
	}

	cl := &client{
		conn:      conn,
		sessionID: sessionID,
		send:      make(chan []byte, 64),
	}
	h.register(cl)

	// goroutine escritura
	go func() {
		defer func() {
			h.unregister(cl)
			conn.Close()
		}()
		for msg := range cl.send {
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		}
	}()

	// goroutine lectura (mantiene la conexión viva, detecta cierre)
	go func() {
		defer func() {
			h.unregister(cl)
			close(cl.send)
			conn.Close()
		}()
		conn.SetReadLimit(512)
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}
