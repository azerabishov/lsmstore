package sst

import (
	"fmt"
	"lsmstore/commitlog"
	"lsmstore/utils"
	"testing"

	log "github.com/jeanphorn/log4go"
	"github.com/stretchr/testify/assert"
)

func TestSSTManager_SanityCheck(t *testing.T) {
	//given
	m := Manager{RootDir: fmt.Sprintf("/tmp/golsm_test/test-for-SSTManager-%d-%d", utils.GetNowMillis(), utils.GetTestIdx())}
	m.InitStorage()

	//when
	actualEntries := getDummyCommitlogEntriesForMultipleTags()
	m.MergeWithCommitlog(actualEntries)

	//then
	assert.Equal(t, 2, len(m.sstForTag), "sst count mismatch")

	st1 := m.sstForTag["tagZero"]
	st2 := m.sstForTag["tagOne"]

	st1e := st1.GetAllEntries()
	st2e := st2.GetAllEntries()

	assert.Equal(t, 3, len(st1e), "dto count in sst mismatch for tagZero")
	assert.Equal(t, 2, len(st2e), "dto count in sst mismatch for tagOne")

	//given
	m = Manager{RootDir: m.RootDir}
	m.InitStorage()

	//when
	actualEntries2 := getDummyCommitlogEntriesForMultipleTags2()
	m.MergeWithCommitlog(actualEntries2)

	//then
	assert.Equal(t, 2, len(m.sstForTag), "sst count mismatch")

	st1 = m.sstForTag["tagZero"]
	st2 = m.sstForTag["tagOne"]

	st1e = st1.GetAllEntries()
	st2e = st2.GetAllEntries()

	assert.Equal(t, 4, len(st1e), "dto count in sst mismatch for tagZero after reopening")
	assert.Equal(t, 3, len(st2e), "dto count in sst mismatch for tagOne after reopening")

	log.Close()
}

func getDummyCommitlogEntriesForMultipleTags() []commitlog.Entry {
	ans := make([]commitlog.Entry, 5)
	ans[0] = commitlog.Entry{Key: []byte("tagZero"), Timestamp: 1337, ExpiresAt: 0, Value: make([]byte, 4)}
	ans[1] = commitlog.Entry{Key: []byte("tagOne"), Timestamp: 1339, ExpiresAt: 0, Value: make([]byte, 2)}
	ans[2] = commitlog.Entry{Key: []byte("tagZero"), Timestamp: 1341, ExpiresAt: 0, Value: make([]byte, 16)}
	ans[3] = commitlog.Entry{Key: []byte("tagOne"), Timestamp: 1343, ExpiresAt: 0, Value: make([]byte, 1)}
	ans[4] = commitlog.Entry{Key: []byte("tagZero"), Timestamp: 1345, ExpiresAt: 0, Value: make([]byte, 1)}
	return ans
}

func getDummyCommitlogEntriesForMultipleTags2() []commitlog.Entry {
	ans := make([]commitlog.Entry, 2)
	ans[0] = commitlog.Entry{Key: []byte("tagZero"), Timestamp: 1339, ExpiresAt: 0, Value: make([]byte, 4)}
	ans[1] = commitlog.Entry{Key: []byte("tagOne"), Timestamp: 1341, ExpiresAt: 0, Value: make([]byte, 2)}
	return ans
}
