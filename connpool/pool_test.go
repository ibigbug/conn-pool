package connpool

import (
	"testing"
	"time"
)

const (
	BLOG  = "104.199.153.228:80"
	NINJA = "106.187.103.116:80"
)

func handleError(err error, t *testing.T) {
	if err != nil {
		t.Error(err)
	}
}

func TestConnpoolGetAndRelease(t *testing.T) {
	p := NewPool()
	p.SetKeepAliveTimeout(3 * time.Second)

	conn1, err := p.GetConn("blog.xiaoba.me")
	handleError(err, t)
	if conn1 == nil {
		t.Error(1)
	}

	if len(p.pool[BLOG]) != 1 {
		t.Error(2)
	}

	conn2, err := p.GetConn("4zai.zhuangbi.ninja")
	handleError(err, t)
	if conn2 == nil {
		t.Error(3)
	}

	if len(p.pool[NINJA]) != 1 {
		t.Error(4)
	}

	timer := time.NewTimer(3 * time.Second)
	<-timer.C
	if len(p.pool[BLOG]) != 1 {
		t.Error(5)
	}
	if len(p.pool[NINJA]) != 1 {
		t.Error(6)
	}

	p.ReleaseConn(conn1)
	p.ReleaseConn(conn2)

	timer = time.NewTimer(4 * time.Second)
	<-timer.C
	if len(p.pool[BLOG]) != 0 {
		t.Error(7)
	}
	if len(p.pool[NINJA]) != 0 {
		t.Error(8)
	}
}
