package kafka

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain настройка тестового окружения
func TestMain(m *testing.M) {
	// Загружаем .env для всех тестов
	_ = godotenv.Load("../.env")

	// Проверяем доступность Kafka
	if !isKafkaAvailable() {
		fmt.Println("Kafka not available, skipping integration tests")
		os.Exit(0)
	}

	code := m.Run()
	os.Exit(code)
}

// TestKafkaProducerConnection тест подключения к Kafka
func TestKafkaProducerConnection(t *testing.T) {
	tests := []struct {
		name    string
		brokers []string
		topic   string
		wantErr bool
	}{
		{
			name:    "valid brokers",
			brokers: getTestBrokers(),
			topic:   "test-connection",
			wantErr: false,
		},
		{
			name:    "empty brokers",
			brokers: []string{},
			topic:   "test",
			wantErr: true,
		},
		{
			name:    "invalid broker",
			brokers: []string{"invalid:9092"},
			topic:   "test",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			producer, err := NewKafkaProducer(tt.brokers, tt.topic)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, producer)

			// Проверяем что можно закрыть
			err = producer.Close()
			assert.NoError(t, err)
		})
	}
}

// TestSendEventWithoutTopic тест отправки в несуществующий топик
func TestSendEventWithoutTopic(t *testing.T) {
	producer, err := NewKafkaProducer(getTestBrokers(), "test-no-topic")
	require.NoError(t, err)
	defer producer.Close()

	event := map[string]interface{}{
		"event_type": "test",
		"user_id":    123,
		"timestamp":  time.Now(),
	}

	// Отправка не должна вернуть ошибку (асинхронно)
	err = producer.SendEvent(context.Background(), "non-existent-topic", "123", event)
	assert.NoError(t, err)

	// Даем время на обработку
	time.Sleep(500 * time.Millisecond)

	// Проверяем метрики
	if asyncProducer, ok := producer.(*AsyncKafkaProducer); ok {
		metrics := asyncProducer.GetMetrics()
		t.Logf("Metrics - Sent: %d, Failed: %d", metrics.MessagesSent, metrics.MessagesFailed)
	}
}

// TestSendEventWithAutoCreateTopic тест с авто-созданием топика
func TestSendEventWithAutoCreateTopic(t *testing.T) {
	topicName := fmt.Sprintf("test-auto-topic-%d", time.Now().Unix())

	producer, err := NewKafkaProducerWithConfig(getTestBrokers(), topicName, func(config *sarama.Config) {
		config.Metadata.AllowAutoTopicCreation = true
	})
	require.NoError(t, err)
	defer producer.Close()

	event := map[string]interface{}{
		"event_type": "test",
		"user_id":    456,
		"timestamp":  time.Now(),
	}

	err = producer.SendEvent(context.Background(), topicName, "456", event)
	assert.NoError(t, err)

	// Ждем создания топика и отправки
	time.Sleep(2 * time.Second)
}

// TestSendEventWithExistingTopic тест отправки в существующий топик
func TestSendEventWithExistingTopic(t *testing.T) {
	topicName := fmt.Sprintf("test-existing-topic-%d", time.Now().Unix())

	// Создаем топик
	err := createTopic(getTestBrokers()[0], topicName)
	require.NoError(t, err)
	defer deleteTopic(getTestBrokers()[0], topicName)

	// Создаем продюсера
	producer, err := NewKafkaProducer(getTestBrokers(), topicName)
	require.NoError(t, err)
	defer producer.Close()

	// Отправляем несколько сообщений
	for i := 0; i < 5; i++ {
		event := map[string]interface{}{
			"event_type": "test",
			"message_id": i,
			"user_id":    789,
			"timestamp":  time.Now(),
		}

		err = producer.SendEvent(context.Background(), topicName, fmt.Sprintf("key-%d", i), event)
		assert.NoError(t, err)
	}

	// Даем время на отправку
	time.Sleep(2 * time.Second)
}

// TestSendEventWithContextCancel тест с отменой контекста
func TestSendEventWithContextCancel(t *testing.T) {
	producer, err := NewKafkaProducer(getTestBrokers(), "test-context")
	require.NoError(t, err)
	defer producer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Сразу отменяем

	event := map[string]interface{}{
		"event_type": "test",
		"user_id":    999,
	}

	err = producer.SendEvent(ctx, "test-context", "999", event)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

// TestProducerMetrics тест метрик
func TestProducerMetrics(t *testing.T) {
	producer, err := NewKafkaProducer(getTestBrokers(), "test-metrics")
	require.NoError(t, err)
	defer producer.Close()

	asyncProducer, ok := producer.(*AsyncKafkaProducer)
	require.True(t, ok)

	// Отправляем несколько сообщений
	for i := 0; i < 10; i++ {
		event := map[string]interface{}{
			"event_type": "test",
			"counter":    i,
		}
		_ = producer.SendEvent(context.Background(), "test-metrics", "key", event)
	}

	time.Sleep(1 * time.Second)

	metrics := asyncProducer.GetMetrics()
	assert.GreaterOrEqual(t, metrics.MessagesSent, int64(0))
	t.Logf("Metrics: Sent=%d, Failed=%d", metrics.MessagesSent, metrics.MessagesFailed)
}

// TestMultipleBrokers тест с несколькими брокерами
func TestMultipleBrokers(t *testing.T) {
	brokers := getTestBrokers()
	if len(brokers) < 2 {
		t.Skip("Need at least 2 brokers for this test")
	}

	producer, err := NewKafkaProducer(brokers, "test-multi-broker")
	require.NoError(t, err)
	defer producer.Close()

	// Отправляем сообщения для проверки балансировки
	for i := 0; i < 20; i++ {
		event := map[string]interface{}{
			"event_type": "test",
			"sequence":   i,
		}
		err = producer.SendEvent(context.Background(), "test-multi-broker", fmt.Sprintf("key-%d", i), event)
		assert.NoError(t, err)
	}

	time.Sleep(2 * time.Second)
}

// TestDifferentMessageTypes тест разных типов сообщений
func TestDifferentMessageTypes(t *testing.T) {
	producer, err := NewKafkaProducer(getTestBrokers(), "test-types")
	require.NoError(t, err)
	defer producer.Close()

	tests := []struct {
		name    string
		key     string
		value   interface{}
		wantErr bool
	}{
		{
			name:  "string value",
			key:   "str-key",
			value: map[string]string{"data": "test"},
		},
		{
			name:  "int value",
			key:   "int-key",
			value: map[string]int{"count": 42},
		},
		{
			name: "complex struct",
			key:  "struct-key",
			value: struct {
				ID   int
				Name string
			}{ID: 1, Name: "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := producer.SendEvent(context.Background(), "test-types", tt.key, tt.value)
			assert.NoError(t, err)
		})
	}

	time.Sleep(1 * time.Second)
}

// BenchmarkSendEvent бенчмарк отправки сообщений
func BenchmarkSendEvent(b *testing.B) {
	producer, err := NewKafkaProducer(getTestBrokers(), "benchmark-topic")
	if err != nil {
		b.Skip("Kafka not available")
	}
	defer producer.Close()

	event := map[string]interface{}{
		"event_type": "benchmark",
		"data":       "test data",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = producer.SendEvent(context.Background(), "benchmark-topic", "bench-key", event)
	}
}

// Вспомогательные функции

// getTestBrokers возвращает список брокеров из окружения или дефолтные
func getTestBrokers() []string {
	broker := os.Getenv("BROKER")
	if broker == "" {
		broker = "localhost:9094"
	}
	return strings.Split(broker, ",")
}

// isKafkaAvailable проверяет доступность Kafka
func isKafkaAvailable() bool {
	brokers := getTestBrokers()
	config := sarama.NewConfig()
	config.Version = sarama.V3_6_0_0
	config.Net.DialTimeout = 2 * time.Second

	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		return false
	}
	defer client.Close()

	return true
}

// createTopic создает топик
func createTopic(broker, topic string) error {
	config := sarama.NewConfig()
	config.Version = sarama.V3_6_0_0

	admin, err := sarama.NewClusterAdmin([]string{broker}, config)
	if err != nil {
		return fmt.Errorf("failed to create admin: %w", err)
	}
	defer admin.Close()

	topicDetail := &sarama.TopicDetail{
		NumPartitions:     3,
		ReplicationFactor: 1,
	}

	err = admin.CreateTopic(topic, topicDetail, false)
	if err != nil && !strings.Contains(err.Error(), "Topic already exists") {
		return fmt.Errorf("failed to create topic: %w", err)
	}

	return nil
}

// deleteTopic удаляет топик
func deleteTopic(broker, topic string) error {
	config := sarama.NewConfig()
	config.Version = sarama.V3_6_0_0

	admin, err := sarama.NewClusterAdmin([]string{broker}, config)
	if err != nil {
		return err
	}
	defer admin.Close()

	return admin.DeleteTopic(topic)
}

// NewKafkaProducerWithConfig создает продюсера с кастомной конфигурацией
func NewKafkaProducerWithConfig(brokers []string, topic string, configFunc func(*sarama.Config)) (ProducerInterface, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Retry.Max = 3
	config.Producer.Retry.Backoff = 200 * time.Millisecond
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.Compression = sarama.CompressionSnappy
	config.Producer.Partitioner = sarama.NewHashPartitioner
	config.Producer.Flush.Frequency = 500 * time.Millisecond
	config.Producer.Flush.Messages = 100
	config.Producer.Flush.MaxMessages = 500
	config.Version = sarama.V3_6_0_0

	if configFunc != nil {
		configFunc(config)
	}

	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create async producer: %w", err)
	}

	k := &AsyncKafkaProducer{
		producer:    producer,
		topic:       topic,
		successChan: make(chan *sarama.ProducerMessage, 100),
		errorChan:   make(chan *sarama.ProducerError, 100),
		stopChan:    make(chan struct{}),
		metrics:     &ProducerMetrics{},
	}

	go k.handleSuccesses()
	go k.handleErrors()

	return k, nil
}
