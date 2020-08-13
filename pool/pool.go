package pool

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/cperez08/h2-proxy/config"
	"github.com/cperez08/h2-proxy/conn"
	"github.com/cperez08/h2-proxy/lb"
	"github.com/cperez08/h2-proxy/resolver"
	"golang.org/x/net/http2"
)

// connectionPool is the implementation for http2.ConnPool interface
type connectionPool struct {
	ctx           context.Context
	t             *http2.Transport
	m             sync.Mutex
	connections   []*conn.Connection
	balancer      lb.LoadBalancer
	r             *resolver.Resolver
	basePort      string
	isDomainBased bool // indicates if the pool was built based on a domain with multiple A / AAA records or 1 or N IPs
}

// NewConnectionPool returns a new instance of the connectionPool object
// also initializes the set of connections based on the Address
func NewConnectionPool(ctx context.Context, cfg *config.ProxyConfig, t *http2.Transport) (http2.ClientConnPool, error) {
	c := &connectionPool{t: t, basePort: cfg.TargetPort, ctx: ctx}
	if ip := net.ParseIP(cfg.TargetHost); ip != nil {
		c.balancer = lb.GetBalancer(lb.None)
		c.connections = append(c.connections, &conn.Connection{Address: cfg.TargetHost + ":" + cfg.TargetPort, IsConnected: false, IsActive: true})
		if err := c.initPool(); err != nil {
			return nil, err
		}

		return c, nil
	}

	c.balancer = lb.GetBalancer(lb.Balancer(cfg.DNSConfig.BalancerAlg))
	c.r = resolver.NewResolver(cfg.DNSConfig.RefreshRate, cfg.DNSConfig.NeedRefresh)
	c.isDomainBased = true

	ips := c.r.Resolve(cfg.TargetHost, cfg.TargetPort)
	for _, i := range ips {
		conn.AddConnection(&c.connections, &conn.Connection{Address: i, IsConnected: false, IsActive: true})
	}

	if err := c.initPool(); err != nil {
		return nil, err
	}

	go c.watchForChanges()
	return c, nil
}

// GetClientConn returns a new connection
func (p *connectionPool) GetClientConn(req *http.Request, addr string) (*http2.ClientConn, error) {
	// TODO handle close request after response
	// if req.Close {
	// 	return conn.Connect(p.t, addr)
	// }

	p.m.Lock()
	defer p.m.Unlock()
	if len(p.connections) == 0 {
		return nil, errors.New("no active connections found")
	}

	c := p.balancer.PickConnection(p.connections)
	if c == nil {
		return nil, errors.New("no active connections found")
	}

	return c.Conn, nil
}

// MarkDead mark a connection as dead removing it from the pool
func (p *connectionPool) MarkDead(cc *http2.ClientConn) {
	p.m.Lock()
	defer p.m.Unlock()
	for i, c := range p.connections {
		if c.Conn == cc {
			p.connections[i] = p.connections[len(p.connections)-1]
			p.connections = p.connections[:len(p.connections)-1]
			break
		}
	}

	if err := cc.Close(); err != nil {
		log.Println("error closing dead connection ")
	}

	p.balancer.RebuildBalancer(p.connections)
}

func (p *connectionPool) initPool() error {
	return conn.ConnectPool(p.t, p.connections)
}

// TODO add upper context to handle cancelation
func (p *connectionPool) watchForChanges() {
	for {
		select {
		case <-p.ctx.Done():
			conn.CloseAllConnections(&p.connections)
			p.r.CloseResolver()
			return
		case <-p.r.C:
			p.refreshConnections(p.r.GetCurrentIPs())
		}
	}
}

func (p *connectionPool) refreshConnections(refreshedIPs []string) {
	p.m.Lock()
	defer p.m.Unlock()

	conn.RefreshConnections(&p.connections, refreshedIPs)

	// let's create the connections for the new ips
	if err := conn.ConnectPool(p.t, p.connections); err != nil {
		log.Println("error refreshing connection ", err)
	}

	p.balancer.RebuildBalancer(p.connections)
}

// func traceGetConn(req *http.Request, hostPort string) {
// 	trace := httptrace.ContextClientTrace(req.Context())
// 	if trace == nil || trace.GetConn == nil {
// 		return
// 	}
// 	trace.GetConn(hostPort)
// }
