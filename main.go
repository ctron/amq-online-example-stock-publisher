package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/namsral/flag"

	"log"
	"time"

	"pack.ag/amqp"
)

var insecure = false
var uri = ""
var stock = ""

var last = Stock{
	Price: 100.0,
	Trend: 0,
}

func createTlsConfig() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: insecure,
	}
}

type Stock struct {
	Price float64 `json:"price"`
	Trend int     `json:"trend"`
}

func publish() error {

	opts := make([]amqp.ConnOption, 0)
	if insecure {
		opts = append(opts, amqp.ConnTLSConfig(createTlsConfig()))
	}

	// create new connection

	client, err := amqp.Dial(uri, opts...)
	if err != nil {
		return err
	}

	defer func() {
		if err := client.Close(); err != nil {
			log.Fatal("Failed to close client:", err)
		}
	}()

	var ctx = context.Background()

	// create new session

	session, err := client.NewSession()
	if err != nil {
		return err
	}

	defer func() {
		if err := session.Close(ctx); err != nil {
			log.Fatal("Failed to close session:", err)
		}
	}()

	// create new sender

	sender, err := session.NewSender(
		amqp.LinkTargetAddress("stock/" + stock),
	)
	if err != nil {
		return err
	}

	defer func() {
		if err := sender.Close(ctx); err != nil {
			log.Fatal("Failed to close sender: ", err)
		}
	}()

	for {

		last := createNext()

		data, err := json.Marshal(&last)
		if err != nil {
			return err
		}

		msg := amqp.NewMessage(data)
		if err := sender.Send(ctx, msg); err != nil {
			return err
		}

		time.Sleep(time.Second * 5)

	}
}

func createNext() Stock {
	a := rand.Float64() - 0.5
	p := last.Price + a
	if p < 0 {
		p = 0
	}

	t := p - last.Price
	if t > 0.0 {
		t = 1
	} else if t < 0.0 {
		t = -1
	} else {
		t = 0
	}

	return Stock{
		Price: p,
		Trend: int(t),
	}
}

func main() {

	flag.BoolVar(&insecure, "insecure", false, "disable all TLS validation")
	flag.StringVar(&uri, "uri", "", "AMQP endpoint URL")
	flag.StringVar(&stock, "stock", "RHAT", "AMQP endpoint URL")

	flag.Parse()

	fmt.Printf("URI: %s, Insecure TLS: %v, Stock: %s\n", uri, insecure, stock)

	if err := publish(); err != nil {
		fmt.Println("Failed to run: ", err)
	}
}
