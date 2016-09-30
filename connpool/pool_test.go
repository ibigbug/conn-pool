package connpool

import (
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

func TestConnpoolGetAndRelease(t *testing.T) {
	p := NewPool()
	p.SetKeepAliveTimeout(3 * time.Second)

	conn1, err := p.GetConn("a1.alipay-inc.xyz")
	handleError(err, t)
	if conn1 == nil {
		t.Error(1)
	}

	if len(p.pool[A1]) != 1 {
		t.Error(2)
	}

	conn2, err := p.GetConn("a2.alipay-inc.xyz")
	handleError(err, t)
	if conn2 == nil {
		t.Error(3)
	}

	if len(p.pool[A2]) != 1 {
		t.Error(4)
	}

	timer := time.NewTimer(3 * time.Second)
	<-timer.C
	if len(p.pool[A1]) != 1 {
		t.Error(5)
	}
	if len(p.pool[A2]) != 1 {
		t.Error(6)
	}

	p.ReleaseConn(conn1)
	p.ReleaseConn(conn2)

	timer = time.NewTimer(4 * time.Second)
	<-timer.C
	if len(p.pool[A1]) != 0 {
		t.Error(7)
	}
	if len(p.pool[A2]) != 0 {
		t.Error(8)
	}
}
