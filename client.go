package transaction_client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

var (
	ErrAlreadyBroadcast = errors.New("err_already_broadcast")
	ErrHashNotFound     = errors.New("err_hash_not_found")
	ErrNotSuccessStatus = errors.New("err_http_status")
	ErrPolling          = errors.New("err_polling")
)

// A Client makes HTTP request to BroadcastUrl or BasePollingUrl.
type Client struct {
	BroadcastUrl string

	BasePollingUrl string
	*PollingOptions
}

type PollingOptions struct {
	Interval time.Duration
	Handler  PollingHandler
}

var defaultPollingOptions = &PollingOptions{
	Interval: 5 * time.Second,
	Handler: func(s TransactionStatus) bool {
		return s == StatusPending
	},
}

// Handler receives the updated status
// and return a bool `true` to continue polling
// or `false` to stop.
//
// # Default PollingHandler:
//
//	func handler(t *Transaction) bool {
//	    return t.Status() == client.StatusPending
//	}
type PollingHandler func(TransactionStatus) bool

func NewClient(broadcastPath, basePollingPath string) *Client {
	return &Client{
		BroadcastUrl:   broadcastPath,
		BasePollingUrl: basePollingPath,
		PollingOptions: defaultPollingOptions,
	}
}

// Broadcast sends a POST request to the provided url with a trasaction as body.
// The endpoint should return a json with 'tx_hash'
//
// If err == nil, transaction hash and status will be set for polling
func (c *Client) Broadcast(tx *transaction) error {
	if tx.status != statusInit {
		return ErrAlreadyBroadcast
	}
	reqData, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	res, err := http.Post(c.BroadcastUrl, "application/json", bytes.NewReader(reqData))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		slog.Error("Error requesting", "url", c.BroadcastUrl, "response", res)
		return ErrNotSuccessStatus
	}

	resData := new(struct {
		Hash string `json:"tx_hash"`
	})
	err = json.NewDecoder(res.Body).Decode(resData)
	if err != nil {
		return err
	}

	if resData.Hash == "" {
		return ErrHashNotFound
	}

	tx.hash = resData.Hash
	tx.status = StatusPending
	return nil
}

func (c *Client) setPollingOptions() {
	if c.PollingOptions == nil {
		c.PollingOptions = defaultPollingOptions
		return
	}
	if c.PollingOptions.Interval == 0 {
		c.PollingOptions.Interval = defaultPollingOptions.Interval
	}
	if c.PollingOptions.Handler == nil {
		c.PollingOptions.Handler = defaultPollingOptions.Handler
	}
}

// Broadcast sends a POST request to the provided url. with a trasaction as body.
//
// If `PollingOptions.Handler` is not set, this function polls with default handler.
//
// Poll blocks the current thread execution.
//
// # Example usage:
//
//	exit := make(chan error, 1)
//	go func() {
//		exit <- client.Poll(ctx, tx)
//	}()
//	// do something
//	....
//	<- exit
func (c *Client) Poll(ctx context.Context, tx *transaction) (err error) {
	c.setPollingOptions()

	if tx.hash == "" {
		return ErrHashNotFound
	}
	url, err := url.JoinPath(c.BasePollingUrl, tx.hash)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(c.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			var status *TransactionStatus
			status, err = makeReq(url)
			if err != nil {
				slog.Error("Error requesting", "url", url, "error", err)
				return
			}
			tx.status = *status
			if !c.Handler(tx.status) {
				return
			}
		}
	}
}

func makeReq(url string) (*TransactionStatus, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		slog.Error("Error requesting", "url", url, "response", res)
		return nil, ErrNotSuccessStatus
	}

	resData := new(struct {
		Status TransactionStatus `json:"tx_status"`
	})
	err = json.NewDecoder(res.Body).Decode(resData)
	if err != nil {
		return nil, err
	}

	return &resData.Status, nil
}
