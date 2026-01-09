package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type VoteEvent struct {
	VoteID          string    `json:"vote_id"`
	WinnerID        string    `json:"winner_id"`
	LoserID         string    `json:"loser_id"`
	Timestamp       time.Time `json:"timestamp"`
	WinnerEloBefore int       `json:"winner_elo_before"`
	LoserEloBefore  int       `json:"loser_elo_before"`
}

type KafkaProducer struct {
	writer *kafka.Writer
}

func newKafkaProducer() *KafkaProducer {
	return &KafkaProducer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP("localhost:9092"),
			Topic:    "votes",
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (kp *KafkaProducer) PublishVote(event VoteEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return kp.writer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(event.VoteID),
			Value: data,
		},
	)
}

func (kp *KafkaProducer) Close() error {
	return kp.writer.Close()
}

type KafkaConsumer struct {
	reader *kafka.Reader
	store  *CassandraStore
}

func newKafkaConsumer(store *CassandraStore) *KafkaConsumer {
	return &KafkaConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  []string{"localhost:9092"},
			Topic:    "votes",
			GroupID:  "elo-calculator",
			MinBytes: 10e3, // 10KB
			MaxBytes: 10e6, // 10MB
		}),
		store: store,
	}
}

func (kc *KafkaConsumer) StartEloProcessor(ctx context.Context) {
	log.Println("Starting ELO processor...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping ELO processor")
			return
		default:
			message, err := kc.reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Failed to read message: %v", err)
				continue
			}

			var event VoteEvent
			if err := json.Unmarshal(message.Value, &event); err != nil {
				log.Printf("Failed to unmarshal vote event: %v", err)
				continue
			}

			if err := kc.store.ProcessVoteEvent(event); err != nil {
				log.Printf("Failed to process vote event: %v", err)
				continue
			}

			log.Printf("Processed vote: %s beat %s", event.WinnerID, event.LoserID)
		}
	}
}

func (kc *KafkaConsumer) Close() error {
	return kc.reader.Close()
}
