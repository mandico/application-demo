package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

const (
	endpoint      = "https://azr-cosmos-db-demo-lab.documents.azure.com:443/"
	key           = "piDfygzZiN67AbVs9etAu4MAaLICMoVxLwBjD6LyfLVpBGvkyXn2Uh6EQxOlFfiVGJGP5Uw8QZ9zACDbBO4iug=="
	databaseName  = "database_demo"
	containerName = "order"
	partitionKey  = "/order"
)

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func connectionEventHub() *kafka.Consumer {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Usage: %s <bootstrap-servers> <group> <topics..>\n",
			os.Args[0])
		os.Exit(1)
	}

	bootstrapServers := os.Args[1]
	group := os.Args[2]
	topics := os.Args[3:]
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":        bootstrapServers,
		"broker.address.family":    "v4",
		"group.id":                 group,
		"session.timeout.ms":       6000,
		"auto.offset.reset":        "earliest",
		"enable.auto.offset.store": false,
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create consumer: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created Consumer %v\n", c)
	err = c.SubscribeTopics(topics, nil)
	return c
}

func RecordRegistry(client *azcosmos.Client, msg string, msg_partition int32, offset int32) {
	item := struct {
		ID string `json:"id"`
		//OrderId      string `json:"orderId"`
		Msg          string
		MsgPartition int32
		MsgOffset    int32
	}{
		ID: "1",
		//OrderId:      "1",
		Msg:          msg,
		MsgPartition: msg_partition,
		MsgOffset:    offset,
	}

	err := createItem(client, databaseName, containerName, item.OrderId, item)
	if err != nil {
		log.Printf("createItem failed: %s\n", err)
	}
}

func createItem(client *azcosmos.Client, databaseName, containerName, partitionKey string, item any) error {
	//	databaseName = "adventureworks"
	//	containerName = "customer"
	//	partitionKey = "1"

	/*	item = struct {
			ID           string `json:"id"`
			CustomerId   string `json:"customerId"`
			Title        string
			FirstName    string
			LastName     string
			EmailAddress string
			PhoneNumber  string
			CreationDate string
		}{
			ID:           "1",
			CustomerId:   "1",
			Title:        "Mr",
			FirstName:    "Luke",
			LastName:     "Hayes",
			EmailAddress: "luke12@adventure-works.com",
			PhoneNumber:  "879-555-0197",
			CreationDate: "2014-02-25T00:00:00",
		}
	*/
	// create container client
	containerClient, err := client.NewContainer(databaseName, containerName)
	if err != nil {
		return fmt.Errorf("failed to create a container client: %s", err)
	}

	// specifies the value of the partiton key
	pk := azcosmos.NewPartitionKeyString(partitionKey)

	b, err := json.Marshal(item)
	if err != nil {
		return err
	}
	// setting the item options upon creating ie. consistency level
	itemOptions := azcosmos.ItemOptions{
		ConsistencyLevel: azcosmos.ConsistencyLevelSession.ToPtr(),
	}

	// this is a helper function that swallows 409 errors
	errorIs409 := func(err error) bool {
		var responseErr *azcore.ResponseError
		return err != nil && errors.As(err, &responseErr) && responseErr.StatusCode == 409
	}

	ctx := context.TODO()
	itemResponse, err := containerClient.CreateItem(ctx, pk, b, &itemOptions)

	switch {
	case errorIs409(err):
		log.Printf("Item with partitionkey value %s already exists\n", pk)
	case err != nil:
		return err
	default:
		log.Printf("Status %d. Item %v created. ActivityId %s. Consuming %v Request Units.\n", itemResponse.RawResponse.StatusCode, pk, itemResponse.ActivityID, itemResponse.RequestCharge)
	}

	return nil
}

func main() {

	c := connectionEventHub()

	// DATABASE
	cred, err := azcosmos.NewKeyCredential(key)
	if err != nil {
		log.Fatal("Failed to create a credential: ", err)
	}

	// Create a CosmosDB client
	client, err := azcosmos.NewClientWithKey(endpoint, cred, nil)
	if err != nil {
		log.Fatal("Failed to create Azure Cosmos DB db client: ", err)
	}
	// DATABASE

	run := true

	for run {
		select {
		default:
			ev := c.Poll(100)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				fmt.Printf("%% RECEIVE :: TOPIC: %s :: PARTITION: %d :: OFFSET: %d :: MSG: %s\n",
					*e.TopicPartition.Topic, e.TopicPartition.Partition, e.TopicPartition.Offset, string(e.Value))
				RecordRegistry(client, string(e.Value), e.TopicPartition.Partition, int32(e.TopicPartition.Offset))

				if e.Headers != nil {
					fmt.Printf("%% Headers: %v\n", e.Headers)
				}
				_, err := c.StoreMessage(e)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%% Error storing offset after message %s:\n",
						e.TopicPartition)
				}

			case kafka.Error:
				fmt.Fprintf(os.Stderr, "%% Error: %v: %v\n", e.Code(), e)
				if e.Code() == kafka.ErrAllBrokersDown {
					run = false
				}
			default:
				fmt.Printf("Ignored %v\n", e)
			}
		}
	}
	fmt.Printf("Closing consumer\n")
	c.Close()
}
