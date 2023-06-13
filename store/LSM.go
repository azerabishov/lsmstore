package store

import (
	"lsmstore/commitlog"
	"lsmstore/memt"
	"lsmstore/sst"
	"lsmstore/writer"
	"time"
)

func InitStorage(commitlogPath string, entriesPerCommitlog int, periodBetweenFlushes time.Duration, memtPerformExpirationEvery time.Duration, memtPrefetchSeconds time.Duration, sstPath string, memtMaxEntriesPerTag int) (*StorageReader, *StorageWriter) {
	clm := commitlog.Manager{Path: commitlogPath}
	sstm := sst.Manager{RootDir: sstPath}
	dw := writer.DiskWriter{SstManager: &sstm, ClManager: &clm, EntriesPerCommitlog: entriesPerCommitlog, PeriodBetweenFlushes: periodBetweenFlushes}
	dw.Init()

	memtm := memt.Manager{MaxEntriesPerTag: memtMaxEntriesPerTag, PerformExpirationEvery: memtPerformExpirationEvery}
	memtm.InitStorage()

	storageWriter := StorageWriter{MemTable: &memtm, DiskWriter: &dw}
	storageWriter.Init()

	storageReader := StorageReader{MemTable: &memtm, SSTManager: &sstm, MemtPrefetch: memtPrefetchSeconds}
	storageReader.Init()

	return &storageReader, &storageWriter
}
