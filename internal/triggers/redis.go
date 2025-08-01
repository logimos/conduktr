package triggers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/logimos/conduktr/internal/engine"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// RedisTrigger implements Redis-based event triggering
type RedisTrigger struct {
	client *redis.Client
	engine *engine.Engine
	logger *zap.Logger
	ctx    context.Context
	cancel context.CancelFunc
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Address  string   `yaml:"address"`
	Password string   `yaml:"password"`
	DB       int      `yaml:"db"`
	Channels []string `yaml:"channels"`
	Streams  []string `yaml:"streams"`
}

// NewRedisTrigger creates a new Redis trigger
func NewRedisTrigger(config RedisConfig, engine *engine.Engine, logger *zap.Logger) *RedisTrigger {
	ctx, cancel := context.WithCancel(context.Background())

	client := redis.NewClient(&redis.Options{
		Addr:     config.Address,
		Password: config.Password,
		DB:       config.DB,
	})

	return &RedisTrigger{
		client: client,
		engine: engine,
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start begins listening for Redis events
func (r *RedisTrigger) Start() error {
	// Test connection
	if err := r.client.Ping(r.ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	r.logger.Info("Redis trigger started", zap.String("address", r.client.Options().Addr))

	// Start pub/sub listener
	go r.listenPubSub()

	// Start stream listener
	go r.listenStreams()

	return nil
}

// Stop stops the Redis trigger
func (r *RedisTrigger) Stop() error {
	r.cancel()
	return r.client.Close()
}

// listenPubSub listens to Redis pub/sub channels
func (r *RedisTrigger) listenPubSub() {
	pubsub := r.client.PSubscribe(r.ctx, "reactor:*")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case msg := <-ch:
			r.handlePubSubMessage(msg)
		case <-r.ctx.Done():
			return
		}
	}
}

// listenStreams listens to Redis streams
func (r *RedisTrigger) listenStreams() {
	streams := []string{"reactor:events", ">"}

	for {
		select {
		case <-r.ctx.Done():
			return
		default:
			result, err := r.client.XRead(r.ctx, &redis.XReadArgs{
				Streams: streams,
				Block:   time.Second * 5,
				Count:   10,
			}).Result()

			if err != nil {
				if err != redis.Nil {
					r.logger.Error("Redis stream read error", zap.Error(err))
				}
				continue
			}

			for _, stream := range result {
				for _, message := range stream.Messages {
					r.handleStreamMessage(message)
				}
			}
		}
	}
}

// handlePubSubMessage processes pub/sub messages
func (r *RedisTrigger) handlePubSubMessage(msg *redis.Message) {
	r.logger.Info("Received Redis pub/sub message",
		zap.String("channel", msg.Channel),
		zap.String("payload", msg.Payload))

	var eventData map[string]interface{}
	if err := json.Unmarshal([]byte(msg.Payload), &eventData); err != nil {
		r.logger.Error("Failed to parse Redis message", zap.Error(err))
		return
	}

	// Extract event type from channel name (reactor:event.type)
	eventType := msg.Channel[8:] // Remove "reactor:" prefix

	// Create context with Redis-specific metadata
	context := map[string]interface{}{
		"trigger_type": "redis_pubsub",
		"channel":      msg.Channel,
		"event_type":   eventType,
		"timestamp":    time.Now().Unix(),
	}

	// Merge event data with context
	for k, v := range eventData {
		context[k] = v
	}

	// Execute workflow asynchronously
	go r.executeWorkflow(eventType, context)
}

// handleStreamMessage processes stream messages
func (r *RedisTrigger) handleStreamMessage(msg redis.XMessage) {
	r.logger.Info("Received Redis stream message",
		zap.String("id", msg.ID),
		zap.Any("values", msg.Values))

	eventType, ok := msg.Values["event"].(string)
	if !ok {
		r.logger.Error("No event type in stream message")
		return
	}

	// Create context with stream-specific metadata
	context := map[string]interface{}{
		"trigger_type": "redis_stream",
		"stream_id":    msg.ID,
		"event_type":   eventType,
		"timestamp":    time.Now().Unix(),
	}

	// Add all stream values to context
	for k, v := range msg.Values {
		if k != "event" { // Don't duplicate event type
			context[k] = v
		}
	}

	// Execute workflow asynchronously
	go r.executeWorkflow(eventType, context)
}

// PublishEvent publishes an event to Redis (utility method)
func (r *RedisTrigger) PublishEvent(eventType string, data map[string]interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	channel := fmt.Sprintf("reactor:%s", eventType)
	return r.client.Publish(r.ctx, channel, jsonData).Err()
}

// AddToStream adds an event to Redis stream (utility method)
func (r *RedisTrigger) AddToStream(eventType string, data map[string]interface{}) error {
	streamData := make(map[string]interface{})
	streamData["event"] = eventType

	for k, v := range data {
		streamData[k] = v
	}

	return r.client.XAdd(r.ctx, &redis.XAddArgs{
		Stream: "reactor:events",
		Values: streamData,
	}).Err()
}

// executeWorkflow helper function to execute workflows
func (r *RedisTrigger) executeWorkflow(eventType string, context map[string]interface{}) {
	executeWorkflow(r.ctx, r.engine, r.logger, eventType, context)
}
