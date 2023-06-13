package main

import (
	"fmt"
	"lsmstore/dto"
	"lsmstore/store"
	"sync"
	"time"
)

func main() {
	// read()
	write()
}

func read() {
	storageReader, _ := store.InitStorage(
		"./tmp/golsm_test/diskwriter/commitlog",
		10,
		1*time.Second,
		1*time.Second,
		1*time.Second,
		"./tmp/golsm_test/diskwriter/sstm",
		100)

	from, to := storageReader.Availability()

	fmt.Println(time.UnixMilli(int64(from)))
	fmt.Println(time.UnixMilli(int64(to)))
	response := storageReader.Retrieve([]string{"tag1", "tag100"}, uint64(time.Now().Add(-time.Minute*50).UnixMilli()), uint64(time.Now().Add(-time.Minute*1).UnixMilli()))
	fmt.Println("response: ", len(response["tag1"]))

}

func write() {
	const numUsers = 100
	const duration = 1 * time.Second

	wg := &sync.WaitGroup{}
	wg.Add(numUsers)

	_, storageWriter := store.InitStorage(
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
