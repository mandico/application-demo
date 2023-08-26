package main

import (
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	_ "github.com/go-sql-driver/mysql"
)

const (
	host     = "azr-mysql-demo.mysql.database.azure.com"
	database = "demo"
	user     = "demo"
	password = "P@ssw0rd1234"
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

func RecordRegistry(db *sql.DB, msg string, msg_partition int32, offset int32) {
	// Insert some data into table.
	sqlStatement, err := db.Prepare("INSERT INTO message (msg, msg_partition, msg_offset) VALUES (?,?,?);")
	res, err := sqlStatement.Exec(msg, msg_partition, offset)

	checkError(err)
	rowCount, err := res.RowsAffected()
	fmt.Printf("RECORDED ORDER - Inserted %d row(s) of data.\n", rowCount)
}

func main() {

	c := connectionEventHub()

	// DATABASE
	// Initialize connection string.
	var connectionString = fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?allowNativePasswords=true", user, password, host, database)

	// Initialize connection object.
	db, err := sql.Open("mysql", connectionString)
	checkError(err)
	defer db.Close()

	err = db.Ping()
	checkError(err)
	fmt.Println("Successfully created connection to database.")
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
				RecordRegistry(db, string(e.Value), e.TopicPartition.Partition, int32(e.TopicPartition.Offset))

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
