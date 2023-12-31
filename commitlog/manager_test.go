package commitlog_test

import (
	"lsmstore/commitlog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommitlog(t *testing.T) {
	//given
	m := commitlog.Manager{Path: "/tmp/golsm/test"}
	m.Init()
	dummy1 := commitlog.Entry{
		Key:       []byte("tagZero"),
		Timestamp: 1337,
		Value:     make([]byte, 2),
		ExpiresAt: 9999,
	}
	dummy2 := commitlog.Entry{
		Key:       []byte("tagOne"),
		Timestamp: 1489,
		Value:     make([]byte, 3),
		ExpiresAt: 9999,
	}
	dummy3 := commitlog.Entry{
		Key:       []byte("tagTwo"),
		Timestamp: 1490,
		Value:     make([]byte, 4),
		ExpiresAt: 9999,
	}
	dummy4 := commitlog.Entry{
		Key:       []byte("tagThree"),
		Timestamp: 1338,
		Value:     make([]byte, 2),
		ExpiresAt: 9999,
	}
	//when
	m.Store(dummy1)
	m.SwapCommitlogs()
	m.Store(dummy2)
	m.Store(dummy3)
	m.SwapCommitlogs()
	m.Store(dummy4)

	//then
	all1 := m.RetrieveAll()
	m.SwapCommitlogs()
	all2 := m.RetrieveAll()

	m.Store(dummy1)

	assert.Equal(t, dummy1, all1[0], "commitlogA failed")
	assert.Equal(t, dummy2, all2[0], "commitlogB failed")
	assert.Equal(t, dummy3, all2[1], "commitlogB failed on second item")
	assert.Equal(t, dummy4, all1[1], "commitlogA failed on second item")
}
