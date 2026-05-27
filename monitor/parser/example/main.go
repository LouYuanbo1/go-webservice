package main

import (
	"log"
	"net/http"
	"time"

	"github.com/LouYuanbo1/go-webservice/monitor"
	"github.com/LouYuanbo1/go-webservice/monitor/parser/ws"
	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	mw, err := monitor.NewMetricsMiddleware(monitor.MetricsConfig{
		Namespace: "demo",
		Subsystem: "app",
	}, nil)
	if err != nil {
		log.Fatalf("Failed to create metrics middleware: %v", err)
	}

	wsRegister := ws.NewRegister()
	go wsRegister.Run()

	go ws.StartMetricsBroadcaster(wsRegister, time.Second)

	r := ws.SetupRouter(wsRegister)

	r.GET("/api/hello", func(c *gin.Context) {
		start := time.Now()
		time.Sleep(time.Millisecond * 50)
		duration := time.Since(start).Seconds()
		mw.Record("/api/hello", http.StatusOK, duration)
		c.JSON(http.StatusOK, gin.H{"message": "Hello, World!"})
	})

	r.GET("/api/test", func(c *gin.Context) {
		start := time.Now()
		time.Sleep(time.Millisecond * 30)
		duration := time.Since(start).Seconds()
		mw.Record("/api/test", http.StatusOK, duration)
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	log.Println("Server started at http://localhost:8080")
	log.Println("Dashboard: http://localhost:8080/dashboard/index.html")
	log.Fatal(r.Run(":8080"))
}
