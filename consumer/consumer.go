package main

import (
	"log"
	"sync"

	"github.com/nsqio/go-nsq"
)

func main() {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	decodeConfig := nsq.NewConfig()
	c, err := nsq.NewConsumer("write_test", "ch", decodeConfig)
	if err != nil {
		log.Panic("Could not create consumer")
	}

	c.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
		log.Println("NSQ message received:")
		log.Println(string(message.Body))
		return nil
	}))

	err = c.ConnectToNSQD("devel-go.tkpd:4150")
	if err != nil {
		log.Panic("Could not connect")
	}
	log.Println("Awaiting visitor to come...")
	wg.Wait()
}
