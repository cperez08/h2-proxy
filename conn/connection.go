package conn

import (
	"fmt"
	"log"
	"net"

	"golang.org/x/net/http2"
)

// Connection represents the connection for an specific Address
type Connection struct {
	Address     string
	Conn        *http2.ClientConn
	IsConnected bool // used to init the connection after rehresing the ips
	IsActive    bool // indicates if the connection is active, can be deactivated/remmoved in case of multiple failures (TODO: circuit break)
}

// AddConnection adds a new connection to the pool
// if it is not duplicated
func AddConnection(pool *[]*Connection, con *Connection) {
	if con.Address == "" {
		return
	}

	for _, p := range *pool {
		if con.Address == p.Address {
			return
		}
	}

	*pool = append(*pool, con)
}

// Connect creates the actual connections available in the pool
// for those connections marked as active and not connected
func ConnectPool(t *http2.Transport, pool []*Connection) error {
	for _, p := range pool {
		if p.IsActive && !p.IsConnected {
			c, err := Connect(t, p.Address)
			if err != nil {
				return err
			}
			p.IsConnected = true
			p.Conn = c
		}
	}

	return nil
}

// Connect creates a new connection
func Connect(t *http2.Transport, host string) (*http2.ClientConn, error) {
	c, err := net.Dial("tcp", host)
	if err != nil {
		return nil, fmt.Errorf("[h2-proxy]: %w ", err)
	}

	h2conn, err := t.NewClientConn(c)
	if err != nil {
		return nil, fmt.Errorf("[h2-proxy]: %w", err)
	}

	return h2conn, nil
}

// RefreshConnections compares the refreshed IPs removing the non existing ones
// and creating the new ones
func RefreshConnections(pool *[]*Connection, refreshedAddrs []string) {
	var toRemove []string
	refreshedMap := make(map[string]uint8, len(refreshedAddrs))
	for _, ip := range refreshedAddrs {
		refreshedMap[ip] = 0
	}

	for _, p := range *pool {
		if _, ok := refreshedMap[p.Address]; ok {
			delete(refreshedMap, p.Address)
		} else {
			toRemove = append(toRemove, p.Address)
			delete(refreshedMap, p.Address)
		}
	}

	for _, r := range toRemove {
		removeConnection(pool, r)
	}

	for k := range refreshedMap {
		*pool = append(*pool, &Connection{Address: k, IsActive: true, IsConnected: false})
	}
}

// removeConnection removes a connection by Address
func removeConnection(pool *[]*Connection, Address string) {
	for i, c := range *pool {
		if c.Address == Address {
			(*pool)[i] = (*pool)[len(*pool)-1]
			*pool = (*pool)[:len(*pool)-1]
			break
		}
	}
}

// CloseAllConnections closes all connections in the pool
func CloseAllConnections(pool *[]*Connection) {
	for _, c := range *pool {
		if err := c.Conn.Close(); err != nil {
			log.Println("error closing connection ", err)
		}
	}

	*pool = (*pool)[:0]
}
