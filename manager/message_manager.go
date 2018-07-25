package manager

import (
	"flag"
	"os"

	"github.com/Shopify/sarama"
)

// ProducerManager interface
type ProducerManager interface {
	Init() error
}

// SaramaProducerManager structure
type SaramaProducerManager struct {
	Broker   string
	Config   *sarama.Config
	Producer sarama.SyncProducer
}

// Init function implementation
func (m *SaramaProducerManager) Init() error {

	var err error

	flag.StringVar(&m.Broker, "brokers.addr", "localhost:9092", "Broker server addr")

	// Use environment variables, if set. Flags have priority over Env vars.
	if broker := os.Getenv("BROKER_SERVER_ADDR"); broker != "" {
		m.Broker = broker
	}

	m.Config = sarama.NewConfig()
	m.Config.Producer.RequiredAcks = sarama.WaitForAll
	m.Config.Producer.Retry.Max = 5
	m.Config.Producer.Return.Successes = true
	m.Config.Producer.Return.Errors = true

	// Instanciate the producer
	brokers := []string{m.Broker}
	m.Producer, err = sarama.NewSyncProducer(brokers, m.Config)

	defer func() {
		if err := m.Producer.Close(); err != nil {
			panic(err)
		}
	}()

	if err != nil {
		return err
	}

	return nil
}
