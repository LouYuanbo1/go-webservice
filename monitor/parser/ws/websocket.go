package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/LouYuanbo1/go-webservice/monitor/parser"
	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Client struct {
	conn     *websocket.Conn
	register *Register
	send     chan []byte
}

type Register struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
}

func NewRegister() *Register {
	return &Register{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (r *Register) Run() {
	for {
		select {
		case client := <-r.register:
			r.mu.Lock()
			r.clients[client] = true
			r.mu.Unlock()
			slog.Info("client connected", "total", len(r.clients))
		case client := <-r.unregister:
			r.mu.Lock()
			if _, ok := r.clients[client]; ok {
				delete(r.clients, client)
				close(client.send)
			}
			r.mu.Unlock()
			slog.Info("client disconnected", "total", len(r.clients))
		case message := <-r.broadcast:
			r.mu.Lock()
			for client := range r.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(r.clients, client)
				}
			}
			r.mu.Unlock()
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.register.unregister <- c
		_ = c.conn.Close(websocket.StatusNormalClosure, "")
	}()

	c.conn.SetReadLimit(512)

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		_, _, err := c.conn.Read(ctx)
		cancel()
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
				websocket.CloseStatus(err) == websocket.StatusGoingAway {
				slog.Debug("client closed connection gracefully")
			} else {
				slog.Error("read error", "err", err)
			}
			break
		}
	}
}

func (c *Client) writePump() {
	defer func() {
		_ = c.conn.Close(websocket.StatusNormalClosure, "")
	}()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			err := c.conn.Write(ctx, websocket.MessageText, message)
			cancel()
			if err != nil {
				return
			}

			n := len(c.send)
			for i := 0; i < n; i++ {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				_ = c.conn.Write(ctx, websocket.MessageText, append([]byte("\n"), <-c.send...))
				cancel()
			}
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			if err := c.conn.Ping(ctx); err != nil {
				cancel()
				return
			}
			cancel()
		}
	}
}

func ServeWs(register *Register, c *gin.Context) {
	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		OriginPatterns:  []string{"*"},
		CompressionMode: websocket.CompressionDisabled,
	})
	if err != nil {
		slog.Error("failed to accept websocket", "err", err)
		return
	}

	client := &Client{
		conn:     conn,
		register: register,
		send:     make(chan []byte, 256),
	}

	client.register.register <- client

	go client.writePump()
	go client.readPump()
}

type MetricsData struct {
	Timestamp   int64                  `json:"timestamp"`
	MetricTypes []string               `json:"metric_types"`
	Families    []*parser.MetricFamily `json:"families"`
	Summary     *MetricsSummary        `json:"summary"`
}

type MetricsSummary struct {
	TotalRequests     float64 `json:"total_requests"`
	AvgResponseTime   float64 `json:"avg_response_time"`
	ActiveConnections int     `json:"active_connections"`
}

func StartMetricsBroadcaster(register *Register, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		data, err := gatherMetrics(len(register.clients))
		if err != nil {
			slog.Error("failed to gather metrics", "err", err)
			continue
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			slog.Error("failed to marshal metrics", "err", err)
			continue
		}

		register.broadcast <- jsonData
	}
}

func gatherMetrics(activeConnections int) (*MetricsData, error) {
	families, err := parser.GatherAndParse()
	if err != nil {
		return nil, fmt.Errorf("failed to gather metrics: %w", err)
	}

	var totalRequests float64
	var avgResponseTime float64

	for _, family := range families {
		if family.Name == "demo_app_http_requests_total" {
			totalRequests = family.Sum()
		}
		if family.Name == "demo_app_http_request_duration_seconds" {
			avgResponseTime = family.Avg()
		}
	}

	return &MetricsData{
		Timestamp:   time.Now().Unix(),
		MetricTypes: getMetricTypes(families),
		Families:    families,
		Summary: &MetricsSummary{
			TotalRequests:     totalRequests,
			AvgResponseTime:   avgResponseTime,
			ActiveConnections: activeConnections,
		},
	}, nil
}

func getMetricTypes(families []*parser.MetricFamily) []string {
	types := make(map[string]bool)
	for _, family := range families {
		types[family.Type.String()] = true
	}
	result := make([]string, 0, len(types))
	for t := range types {
		result = append(result, t)
	}
	return result
}

func SetupRouter(register *Register) *gin.Engine {
	r := gin.Default()

	r.GET("/ws", func(c *gin.Context) {
		ServeWs(register, c)
	})

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	r.Static("/dashboard", "../dashboard")
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/dashboard/index.html")
	})

	return r
}
