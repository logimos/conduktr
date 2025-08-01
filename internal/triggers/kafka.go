package triggers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/logimos/conduktr/internal/engine"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// KafkaTrigger implements Kafka-based event triggering
type KafkaTrigger struct {
	readers []*kafka.Reader
	writer  *kafka.Writer
	engine  *engine.Engine
	logger  *zap.Logger
	ctx     context.Context
	cancel  context.CancelFunc
}

// KafkaConfig holds Kafka connection configuration
type KafkaConfig struct {
	Brokers    []string `yaml:"brokers"`
	GroupID    string   `yaml:"group_id"`
	Topics     []string `yaml:"topics"`
	AutoCommit bool     `yaml:"auto_commit"`
}

// NewKafkaTrigger creates a new Kafka trigger
func NewKafkaTrigger(config KafkaConfig, engine *engine.Engine, logger *zap.Logger) *KafkaTrigger {
	ctx, cancel := context.WithCancel(context.Background())

	var readers []*kafka.Reader
	for _, topic := range config.Topics {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:     config.Brokers,
			Topic:       topic,
			GroupID:     config.GroupID,
			StartOffset: kafka.LastOffset,
			MinBytes:    10e3, // 10KB
			MaxBytes:    10e6, // 10MB
		})
		readers = append(readers, reader)
	}

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(config.Brokers...),
		Topic:                  "reactor-events",
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}

	return &KafkaTrigger{
		readers: readers,
		writer:  writer,
		engine:  engine,
		logger:  logger,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Start begins consuming from Kafka topics
func (k *KafkaTrigger) Start() error {
	k.logger.Info("Kafka trigger started", zap.Int("topics", len(k.readers)))

	// Start a consumer for each topic
	for _, reader := range k.readers {
		go k.consume(reader)
	}

	return nil
}

// Stop stops the Kafka trigger
func (k *KafkaTrigger) Stop() error {
	k.cancel()

	for _, reader := range k.readers {
		if err := reader.Close(); err != nil {
			k.logger.Error("Failed to close Kafka reader", zap.Error(err))
		}
	}

	return k.writer.Close()
}

// consume reads messages from a Kafka topic
func (k *KafkaTrigger) consume(reader *kafka.Reader) {
	defer reader.Close()

	for {
		select {
		case <-k.ctx.Done():
			return
		default:
			msg, err := reader.ReadMessage(k.ctx)
			if err != nil {
				if err == context.Canceled {
					return
				}
				k.logger.Error("Kafka read error", zap.Error(err))
				time.Sleep(time.Second * 5) // Backoff on error
				continue
			}

			k.handleMessage(msg)
		}
	}
}

// handleMessage processes a Kafka message
func (k *KafkaTrigger) handleMessage(msg kafka.Message) {
	k.logger.Info("Received Kafka message",
		zap.String("topic", msg.Topic),
		zap.Int("partition", msg.Partition),
		zap.Int64("offset", msg.Offset),
		zap.Time("timestamp", msg.Time))

	// Parse message value as JSON
	var eventData map[string]interface{}
	if err := json.Unmarshal(msg.Value, &eventData); err != nil {
		k.logger.Error("Failed to parse Kafka message", zap.Error(err))
		return
	}

	// Determine event type from topic or message headers
	eventType := k.extractEventType(msg)

	// Create context with Kafka-specific metadata
	context := map[string]interface{}{
		"trigger_type": "kafka",
		"topic":        msg.Topic,
		"partition":    msg.Partition,
		"offset":       msg.Offset,
		"event_type":   eventType,
		"timestamp":    msg.Time.Unix(),
		"key":          string(msg.Key),
	}

	// Add message headers to context
	if len(msg.Headers) > 0 {
		headers := make(map[string]string)
		for _, header := range msg.Headers {
			headers[header.Key] = string(header.Value)
		}
		context["headers"] = headers
	}

	// Merge event data with context
	for k, v := range eventData {
		context[k] = v
	}

	// Execute workflow asynchronously with retry
	go executeWorkflow(k.ctx, k.engine, k.logger, eventType, context)
}

// extractEventType determines event type from Kafka message
func (k *KafkaTrigger) extractEventType(msg kafka.Message) string {
	// Check for event type in headers first
	for _, header := range msg.Headers {
		if header.Key == "event-type" || header.Key == "eventType" {
			return string(header.Value)
		}
	}

	// Parse from message content
	var eventData map[string]interface{}
	if err := json.Unmarshal(msg.Value, &eventData); err == nil {
		if eventType, ok := eventData["event_type"].(string); ok {
			return eventType
		}
		if eventType, ok := eventData["type"].(string); ok {
			return eventType
		}
	}

	// Fallback to topic-based event type
	// Convert topic like "user-events" to "user.event"
	parts := strings.Split(msg.Topic, "-")
	if len(parts) >= 2 {
		return fmt.Sprintf("%s.%s", parts[0], strings.Join(parts[1:], "."))
	}

	return msg.Topic
}

// PublishEvent publishes an event to Kafka (utility method)
func (k *KafkaTrigger) PublishEvent(eventType string, data map[string]interface{}) error {
	// Add event type to data
	data["event_type"] = eventType
	data["timestamp"] = time.Now().Unix()

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	message := kafka.Message{
		Key:   []byte(eventType),
		Value: jsonData,
		Headers: []kafka.Header{
			{Key: "event-type", Value: []byte(eventType)},
			{Key: "source", Value: []byte("reactor")},
		},
	}

	return k.writer.WriteMessages(k.ctx, message)
}

// PublishToTopic publishes an event to a specific Kafka topic
func (k *KafkaTrigger) PublishToTopic(topic, eventType string, data map[string]interface{}) error {
	writer := &kafka.Writer{
		Addr:                   k.writer.Addr,
		Topic:                  topic,
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}
	defer writer.Close()

	data["event_type"] = eventType
	data["timestamp"] = time.Now().Unix()

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	message := kafka.Message{
		Key:   []byte(eventType),
		Value: jsonData,
		Headers: []kafka.Header{
			{Key: "event-type", Value: []byte(eventType)},
			{Key: "source", Value: []byte("reactor")},
		},
	}

	return writer.WriteMessages(k.ctx, message)
}
