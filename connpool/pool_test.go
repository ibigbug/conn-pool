package connpool

import (
	"net"
	"testing"
	"time"
)

const (
	A1 = "110.76.19.33:80"
	A2 = "110.76.20.33:80"
)

func handleError(err error, t *testing.T) {
	if err != nil {
		t.Error(err)
	}
}

func TestUseReleasedConnImmediately(t *testing.T) {
	p := NewPool()
	p.SetKeepAliveTimeout(3 * time.Second)

	conn1, err := p.Get("a1.alipay-inc.xyz")
	handleError(err, t)
	if conn1 == nil {
		t.Error(1)
	}

	if len(p.pool[A1].conns) != 1 {
		t.Error(2)
	}

	// mark conn1 as idle
	p.Put(conn1)

	// reuse it, conn1.idle should be false
	conn2, err := p.Get("a1.alipay-inc.xyz")

	// mark it idle again
	// now there should be to Release go routine hold this conn
	p.Put(conn2)

	timer := time.NewTimer(5 * time.Second)
	<-timer.C

	// there should be nothing
	mgr := p.pool[A1]
	mgr.Lock()
	defer mgr.Unlock()

	if len(mgr.conns) != 0 {
		t.Error(3)
	}
}

func TestConnCache(t *testing.T) {
	p := NewPool()
	p.SetKeepAliveTimeout(3 * time.Second)

	conn1, err := p.Get("a1.alipay-inc.xyz")
	handleError(err, t)
	if conn1 == nil {
		t.Error(1)
	}

	if len(p.pool[A1].conns) != 1 {
		t.Error(2)
	}

	conn2, err := p.Get("a1.alipay-inc.xyz")
	handleError(err, t)
	if conn2 == nil {
		t.Error(3)
	}
	if conn1 == conn2 {
		// should create new conn
		t.Error(4)
	}

	if len(p.pool[A1].conns) != 2 {
		t.Error(5)
	}

	p.Put(conn1)

	conn3, err := p.Get("a1.alipay-inc.xyz")
	handleError(err, t)
	if conn3 == nil {
		t.Error(6)
	}

	// wait Release conn1 go-routine pass
	t1 := time.NewTimer(5 * time.Second)
	<-t1.C

	// conn1 should be kept
	if len(p.pool[A1].conns) != 2 {
		t.Error(7)
	}

	if conn1 != conn3 {
		// should reuse conn
		t.Error(8)
	}

	p.Put(conn2)
	p.Put(conn3)

	conn1 = nil
	conn2 = nil
	conn3 = nil

	timer := time.NewTimer(5 * time.Second)
	<-timer.C

	mgr := p.pool[A1]
	mgr.Lock()
	defer mgr.Unlock()
	if len(mgr.conns) != 0 {
		t.Error(9)
	}
}

func TestConnpoolGetAndRelease(t *testing.T) {
	p := NewPool()
	p.SetKeepAliveTimeout(3 * time.Second)

	conn1, err := p.Get("a1.alipay-inc.xyz")
	handleError(err, t)
	if conn1 == nil {
		t.Error(1)
	}

	if len(p.pool[A1].conns) != 1 {
		t.Error(2)
	}

	conn2, err := p.Get("a2.alipay-inc.xyz")
	handleError(err, t)
	if conn2 == nil {
		t.Error(3)
	}

	if len(p.pool[A2].conns) != 1 {
		t.Error(4)
	}

	timer := time.NewTimer(3 * time.Second)
	<-timer.C
	if len(p.pool[A1].conns) != 1 {
		t.Error(5)
	}
	if len(p.pool[A2].conns) != 1 {
		t.Error(6)
	}

	p.Put(conn1)
	p.Put(conn2)

	timer = time.NewTimer(4 * time.Second)
	<-timer.C

	mgr1 := p.pool[A1]
	mgr1.Lock()
	defer mgr1.Unlock()

	if len(p.pool[A1].conns) != 0 {
		t.Error(7)
	}

	mgr2 := p.pool[A2]
	mgr2.Lock()
	defer mgr2.Unlock()
	if len(p.pool[A2].conns) != 0 {
		t.Error(8)
	}
}

func TestRemoveConn(t *testing.T) {
	p := NewPool()
	p.SetKeepAliveTimeout(3 * time.Second)

	conn1, err := p.Get("a1.alipay-inc.xyz")
	handleError(err, t)
	if conn1 == nil {
		t.Error(1)
	}

	if len(p.pool[A1].conns) != 1 {
		t.Error(2)
	}

	conn2, err := p.Get("a2.alipay-inc.xyz")
	handleError(err, t)
	if conn2 == nil {
		t.Error(3)
	}

	if len(p.pool[A2].conns) != 1 {
		t.Error(4)
	}

	timer := time.NewTimer(3 * time.Second)
	<-timer.C
	if len(p.pool[A1].conns) != 1 {
		t.Error(5)
	}
	if len(p.pool[A2].conns) != 1 {
		t.Error(6)
	}

	p.Remove(conn1)
	p.Remove(conn2)

	mgr1 := p.pool[A1]
	mgr1.Lock()
	defer mgr1.Unlock()

	if len(p.pool[A1].conns) != 0 {
		t.Error(7)
	}

	mgr2 := p.pool[A2]
	mgr2.Lock()
	defer mgr2.Unlock()
	if len(p.pool[A2].conns) != 0 {
		t.Error(8)
	}
}

func TestTimeout(t *testing.T) {
	p := NewPool()
	p.SetKeepAliveTimeout(3 * time.Second)
	_, err := p.GetTimeout("a1.alipay-inc.xyz", 1*time.Millisecond)
	if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
		return
	}
	t.Error(9)
}
