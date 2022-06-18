package middleware

import (
	"encoding/json"
	"github.com/Shopify/sarama"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"log"
	"strconv"
	"strings"
	"time"
)

var (
	brokers = []string{"127.0.0.1:9092"}
	Topic   = "Test"
)

func PostBySync() gin.HandlerFunc {
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
		log.Printf("Kafka brokers: %s", strings.Join(brokers, ", "))
		producer := newSyncProducer(brokers)
		defer producer.Close()
		partition, offset, err := producer.SendMessage(&sarama.ProducerMessage{
			Topic: Topic,
			Value: sle,
		})
		if err != nil {
			log.Printf("err!! %v", err)
		} else {
			log.Printf("Your data is stored with unique identifier important/%d/%d", partition, offset)
		}
	}
}
func PostByAsync() gin.HandlerFunc {
	return func(c *gin.Context) {
		//	处理逻辑
		log.Printf("Kafka brokers: %s", strings.Join(brokers, ", "))
		//	如果是第一次，初始化

		//	把搜索请求异步发送给kafka
		var sle = &searchLogEntry{
			Time:           strconv.FormatInt(time.Now().Unix(), 10),
			Identification: c.ClientIP(),
		}
		if err := c.ShouldBind(sle); err != nil {
			log.Printf("paras bind err: %v", err)
		}
		//new producer?
		producer := newAsyncProducer(brokers)
		//producer
		producer.Input() <- &sarama.ProducerMessage{
			Topic: Topic,
			Key:   sarama.StringEncoder(sle.Identification),
			Value: sle,
		}
		c.Next()
	}

}

func newAsyncProducer(brokerList []string) sarama.AsyncProducer {
	// For the search log, we are looking for AP semantics, with high throughput.
	// By creating batches of compressed messages, we reduce network I/O at a cost of more latency.
	config := sarama.NewConfig()
	//tlsConfig := createTlsConfiguration()
	//if tlsConfig != nil {
	//	config.Net.TLS.Enable = true
	//	config.Net.TLS.Config = tlsConfig
	//}
	config.Producer.RequiredAcks = sarama.WaitForLocal       // Only wait for the leader to ack
	config.Producer.Compression = sarama.CompressionSnappy   // Compress messages
	config.Producer.Flush.Frequency = 500 * time.Millisecond // Flush batches every 500ms

	producer, err := sarama.NewAsyncProducer(brokerList, config)
	if err != nil {
		log.Fatalln("Failed to start Sarama producer:", err)
	}
	// We will just log to STDOUT if we're not able to produce messages.
	// Note: messages will only be returned here after all retry attempts are exhausted.
	go func() {
		for err := range producer.Errors() {
			log.Println("Failed to write search log entry:", err)
		}
	}()
	return producer
}

func newSyncProducer(brokerList []string) sarama.SyncProducer {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll // Wait for all in-sync replicas to ack the message
	config.Producer.Retry.Max = 10                   // Retry up to 10 times to produce the message
	config.Producer.Return.Successes = true
	//tlsConfig := createTlsConfiguration()
	//if tlsConfig != nil {
	//	config.Net.TLS.Config = tlsConfig
	//	config.Net.TLS.Enable = true
	//}

	// On the broker side, you may want to change the following settings to get
	// stronger consistency guarantees:
	// - For your broker, set `unclean.leader.election.enable` to false
	// - For the topic, you could increase `min.insync.replicas`.

	producer, err := sarama.NewSyncProducer(brokerList, config)
	if err != nil {
		log.Fatalln("Failed to start Sarama producer:", err)
	}

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
