package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
)

var errNotFound = errors.New("artist not found")

type voteRequest struct {
	WinnerID string `json:"winner_id" binding:"required"`
	LoserID  string `json:"loser_id" binding:"required"`
}

type matchupResponse struct {
	Artists [2]Artist `json:"artists"`
}

func main() {
	// Setup Cassandra
	session, err := setupCassandra()
	if err != nil {
		log.Fatalf("Failed to connect to Cassandra: %v", err)
	}
	defer session.Close()

	// Setup Kafka
	kafka := newKafkaProducer()
	defer kafka.Close()

	// Setup store
	store := newCassandraStore(session, kafka)

	// Setup Kafka consumer for background ELO processing
	consumer := newKafkaConsumer(store)
	defer consumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start background ELO processor
	go consumer.StartEloProcessor(ctx)

	// Setup Gin router
	r := gin.Default()
	r.Use(cors())

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "store": "cassandra", "queue": "kafka"})
	})

	r.GET("/api/matchup", func(c *gin.Context) {
		pair, err := store.matchup()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate matchup"})
			return
		}
		if pair[0].ID == "" || pair[1].ID == "" {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "not enough artists"})
			return
		}
		c.JSON(http.StatusOK, matchupResponse{Artists: pair})
	})

	r.POST("/api/vote", func(c *gin.Context) {
		var req voteRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}
		if req.WinnerID == req.LoserID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "winner and loser must differ"})
			return
		}
		if err := store.vote(req.WinnerID, req.LoserID); err != nil {
			if errors.Is(err, errNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "artist not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not record vote"})
			return
		}
		c.Status(http.StatusNoContent)
	})

	r.GET("/api/leaderboard", func(c *gin.Context) {
		kind := c.DefaultQuery("type", "overall")
		lb, err := store.leaderboard(kind)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch leaderboard"})
			return
		}
		c.JSON(http.StatusOK, lb)
	})

	r.GET("/api/trending", func(c *gin.Context) {
		// For now, return same as leaderboard since we don't have trending metrics implemented yet
		rows, err := store.leaderboard("velocity")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch trending"})
			return
		}
		c.JSON(http.StatusOK, rows)
	})

	// Graceful shutdown
	go func() {
		log.Println("Server starting on :8080")
		if err := r.Run(":8080"); err != nil {
			log.Printf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	cancel()
}
