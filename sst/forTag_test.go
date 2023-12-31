package sst

import (
	"fmt"
	"lsmstore/commitlog"
	"lsmstore/utils"
	"os"
	"sync"
	"testing"
	"time"

	log "github.com/jeanphorn/log4go"
	"github.com/stretchr/testify/assert"
)

func TestSSTforTag_SanityCheck(t *testing.T) {
	//given
	st := SSTforTag{FileName: fmt.Sprintf("/tmp/golsm_test/testForTag-%d-%d.db", utils.GetNowMillis(), utils.GetTestIdx())}
	st.InitStorage()

	//when
	actualEntries := getDummyCommitlogEntries()
	st.MergeWithCommitlog(actualEntries)
	min, max := st.Availability()

	//then
	retrievedEntries := st.GetAllEntries()
	assert.Equal(t, 4, len(retrievedEntries), "entries mismatch")
	i := 0
	for i < len(retrievedEntries)-1 {
		j := i + 1
		assert.LessOrEqual(t, retrievedEntries[i].Timestamp, retrievedEntries[j].Timestamp, "entries in SST are not sorted")
		i += 1
	}
	assert.Equal(t, uint64(1337), min, "min ts incorrect")
	assert.Equal(t, uint64(1343), max, "max ts incorrect")

	//given
	st = SSTforTag{FileName: st.FileName}
	st.InitStorage()

	//when
	min, max = st.Availability()

	//then
	assert.Equal(t, uint64(1337), min, "min ts incorrect after file reopening")
	assert.Equal(t, uint64(1343), max, "max ts incorrect after file reopening")

	//when
	actualEntries2 := getDummyCommitlogEntries2()
	st.MergeWithCommitlog(actualEntries2)
	min2, max2 := st.Availability()

	//then
	retrievedEntries2 := st.GetAllEntries()
	assert.Equal(t, 6, len(retrievedEntries2), "entries mismatch")
	i = 0
	for i < len(retrievedEntries2)-1 {
		j := i + 1
		assert.LessOrEqual(t, retrievedEntries2[i].Timestamp, retrievedEntries2[j].Timestamp, "entries in SST are not sorted after overriding/resorting merge")
		i += 1
	}
	assert.Equal(t, uint64(1337), min, "min ts incorrect")
	assert.Equal(t, uint64(1343), max, "max ts incorrect")
	assert.Equal(t, uint64(1337), min2, "min ts incorrect after overriding/resorting merge")
	assert.Equal(t, uint64(1345), max2, "max ts incorrect after overriding/resorting merge")
}

func TestSSTforTag_IndexWorks(t *testing.T) {
	//given
	st := SSTforTag{FileName: fmt.Sprintf("/tmp/golsm_test/testForTag-%d-%d.db", utils.GetNowMillis(), utils.GetTestIdx())}
	st.InitStorage()

	//when
	actualEntries := getBigBatchOfEntries(1000, 1000, 0)
	st.MergeWithCommitlog(actualEntries)
	min, max := st.Availability()

	//then
	assert.Equal(t, uint64(10000), min, "min ts incorrect")
	assert.Equal(t, uint64(19990), max, "max ts incorrect")

	//when
	slice1 := st.GetEntriesWithoutIndex(15000, 16000)
	//then
	assert.Equal(t, 101, len(slice1), "entries count is incorrect without index")

	//when
	slice2 := st.GetEntriesWithIndex(15000, 16000)
	//then
	assert.Equal(t, 101, len(slice2), "entries count is incorrect with index")

	for i := range slice1 {
		assert.Equal(t, slice1[i], slice2[i], "entry is not the same when with and without index")
	}
}

func TestSSTforTag_ParallelReadsWritesWork(t *testing.T) {
	//given
	st := SSTforTag{FileName: fmt.Sprintf("/tmp/golsm_test/testForTag-%d-%d.db", utils.GetNowMillis(), utils.GetTestIdx())}
	st.InitStorage()

	//when
	actualEntries := getBigBatchOfEntries(1000, 1000, 0)
	st.MergeWithCommitlog(actualEntries)
	min, max := st.Availability()

	//then
	assert.Equal(t, uint64(10000), min, "min ts incorrect")
	assert.Equal(t, uint64(19990), max, "max ts incorrect")

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			newEntries := getBigBatchOfEntries(10, utils.GetNowMillis()/10, 0)
			st.MergeWithCommitlog(newEntries)
			time.Sleep(2 * time.Second)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			slice := st.GetEntriesWithoutIndex(15000, utils.GetNowMillis()/10)
			assert.Less(t, 0, len(slice), "entries count is incorrect with index")
			time.Sleep(time.Second)
		}
	}()

	wg.Wait()
}

func TestSSTforTag_WritesSortedFile(t *testing.T) {
	//given
	path := "/tmp/golsm_test/test_3yYHfn_2"
	os.Remove(path)
	st := SSTforTag{FileName: path}
	st.InitStorage()

	//when
	actualEntries1 := getBigBatchOfEntries(1000, 1000, 0)
	st.MergeWithCommitlog(actualEntries1)
	actualEntries2 := getBigBatchOfEntries(1000, 500, 0)
	st.MergeWithCommitlog(actualEntries2)
	actualEntries3 := getBigBatchOfEntries(1000, 750, 0)
	st.MergeWithCommitlog(actualEntries3)

	entries := st.GetAllEntries()

	//then
	assert.Equal(t, 1500, len(entries), "size incorrect") //not 3000 because of repeating TSs
}

func TestSSTforTag_ReadsExistingFile(t *testing.T) {
	//given
	st := SSTforTag{FileName: "test_3yYHfn"}
	st.InitStorage()

	//when
	min, max := st.Availability()

	//then
	assert.Equal(t, uint64(1599759420524), min, "min ts incorrect") //Thursday, 10 September 2020 г., 17:37:00.524
	assert.Equal(t, uint64(1599763019524), max, "max ts incorrect") //Thursday, 10 September 2020 г., 18:36:59.524

	//when
	all := st.GetAllEntries()
	assert.Equal(t, 3600, len(all), "all entries count is incorrect")

	for i := 0; i < 10000; i++ {
		from := randomTs(min+20, min+(max-min)/2)
		to := randomTs(from, max)
		if to-from <= 10 {
			from -= 10
		}
		dWithoutIndex := st.GetEntriesWithoutIndex(from, to)
		//fmt.Printf("without index for %d - %d : %d points\n", from, to, len(dWithoutIndex))

		dWithIndex := st.GetEntriesWithIndex(from, to)
		//fmt.Printf("with index for %d - %d : %d points\n", from, to, len(dWithIndex))

		assert.Equal(t, len(dWithoutIndex), len(dWithIndex), fmt.Sprintf("size incorrect for %d-%d: w/ %d, w/o %d", from, to, len(dWithIndex), len(dWithoutIndex)))
	}
}

func Teardown(t *testing.T) {
	log.Close()
}

func getDummyCommitlogEntries() []commitlog.Entry {
	ans := make([]commitlog.Entry, 4)
	ans[0] = commitlog.Entry{Key: []byte("tagZero"), Timestamp: 1337, ExpiresAt: 0, Value: make([]byte, 4)}
	ans[1] = commitlog.Entry{Key: []byte("tagZero"), Timestamp: 1339, ExpiresAt: 0, Value: make([]byte, 2)}
	ans[2] = commitlog.Entry{Key: []byte("tagZero"), Timestamp: 1341, ExpiresAt: 0, Value: make([]byte, 16)}
	ans[3] = commitlog.Entry{Key: []byte("tagZero"), Timestamp: 1343, ExpiresAt: 0, Value: make([]byte, 1)}
	return ans
}

func getDummyCommitlogEntries2() []commitlog.Entry {
	ans := make([]commitlog.Entry, 2)
	ans[0] = commitlog.Entry{Key: []byte("tagZero"), Timestamp: 1338, ExpiresAt: 0, Value: make([]byte, 4)}
	ans[1] = commitlog.Entry{Key: []byte("tagZero"), Timestamp: 1345, ExpiresAt: 0, Value: make([]byte, 2)}
	return ans
}
func getBigBatchOfEntries(count int, firstTs uint64, delta uint64) []commitlog.Entry {
	return getBigBatchOfEntriesOfSize(count, firstTs, delta, 4)
}

func getBigBatchOfEntriesOfSize(count int, firstTs uint64, delta uint64, size int) []commitlog.Entry {
	ans := make([]commitlog.Entry, count)
	i := 0
	for i < count {
		ans[i] = commitlog.Entry{Key: []byte("tagZero"), Timestamp: (firstTs+uint64(i))*10 + delta, ExpiresAt: 0, Value: make([]byte, size)}
		i++
	}
	return ans
}
