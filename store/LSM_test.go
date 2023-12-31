package store

import (
	"fmt"
	"lsmstore/commitlog"
	"lsmstore/dto"
	"lsmstore/memt"
	"lsmstore/sst"
	"lsmstore/utils"
	"lsmstore/writer"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLSM_StorageWriterWorks(t *testing.T) {
	//given
	clm := commitlog.Manager{Path: fmt.Sprintf("/tmp/golsm_test/diskwriter/commitlog-%d-%d", utils.GetNowMillis(), utils.GetTestIdx())}
	sstm := sst.Manager{RootDir: fmt.Sprintf("/tmp/golsm_test/diskwriter/sstm-%d-%d", utils.GetNowMillis(), utils.GetTestIdx())}
	dw := writer.DiskWriter{SstManager: &sstm, ClManager: &clm, EntriesPerCommitlog: 10, PeriodBetweenFlushes: 5 * time.Second}
	dw.Init()

	memtm := memt.Manager{MaxEntriesPerTag: 9999}
	memtm.InitStorage()

	storageWriter := StorageWriter{MemTable: &memtm, DiskWriter: &dw}
	storageWriter.Init()
	const tagName = "whatever"
	const expiration = 0

	dummyData := buildDummyData(25)

	//when
	storageWriter.StoreMultiple(slice(dummyData, tagName, 0, 16), expiration)
	time.Sleep(1 * time.Second)
	storageWriter.StoreMultiple(slice(dummyData, tagName, 16, 20), expiration)
	time.Sleep(1 * time.Second)
	storageWriter.StoreMultiple(slice(dummyData, tagName, 20, 25), expiration)
	time.Sleep(10 * time.Second)

	storedDataOnDisk := sstm.SstForTag(tagName).GetAllEntries()
	storedDataInMemT := memtm.MemTableForTag(tagName).RetrieveAll()

	//then
	assert.Equal(t, len(dummyData), len(storedDataOnDisk), "some dto was lost")
	assert.Equal(t, len(dummyData), len(storedDataInMemT), "some dto was lost")
	for i := 0; i < 25; i++ {
		assert.Equal(t, dummyData[i].Timestamp, storedDataOnDisk[i].Timestamp, "entry timestamp incorrect")
		assert.Equal(t, dummyData[i].Value, storedDataOnDisk[i].Value, "entry value incorrect")

		assert.Equal(t, dummyData[i].Timestamp, storedDataInMemT[i].Timestamp, "entry timestamp incorrect")
		assert.Equal(t, dummyData[i].Value, storedDataInMemT[i].Value, "entry value incorrect")
	}
}

func TestLSM_StorageReaderWorks(t *testing.T) {
	storageReader, storageWriter := InitStorage(
		fmt.Sprintf("/tmp/golsm_test/diskwriter/commitlog-%d-%d", utils.GetNowMillis(), utils.GetTestIdx()),
		10,
		5*time.Second,
		10*time.Second,
		10*time.Second,
		fmt.Sprintf("/tmp/golsm_test/diskwriter/sstm-%d-%d", utils.GetNowMillis(), utils.GetTestIdx()),
		9999)

	const tagName = "whatever"
	const expiration = 0

	dummyData := buildDummyData(25)

	//when
	storageWriter.StoreMultiple(slice(dummyData, tagName, 0, 25), expiration)
	retrievedData := storageReader.Retrieve(toList(tagName), 1336, 1500)
	availFrom, availTo := storageReader.Availability()

	//then
	assert.Equal(t, 1, len(retrievedData), "weird stuff returned from StorageReader")
	assert.Equal(t, dummyData[0].Timestamp, availFrom, "availFrom incorrect")
	assert.Equal(t, dummyData[24].Timestamp, availTo, "availTo incorrect")
	retrievedDataForTag := retrievedData[tagName]
	assert.Equal(t, len(dummyData), len(retrievedDataForTag), "some dto was lost")
	for i := 0; i < 25; i++ {
		assert.Equal(t, dummyData[i].Timestamp, retrievedDataForTag[i].Timestamp, "measurement timestamp incorrect")
		assert.Equal(t, dummyData[i].Value, retrievedDataForTag[i].Value, "measurement value incorrect")
	}
}

func TestLSM_StorageWriterBatchingWorks(t *testing.T) {
	//given
	clm := commitlog.Manager{Path: fmt.Sprintf("/tmp/golsm_test/diskwriter/commitlog-%d-%d", utils.GetNowMillis(), utils.GetTestIdx())}
	sstm := sst.Manager{RootDir: fmt.Sprintf("/tmp/golsm_test/diskwriter/sstm-%d-%d", utils.GetNowMillis(), utils.GetTestIdx())}
	dw := writer.DiskWriter{SstManager: &sstm, ClManager: &clm, EntriesPerCommitlog: 10, PeriodBetweenFlushes: 5 * time.Second}
	dw.Init()

	memtm := memt.Manager{MaxEntriesPerTag: 9999}
	memtm.InitStorage()

	storageWriter := StorageWriter{MemTable: &memtm, DiskWriter: &dw}
	storageWriter.Init()
	const tagName = "whatever"
	const expiration = 0

	dummyData := buildDummyData(25)

	//when
	storageWriter.StoreBatch(sliceAndToBatch(dummyData, tagName, 0, 16), expiration)
	time.Sleep(1 * time.Second)
	storageWriter.StoreBatch(sliceAndToBatch(dummyData, tagName, 16, 20), expiration)
	time.Sleep(1 * time.Second)
	storageWriter.StoreBatch(sliceAndToBatch(dummyData, tagName, 20, 25), expiration)
	time.Sleep(10 * time.Second)

	storedDataOnDisk := sstm.SstForTag(tagName).GetAllEntries()
	storedDataInMemT := memtm.MemTableForTag(tagName).RetrieveAll()

	//then
	assert.Equal(t, len(dummyData), len(storedDataOnDisk), "some dto was lost")
	assert.Equal(t, len(dummyData), len(storedDataInMemT), "some dto was lost")
	for i := 0; i < 25; i++ {
		assert.Equal(t, dummyData[i].Timestamp, storedDataOnDisk[i].Timestamp, "entry timestamp incorrect")
		assert.Equal(t, dummyData[i].Value, storedDataOnDisk[i].Value, "entry value incorrect")

		assert.Equal(t, dummyData[i].Timestamp, storedDataInMemT[i].Timestamp, "entry timestamp incorrect")
		assert.Equal(t, dummyData[i].Value, storedDataInMemT[i].Value, "entry value incorrect")
	}
}

func TestLSM_StorageReaderOnBigDataTest(t *testing.T) {
	storageReader, storageWriter := InitStorage(
		fmt.Sprintf("/tmp/golsm_test/diskwriter/commitlog-%d-%d", utils.GetNowMillis(), utils.GetTestIdx()),
		10,
		1*time.Second,
		1*time.Second,
		1*time.Second,
		fmt.Sprintf("/tmp/golsm_test/diskwriter/sstm-%d-%d", utils.GetNowMillis(), utils.GetTestIdx()),
		100)

	dataFrom := utils.GetNowMillis()
	dataTo := dataFrom + 60*60*1000
	tagsCount := 10
	data := buildDummyDataForBenchmark(tagsCount, dataFrom, dataTo)

	//when
	storageWriter.StoreMultiple(data, 0)

	for i := 0; i < 10000; i++ {
		from := randomTs(dataFrom+20, dataFrom+(dataTo-dataFrom)/2)
		to := randomTs(from, dataTo)
		if to-from <= 10 {
			from -= 10
		}
		d := storageReader.Retrieve([]string{"tag1", "tag2", "tag3"}, from, to)
		if len(d) != 3 {
			panic("tags mismatch")
		}
	}
}

func TestLoad(t *testing.T) {
	const numUsers = 4
	const duration = 5 * time.Second

	wg := &sync.WaitGroup{}
	wg.Add(numUsers)
	_, storageWriter := InitStorage(
		"./tmp/golsm_test/diskwriter/commitlog",
		10,
		1*time.Second,
		1*time.Second,
		1*time.Second,
		"./tmp/golsm_test/diskwriter/sstm",
		100)

	for i := 0; i < numUsers; i++ {
		go func(id int) {
			defer wg.Done()
			tagName := fmt.Sprintf("tag%d", id)

			ticker := time.NewTicker(duration)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					data := dto.TaggedMeasurement{Tag: tagName, Timestamp: uint64(time.Now().UnixMilli()), Value: make([]byte, 5)}
					storageWriter.Store(data, uint64(time.Now().Add(time.Hour*48).UnixMilli()))
				}
			}
		}(i)
	}

	wg.Wait()
}

func randomTs(from uint64, to uint64) uint64 {
	return uint64(rand.Float64()*float64(to-from) + float64(from))
}

func buildDummyData(count int) []dto.Measurement {
	ans := make([]dto.Measurement, count)
	for i := 0; i < count; i++ {
		ans[i] = dto.Measurement{Timestamp: 1337 + uint64(i), Value: make([]byte, 4)}
	}
	return ans
}

func buildDummyDataForBenchmark(tagsCount int, tsFrom uint64, tsTo uint64) map[string][]dto.Measurement {
	tags := make([]string, tagsCount)
	for i := 0; i < tagsCount; i++ {
		tags[i] = fmt.Sprintf("tag%d", i)
	}
	ans := make(map[string][]dto.Measurement)
	for _, tag := range tags {
		data := make([]dto.Measurement, 0)
		for ts := tsFrom; ts < tsTo; ts += 1000 {
			data = append(data, dto.Measurement{Timestamp: ts, Value: make([]byte, 4)})
			ans[tag] = data
		}
	}
	return ans
}

func slice(data []dto.Measurement, tag string, from int, to int) map[string][]dto.Measurement {
	ans := make(map[string][]dto.Measurement)
	ans[tag] = make([]dto.Measurement, 0)
	for i := from; i < to; i++ {
		ans[tag] = append(ans[tag], data[i])
	}
	return ans
}

func sliceAndToBatch(data []dto.Measurement, tag string, from int, to int) []dto.TaggedMeasurement {
	ans := make([]dto.TaggedMeasurement, 0)
	for i := from; i < to; i++ {
		ans = append(ans, dto.TaggedMeasurement{Tag: tag, Timestamp: data[i].Timestamp, Value: data[i].Value})
	}
	return ans
}

func toList(tag string) []string {
	ans := make([]string, 1)
	ans[0] = tag
	return ans
}
