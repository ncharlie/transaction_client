package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ncharlie/transaction_client"
)

func main() {
	c := &transaction_client.Client{
		BroadcastUrl:   "https://mock-node-wgqbnxruha-as.a.run.app/broadcast",
		BasePollingUrl: "https://mock-node-wgqbnxruha-as.a.run.app/check",
		PollingOptions: &transaction_client.PollingOptions{
			Interval: time.Second,
		},
	}

	tx, err := transaction_client.NewTransaction("ETH", 4500, uint64(time.Now().Unix()))
	if err != nil {
		panic(err)
	}
	fmt.Println(tx)
	// result: &{ETH 4500 1709738070  INIT}

	err = c.Broadcast(tx)
	if err != nil {
		panic(err)
	}
	fmt.Println(tx)
	// result: &{ETH 4500 1709738070 a022978a4cbd6e14e21928e29a756bad04f8e9be93bba91a67e4485e4e9a5875 PENDING}

	err = c.Poll(context.Background(), tx)
	if err != nil {
		panic(err)
	}
	fmt.Println(tx)
	// result: &{ETH 4500 1709738070 a022978a4cbd6e14e21928e29a756bad04f8e9be93bba91a67e4485e4e9a5875 CONFIRMED}
}
