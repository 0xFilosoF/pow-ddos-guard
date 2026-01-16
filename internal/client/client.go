package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
	"os"
	"sync"

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
	go h.GenerateRequestsAsync(ctx, rpcConn)
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
	netDialer := &net.Dialer{
		Timeout: c.cfg.Params.Timeout,
	}

	if c.cfg.TLS.Enabled {
		caPem, _ := os.ReadFile(c.cfg.TLS.Cert)
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(caPem)

		tlsDialer := &tls.Dialer{
			NetDialer: netDialer,
			Config: &tls.Config{
				RootCAs:    pool,
				MinVersion: tls.VersionTLS13,
				ServerName: "server",
			},
		}

		return tlsDialer.DialContext(ctx, "tcp", c.cfg.App.Addr)
	}

	return netDialer.DialContext(ctx, "tcp", c.cfg.App.Addr)
}
