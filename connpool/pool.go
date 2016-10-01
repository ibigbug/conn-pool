package connpool

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// DefaultKeepAliveTimeout default keepalive timeout in seconds
const (
	DefaultKeepAliveTimeout = time.Second * 60
)

// ManagedConn net.TCPConn with idle mark
type ManagedConn struct {
	net.TCPConn
	idle bool
}

// NewPool get a new ConnectionPool
func NewPool() ConnectionPool {
	return ConnectionPool{
		timeout: DefaultKeepAliveTimeout,
		pool:    make(map[string][]*ManagedConn),
		mutex:   &sync.Mutex{},
	}
}

// ConnectionPool a connection pool
type ConnectionPool struct {
	timeout time.Duration
	pool    map[string][]*ManagedConn
	mutex   *sync.Mutex
}

// SetKeepAliveTimeout sets after `to` seconds, conn released to pool will be removed
func (p *ConnectionPool) SetKeepAliveTimeout(to time.Duration) {
	p.timeout = to
}

// GetConn get a connection with specific remote address
// could be domain:port/ip:port
func (p ConnectionPool) GetConn(remoteAddr string) (conn *ManagedConn, err error) {
	remoteAddr = ensurePort(remoteAddr)
	tcpAddr, err := net.ResolveTCPAddr("tcp4", remoteAddr)
	if err != nil {
		return
	}
	id := tcpAddr.String()
	conns := p.pool[id]
	if conns == nil {
		p.pool[id] = make([]*ManagedConn, 0)
	}

	if len(conns) == 0 {
		conn = p.createConn(tcpAddr)
	} else {
		conn = p.getFreeConn(id)
		if conn == nil {
			conn = p.createConn(tcpAddr)
		}
	}
	return
}

// ReleaseConn release to pool after used, and will be closed in `timeout` seconds
func (p ConnectionPool) ReleaseConn(conn *ManagedConn) {
	p.mutex.Lock()
	conn.idle = true
	p.mutex.Unlock()

	go func() {
		timer := time.NewTimer(p.timeout)
		<-timer.C
		remoteAddr := conn.RemoteAddr().String()
		p.mutex.Lock()
		idx := findIdx(p.pool[remoteAddr], conn)
		p.pool[remoteAddr] = append(p.pool[remoteAddr][:idx], p.pool[remoteAddr][idx+1:]...)
		conn.Close()
		p.mutex.Unlock()
	}()
}

func (p ConnectionPool) createConn(tcpAddr *net.TCPAddr) (conn *ManagedConn) {
	rawConn, err := net.DialTCP("tcp4", nil, tcpAddr)
	if err == nil {
		conn = &ManagedConn{
			TCPConn: *rawConn,
			idle:    false,
		}
		p.mutex.Lock()
		p.pool[tcpAddr.String()] = append(p.pool[tcpAddr.String()], conn)
		p.mutex.Unlock()
	}
	return
}

func (p ConnectionPool) getFreeConn(tcpAddr string) (c *ManagedConn) {
	for _, c = range p.pool[tcpAddr] {
		if c.idle {
			return
		}
	}
	return
}

func findIdx(arr []*ManagedConn, ele *ManagedConn) int {
	for idx, v := range arr {
		if v == ele {
			return idx
		}
	}
	return -1
}

func ensurePort(addr string) (rv string) {
	rv = addr
	if !strings.Contains(addr, ":") {
		rv = fmt.Sprintf("%s:80", rv)
	}
	return
}
