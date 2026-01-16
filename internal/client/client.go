package client

import (
	"context"
	"crypto/tls"
	"net"
	"sync"
	"time"

	"github.com/0xFilosoF/pow-ddos-guard/internal/client/handler"
	"github.com/0xFilosoF/pow-ddos-guard/internal/shared/pow"
	"github.com/0xFilosoF/pow-ddos-guard/pkg/config"
	"github.com/sourcegraph/jsonrpc2"
)

type Client struct {
	cfg    *config.Config[config.ClientParams]
	conn   net.Conn
	cancel context.CancelFunc
}

func New(cfg *config.Config[config.ClientParams]) *Client {
	return &Client{
		cfg: cfg,
	}
}

func (c *Client) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel

	conn, err := c.getConn(ctx)
	if err != nil {
		return err
	}
	c.conn = conn
	defer conn.Close()

	h := handler.New(ctx, pow.New(0, c.cfg.Params.PoW.SaltLen, 0, ""))
	stream := jsonrpc2.NewPlainObjectStream(conn)
	rpcConn := jsonrpc2.NewConn(ctx, stream, h)
	defer rpcConn.Close()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-rpcConn.DisconnectNotify():
		return nil
	}
}

func (c *Client) Shutdown(_ context.Context) error {
	errCh := make(chan error, 1)

	wg := sync.WaitGroup{}
	const workersCount = 1
	wg.Add(workersCount)

	go func() {
		defer wg.Done()
		c.cancel()
	}()

	wg.Wait()
	close(errCh)

	if err := <-errCh; err != nil {
		return err
	}

	return nil
}

func (c *Client) getConn(ctx context.Context) (net.Conn, error) {
	d := &net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	if c.cfg.TLS.Enabled {
		tlsCfg := &tls.Config{
			MinVersion: tls.VersionTLS13,
		}

		// conn, err := tls.Dial("tcp", c.cfg.App.Addr, tlsCfg)
		conn, err := tls.DialWithDialer(d, "tcp", c.cfg.App.Addr, tlsCfg)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}

	conn, err := d.DialContext(ctx, "tcp", c.cfg.App.Addr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
