package kafka

import (
	"context"
	"crypto/tls"
	"log"
	"time"

	"github.com/illacloud/illa-resource-manager-backend/src/utils/config"
	kafka "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
	"go.uber.org/zap"
)

const KAFKA_CONNECTION_PROTOCOL = "tcp"
const KAFKA_WRITE_DEADLINE = 10 * time.Second

// Write Client
type KafkaWriterClient struct {
	Writer *kafka.Writer
	Logger *zap.SugaredLogger
}

func GetWriterClient(config *config.Config, topic string, logger *zap.SugaredLogger) *KafkaWriterClient {
	w := &kafka.Writer{
		Addr:     kafka.TCP(config.GetKafkaReplicas()),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	if config.GetKafkaProvider() != "" {
		mechanism, err := scram.Mechanism(scram.SHA256, config.GetKafkaUsername(), config.GetKafkaPassword())
		if err != nil {
			log.Fatalln(err)
		}

		sharedTransport := &kafka.Transport{
			SASL: mechanism,
			TLS:  &tls.Config{},
		}

		w.Transport = sharedTransport
	}

	client := &KafkaWriterClient{
		Writer: w,
		Logger: logger,
	}
	return client
}

func (writeClient *KafkaWriterClient) Write(key, message []byte) error {
	msg := kafka.Message{
		Key:   key,
		Value: message,
	}
	errInWrite := writeClient.Writer.WriteMessages(context.Background(), msg)
	if errInWrite != nil {
		log.Fatal("kafka failed to write messages:", errInWrite)
		return errInWrite
	}
	return nil
}

func (writeClient *KafkaWriterClient) Close() error {
	if err := writeClient.Writer.Close(); err != nil {
		log.Fatal("kafka failed to close writer:", err)
		return err
	}
	return nil
}

// Read Client
type KafkaReaderClient struct {
	Reader *kafka.Reader
	Logger *zap.SugaredLogger
}

func GetReaderClient(config *config.Config, topic string, logger *zap.SugaredLogger) *KafkaReaderClient {
	rCfg := kafka.ReaderConfig{
		Brokers:  config.GetKafkaReplicasInArray(),
		GroupID:  config.GetKafkaGroupID(),
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	}

	if config.GetKafkaProvider() != "" {
		mechanism, err := scram.Mechanism(scram.SHA256, config.GetKafkaUsername(), config.GetKafkaPassword())
		if err != nil {
			log.Fatalln(err)
		}

		dialer := &kafka.Dialer{
			Timeout:       10 * time.Second,
			DualStack:     true,
			SASLMechanism: mechanism,
			TLS:           &tls.Config{},
		}

		rCfg.Dialer = dialer
	}

	r := kafka.NewReader(rCfg)

	client := &KafkaReaderClient{
		Reader: r,
		Logger: logger,
	}
	return client
}

func (writeClient *KafkaReaderClient) Close() error {
	if err := writeClient.Reader.Close(); err != nil {
		log.Fatal("kafka failed to close writer:", err)
		return err
	}
	return nil
}
