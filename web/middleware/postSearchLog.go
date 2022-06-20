package middleware

import (
	"encoding/json"
	"github.com/Shopify/sarama"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"log"
	"strconv"
	"time"
)

func PostBySync(producer *SingletonProducer) gin.HandlerFunc {
	return func(c *gin.Context) {
		t := strconv.FormatInt(time.Now().Unix(), 10)
		c.Next()
		var sle = &searchLogEntry{
			Time:           t,
			Identification: c.ClientIP(),
		}
		if err := c.ShouldBindBodyWith(sle, binding.JSON); err != nil {
			log.Printf("err %v", err)
		}

		//log.Printf("Kafka brokers: %s", strings.Join(producer.BrokerList, ", "))
		log.Println(producer.syncProducer)
		partition, offset, err := producer.GetSyncProducer().SendMessage(&sarama.ProducerMessage{
			Topic: producer.Topic,
			Value: sle,
		})
		if err != nil {
			log.Printf("err!! %v", err)
		} else {
			log.Printf("Your data is stored with unique identifier important/%d/%d", partition, offset)
		}
	}
}
func PostByAsync(producer *SingletonProducer) gin.HandlerFunc {
	return func(c *gin.Context) {
		t := strconv.FormatInt(time.Now().Unix(), 10)
		c.Next()
		var sle = &searchLogEntry{
			Time:           t,
			Identification: c.ClientIP(),
		}
		if err := c.ShouldBind(sle); err != nil {
			log.Printf("paras bind err: %v", err)
		}
		producer.GetAsyncProducer().Input() <- &sarama.ProducerMessage{
			Topic: producer.Topic,
			Key:   sarama.StringEncoder(sle.Identification),
			Value: sle,
		}
	}

}

type SingletonProducer struct {
	syncProducer  sarama.SyncProducer
	asyncProducer sarama.AsyncProducer
	BrokerList    []string
	Topic         string
}

func (sp *SingletonProducer) GetSyncProducer() sarama.SyncProducer {
	if sp.syncProducer == nil {
		sp.syncProducer = sp.newSyncProducer()
	}
	return sp.syncProducer
}
func (sp *SingletonProducer) GetAsyncProducer() sarama.AsyncProducer {
	if sp.asyncProducer == nil {
		sp.asyncProducer = sp.newAsyncProducer()
	}
	return sp.asyncProducer
}

func (sp *SingletonProducer) newSyncProducer() sarama.SyncProducer {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll // Wait for all in-sync replicas to ack the message
	config.Producer.Retry.Max = 10                   // Retry up to 10 times to produce the message
	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(sp.BrokerList, config)
	if err != nil {
		log.Fatalln("Failed to start Sarama producer:", err)
	}
	return producer
}
func (sp *SingletonProducer) newAsyncProducer() sarama.AsyncProducer {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal       // Only wait for the leader to ack
	config.Producer.Compression = sarama.CompressionSnappy   // Compress messages
	config.Producer.Flush.Frequency = 500 * time.Millisecond // Flush batches every 500ms
	producer, err := sarama.NewAsyncProducer(sp.BrokerList, config)
	if err != nil {
		log.Fatalln("Failed to start Sarama producer:", err)
	}
	go func() {
		for err := range producer.Errors() {
			log.Println("Failed to write search log entry:", err)
		}
	}()
	return producer
}

type searchLogEntry struct {
	Query          string `json:"query,omitempty"`
	Identification string `json:"identification,omitempty"`
	Time           string `json:"time,omitempty"`
	encoded        []byte
	err            error
}

func (sle *searchLogEntry) ensureEncoded() {
	if sle.encoded == nil && sle.err == nil {
		sle.encoded, sle.err = json.Marshal(sle)
	}
}

func (sle *searchLogEntry) Length() int {
	sle.ensureEncoded()
	return len(sle.encoded)
}

func (sle *searchLogEntry) Encode() ([]byte, error) {
	sle.ensureEncoded()
	return sle.encoded, sle.err
}
