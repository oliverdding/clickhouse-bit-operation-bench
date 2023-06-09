package cmd

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

var (
	brokers       = ""
	topic         = ""
	producers     = 1
	recordsNumber = 1
	loops         = 1
)

type Field struct {
	Category  string
	UserId    string
	FuncCrash uint64
	FuncLag   uint64
}

func init() {
	flag.StringVar(&brokers, "b", "127.0.0.1:9092", "Kafka bootstrap brokers to connect to, as a comma separated list")
	flag.StringVar(&topic, "t", "test", "Kafka topics where records will be copied from topics.")
	flag.IntVar(&producers, "p", 5, "Number of concurrent producers")
	flag.IntVar(&recordsNumber, "n", 100, "Number of records sent per loop")
	flag.IntVar(&loops, "l", 1000, "Times of loop")
	flag.Parse()

	if len(brokers) == 0 {
		panic("no Kafka bootstrap brokers defined, please set the -brokers flag")
	}

	if len(topic) == 0 {
		panic("no topic given to be consumed, please set the -topic flag")
	}
}

func main() {
	var wg sync.WaitGroup
	wg.Add(producers)

	for producer := 0; producer < producers; producer++ { // start producers
		go func(producer int) {
			log.Printf("producer %d start", producer)
			writer := kafka.Writer{
				Addr:                   kafka.TCP(strings.Split(brokers, ",")...),
				Topic:                  topic,
				Balancer:               &kafka.LeastBytes{},
				BatchTimeout:           10 * time.Millisecond,
				AllowAutoTopicCreation: true,
			}
			defer writer.Close()

			// prepare write messages
			messages := make([]kafka.Message, 100*10)
			for i := 0; i < recordsNumber; i++ {

				user_id := uuid.New()
				func_crash := uint64(rand.Int63() % 2)
				func_lag := uint64(rand.Int63() % 2)

				n := (int(rand.Int31()) % 5) + 1 // [1, 5]
				for j := 0; j < n; j++ {
					bytes, err := json.Marshal(&Field{
						Category:  "PERF_LAUNCH",
						UserId:    user_id.String(),
						FuncCrash: func_crash,
						FuncLag:   func_lag,
					})
					if err != nil {
						log.Fatal("failed to marshal message:", err)
					}
					messages[i] = kafka.Message{
						Value: bytes,
					}
				}

				if func_crash == 1 {
					for j := 0; j < n; j++ {
						bytes, err := json.Marshal(&Field{
							Category:  "PERF_CRASH",
							UserId:    user_id.String(),
							FuncCrash: 0,
							FuncLag:   0,
						})
						if err != nil {
							log.Fatal("failed to marshal message:", err)
						}
						messages[i] = kafka.Message{
							Value: bytes,
						}
					}
				}

				if func_lag == 1 {
					for j := 0; j < int(rand.Int63()%10); j++ {
						bytes, err := json.Marshal(&Field{
							Category:  "PERF_LAG",
							UserId:    user_id.String(),
							FuncCrash: 0,
							FuncLag:   0,
						})
						if err != nil {
							log.Fatal("failed to marshal message:", err)
						}
						messages[i] = kafka.Message{
							Value: bytes,
						}
					}
				}

			}

			ticker := time.NewTicker(time.Millisecond * 100)
			defer ticker.Stop()

			times := 0
			for range ticker.C {
				log.Printf("producer %d will start writing", producer)

				// write to kafka
				if err := writer.WriteMessages(context.Background(), messages...); err != nil {
					log.Fatal("failed to write messages:", err)
				}

				times += 1
				if times >= loops {
					break
				}
			}
			log.Printf("producer %d will stop", producer)
			wg.Done()
		}(producer)
	}

	wg.Wait()
	log.Printf("Exit...")
}
