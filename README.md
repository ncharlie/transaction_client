# Transaction Broadcasting and Monitoring Client

transaction_client package provides a structured way to broadcast and monitor transactions.

```sh
go get github.com/ncharlie/transaction_client
```

## Quickstart

```go
package main

import "github.com/ncharlie/transaction_client"

func main() {
    client := transaction_client.NewClient("broadcast_url", "poll_base_url")

    tx, _ := transaction_client.NewTransaction("ETH", 4500, uint64(time.Now().Unix()))

    c.Broadcast(tx)

    c.Poll(context.Background(), tx)

    fmt.Println(tx)	// result: &{ETH 4500 1709738070 a022978a4cbd6e14e21928e29a756bad04f8e9be93bba91a67e4485e4e9a5875 CONFIRMED}
}
```

## Design

- The `transaction` struct contains transaction data and status.

- The `transaction` struct keeps the `status` field private. This ensure that users follow the business flow by calling `Broadcast` before `Poll`.

- The `client` struct contains the endpoints for broadcasting and status monitoring. It also accepts configurations for polling.

- The client accepts the urls from the user. This allows setting different paths for different environment.

- Allowing users to pass in a polling handler allows them to decide the strategies for handling each transaction status on their own.

- I assume that `BroadcastUrl` endpoint will return JSON response with a `tx_hash` field and `BasePollingUrl` endpoint will return JSON response with `tx_status` field. So the response data structure is embedded in the lib.

- The code returns all the error to the caller and consequently, the caller must also provide error handling logic.

## Preparing struct

### Creating a client

The central point is the `transaction_client.Client`. A client manages requests to the `BroadcastUrl` and `BasePollingUrl` which should be set at the time of creation.

For simplicity, the request and response structure for these endpoints are fixed within the lib.

```go
client := &transaction_client.Client{
    BroadcastUrl:   "https://your-endpoint-to/broadcast",
    BasePollingUrl: "https://your-endpoint-to/check",
}
```

The `client` can now be used to broadcast transaction by calling the `Broadcast` method (see below for more information).

### Creating a transaction

The `NewTransaction` validates input data and returns a `transaction` object to be used with client. New transaction begins with `INIT` status.

```go
tx, err := transaction_client.NewTransaction("ETH", 7000, uint64(time.Now().Unix()))
// ... error handling
fmt.Println(tx) // &{ETH 4500 1709738070  INIT}
```

## Using client

### Broadcasting

As mentioned above, broadcasting can be made by providing a transaction in `INIT` state to the client. If no error occurs, the client sets the `hash` value of the transaction from the response and changes the transaction status to `PENDING`.

```go
err = c.Broadcast(tx)
// ... error handling
fmt.Println(tx) // &{ETH 4500 1709738070 a022978a4cbd6e14e21928e29a756bad04f8e9be93bba91a67e4485e4e9a5875 PENDING}
```

### Monitoring Status

After broadcasting, we can monitor the transaction while the chain is processing it by keep polling the `BasePollingUrl` and read the status.

The `transaction_client.PollingOptions` struct passed to the client contains 2 important configuration options for making the HTTP request.

- Interval option telling the client to make request at every interval tick. Default value is 5 seconds.
- Handler option accept a `transaction_client.PollingHandler`. We can decide what to do for each status response in this function. Returning `false` stop the polling. The default handler sets to keep polling until the transaction status is not `PENDING`.

```go
c.PollingOptions = &transaction_client.PollingOptions{
    Interval: 5 * time.Second,
    Handler: func(ts transaction_client.TransactionStatus) bool {
        return ts == transaction_client.StatusPending
    },
}
```

Start polling by:

```go
err = c.Poll(context.Background(), tx)
// ... error handling
```

Calling `Poll` blocks the execution thread until the `PollingHandler` return `false` so maybe we would want to call it in another goroutine. This way, we can cancel the polling by calling the `context.CancleFunc`.

```go
exit := make(chan error, 1)
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
go func() {
    exit <- c.Poll(ctx, tx)
}()

// do something else

// wait
err = <-exit
// ... error handling
fmt.Println(tx) // &{ETH 4500 1709738070 a022978a4cbd6e14e21928e29a756bad04f8e9be93bba91a67e4485e4e9a5875 CONFIRMED}
```

## Example

See the [example](example/main.go).
