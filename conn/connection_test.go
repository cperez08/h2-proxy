package conn

import (
	"crypto/tls"
	"log"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/http2"
)

func TestAddConnection(t *testing.T) {
	pool := []*Connection{}
	con := &Connection{Address: "localhost:8080"}
	AddConnection(&pool, con)
	assert.Equal(t, 1, len(pool))

	AddConnection(&pool, &Connection{})
	assert.Equal(t, 1, len(pool))

	AddConnection(&pool, &Connection{Address: "localhost:8080"})
	assert.Equal(t, 1, len(pool))

	AddConnection(&pool, &Connection{Address: "localhost:8090"})
	assert.Equal(t, 2, len(pool))
}

func TestConnectPool(t *testing.T) {
	tr := getTransport()
	l := fakeListener("8080")
	defer l.Close()
	pool := []*Connection{}
	con := &Connection{Address: "localhost:8080", IsActive: true}

	AddConnection(&pool, con)
	err := ConnectPool(tr, pool)
	if err != nil {
		t.Log("error connecting to the host")
		t.Fail()
	}

	if !con.IsConnected {
		t.Log("error creating connection")
		t.Fail()
	}

	AddConnection(&pool, &Connection{Address: "localhost:8090", IsActive: true})
	if err := ConnectPool(tr, pool); err == nil {
		t.Log("should not get connected")
		t.Fail()
	}

	pool = []*Connection{}
	con = &Connection{Address: "localhost:8080", IsActive: false}
	AddConnection(&pool, con)
	if err = ConnectPool(tr, pool); err != nil {
		t.Log("error connecting to the host")
		t.Fail()
	}

	if con.IsConnected {
		t.Log("should not connect innactive connection")
		t.Fail()
	}

	AddConnection(&pool, &Connection{Address: "localhost:8080", IsActive: true, IsConnected: true})
	if err = ConnectPool(tr, pool); err != nil {
		t.Log("should not fail since all connections are active and markes as connected")
		t.Fail()
	}
}

func TestConnect(t *testing.T) {
	tr := getTransport()
	l := fakeListener("8080")
	defer l.Close()

	c, err := Connect(tr, "localhost:8080")
	if c == nil || err != nil {
		t.Log("error connecting")
		t.Fail()
	}

	c, err = Connect(tr, "localhost:8090")
	if c != nil || err == nil {
		t.Log("connection should fail")
		t.Fail()
	}
}

func TestRefreshConnections(t *testing.T) {
	tr := getTransport()
	l := fakeListener("8080")
	l2 := fakeListener("8090")
	defer l.Close()
	defer l2.Close()

	pool := []*Connection{}
	con := &Connection{Address: "localhost:8080"}
	con2 := &Connection{Address: "localhost:8090"}
	AddConnection(&pool, con)
	AddConnection(&pool, con2)

	if err := ConnectPool(tr, pool); err != nil {
		t.Log("Fail connecting")
		t.Fail()
	}

	RefreshConnections(&pool, []string{"localhost:8080", "localhost:8090"})
	assert.Equal(t, 2, len(pool))

	RefreshConnections(&pool, []string{"localhost:8080"})
	assert.Equal(t, 1, len(pool))

	RefreshConnections(&pool, []string{"localhost:8080", "localhost:8090"})
	assert.Equal(t, 2, len(pool))

	RefreshConnections(&pool, []string{})
	assert.Equal(t, 0, len(pool))
}

func TestRemoveConnection(t *testing.T) {
	tr := getTransport()
	l := fakeListener("8080")
	l2 := fakeListener("8090")
	defer l.Close()
	defer l2.Close()

	pool := []*Connection{}
	con := &Connection{Address: "localhost:8080"}
	con2 := &Connection{Address: "localhost:8090"}
	AddConnection(&pool, con)
	AddConnection(&pool, con2)

	if err := ConnectPool(tr, pool); err != nil {
		t.Log("Fail connecting")
		t.Fail()
	}

	removeConnection(&pool, "localhost:8091")
	assert.Equal(t, 2, len(pool))

	removeConnection(&pool, "localhost:8080")
	assert.Equal(t, 1, len(pool))

	removeConnection(&pool, "localhost:8090")
	assert.Equal(t, 0, len(pool))
}

func TestRemoveAllConnections(t *testing.T) {
	tr := getTransport()
	l := fakeListener("8080")
	l2 := fakeListener("8090")
	defer l.Close()
	defer l2.Close()

	pool := []*Connection{}
	con := &Connection{Address: "localhost:8080", IsActive: true}
	con2 := &Connection{Address: "localhost:8090", IsActive: true}
	AddConnection(&pool, con)
	AddConnection(&pool, con2)

	if err := ConnectPool(tr, pool); err != nil {
		t.Log("Fail connecting")
		t.Fail()
	}

	CloseAllConnections(&pool)
	assert.Equal(t, 0, len(pool))

	// force error in close connection
	con = &Connection{Address: "localhost:8080", IsActive: true}
	AddConnection(&pool, con)

	if err := ConnectPool(tr, pool); err != nil {
		t.Log("Fail connecting")
		t.Fail()
	}

	con.Conn.Close()
	CloseAllConnections(&pool)
}

func fakeListener(port string) net.Listener {
	l, err := net.Listen("tcp", "localhost:"+port)
	if err != nil {
		log.Fatalln(err)
	}

	return l
}

func getTransport() *http2.Transport {
	return &http2.Transport{
		DisableCompression: true,
		AllowHTTP:          true,
		DialTLS: func(netw, addr string, cfg *tls.Config) (net.Conn, error) {
			return net.Dial(netw, addr)
		},
	}
}
