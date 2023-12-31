package commitlog

import (
	"encoding/binary"
	"lsmstore/utils"
	"os"
)

type Commitlog interface {
	Init()
	Store(entry Entry)
	RetrieveAll() []Entry
	Count() int
	Clear()
}

type OverFile struct {
	commitlogFileName string
	commitlogFile     *os.File
	entriesCount      int
}

func (o *OverFile) Init() {
	file, err := os.OpenFile(o.commitlogFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	utils.Check(err)
	o.commitlogFile = file
}

func (o *OverFile) Store(entry Entry) {
	//log.Debug("STORE on " + o.commitlogFileName + " ts " + strconv.FormatUint(entry.Timestamp, 10))
	o.commitlogFile.Write(entry.ToByteArrayWithLength())
	o.entriesCount += 1
}

func (o *OverFile) RetrieveAll() []Entry {
	//log.Debug("RETRIEVE ALL on " + o.commitlogFileName)
	return o.readAllEntries()
}

func (o *OverFile) Count() int {
	return o.entriesCount
}

func (o *OverFile) readAllEntries() []Entry {
	o.commitlogFile.Close()
	f, err := os.OpenFile(o.commitlogFileName, os.O_RDONLY, 0644)
	utils.Check(err)
	buf := make([]byte, 2)
	ans := make([]Entry, 0)
	n := 2
	n, _ = f.Read(buf)
	for n == 2 {
		lenToRead := int(binary.LittleEndian.Uint16(buf))
		bigbuf := make([]byte, lenToRead)
		n2, _ := f.Read(bigbuf)
		if n2 != lenToRead {
			panic("fail")
		}
		ans = append(ans, FromByteArray(bigbuf))
		n, _ = f.Read(buf)
	}
	f.Close()
	o.Init()
	return ans
}

func (o *OverFile) Clear() {
	//log.Debug("CLEAR on " + o.commitlogFileName)
	o.commitlogFile.Close()
	utils.Check(os.Remove(o.commitlogFileName))
	o.Init()
}
