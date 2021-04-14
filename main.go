package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/appcelerator/chimera-client-go/chimera"
)

var counter int

var destination string
var authToken string
var query string
var queueName string
var lifeTime int

func init() {
	flag.StringVar(&destination, "d", "chimera.platform.axway.com", "Chimera host name")
	flag.StringVar(&authToken, "t", "", "Auth token for Chimera")
	flag.StringVar(&queueName, "qn", "traceability_agent", "Chimera queue name")
	flag.StringVar(&query, "q", `{"$match": {"content.type": "transactionSummary"}}`, "Chimera filter query")
	flag.IntVar(&lifeTime, "lt", 15000, "Chimera queue lifetime")
}

func main() {
	flag.Parse()
	log.Printf("Chimera host: %s\n", destination)
	if authToken == "" {
		log.Println("Failed to create client: chimera auth token not configured")
		flag.Usage()
		os.Exit(1)
	}
	if queueName != "" {
		log.Println("Queue name: " + queueName)
	} else {
		log.Println("Subscribing to auto generated queue")
	}
	log.Printf("Queue lifetime: %d\n", lifeTime)
	log.Println("Query filter: " + query)

	ctx, done := context.WithCancel(context.Background())
	client, err := chimera.NewClient(ctx, chimera.ClientOptions{Protocol: chimera.HTTPS, Host: destination, AuthKey: authToken})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	go func() {
		subscribe := chimera.Subscribe{
			Binding:  []byte(query),
			LifeTime: lifeTime,
		}
		if queueName != "" {
			subscribe.Options = &chimera.SubscriptionOptions{
				Queue: &chimera.SubscriptionQueueOptions{
					Name: queueName,
				},
			}
		}

		subOpts := chimera.SubscribeOptions{
			Resubscribe: 30,
		}

		client.Subscribe(ctx, subscribe, subOpts, write(os.Stdout))
		done()
	}()

	go func() {
		timeDelay := 30 * time.Second // == 900000 * time.Microsecond

		endTime := time.After(timeDelay)

		for {
			select {
			case <-endTime:
				endTime = time.After(timeDelay)
			default:
				time.Sleep(1 * time.Second)
				continue
			}

			fmt.Printf("Current count %d\n", counter)
		}
	}()

	<-ctx.Done()
}

func write(w io.Writer) chan<- []byte {
	lines := make(chan []byte)
	go func() {
		for line := range lines {
			fmt.Fprintln(w, string(line))
			counter++
		}
	}()
	return lines
}
