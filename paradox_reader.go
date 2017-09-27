package main

import (
	"encoding/binary"
	"log"
	"os"

	"github.com/y0ssar1an/q"
)

const sampleFileName = "/Users/Nick/Dropbox/Develop/Upwork/Paradox/Related/Samples/AREA-PDX/AREACODE.DB"

// databaseHeader give the initial layout to the data
type databaseHeader struct {
	recordLength      uint16
	headerBlockSize   uint16
	fileType          uint8
	dataBlockSizeCode byte // 1 K, 2 K, 3K or 4K//
	recordCount       uint32
	blocksUsedCount   uint16
	blocksTotalCount  uint16
	lastBlockInUse    uint16
	fieldCount        uint8
	keyFieldsCount    uint8
}

// blockHeader contains the block record information
type blockHeader struct {
	nextBlockNumber  uint16
	prevBlockNumber  uint16
	offsetLastRecord uint16
}

type fieldDescription struct {
	ordinal   int
	fieldType uint8
	length    uint8
	name      string
}

var fields map[byte]fieldDescription

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func readByteLittleEnd(fileHandle *os.File) (uint8, error) {
	var result byte
	input8 := make([]byte, 1)

	_, err := fileHandle.Read(input8)
	check(err)

	result = input8[0]
	return result, err

}

func readLongLittleEnd(fileHandle *os.File) (uint32, error) {
	var result uint32
	input32 := make([]byte, 4)

	_, err := fileHandle.Read(input32)
	check(err)

	result = binary.LittleEndian.Uint32(input32)
	return result, err
}

func readShortLittleEnd(fileHandle *os.File) (uint16, error) {
	var result uint16
	input16 := make([]byte, 2)

	_, err := fileHandle.Read(input16)
	check(err)

	result = binary.LittleEndian.Uint16(input16)
	return result, err
}

func setupDatabaseHeader(inFile *os.File) (databaseHeader, error) {
	var err error
	var header databaseHeader

	header.recordLength, err = readShortLittleEnd(inFile)
	check(err)

	header.headerBlockSize, err = readShortLittleEnd(inFile)
	check(err)

	header.fileType, err = readByteLittleEnd(inFile)
	check(err)

	header.dataBlockSizeCode, err = readByteLittleEnd(inFile)
	check(err)

	header.recordCount, err = readLongLittleEnd(inFile)
	check(err)

	header.blocksUsedCount, err = readShortLittleEnd(inFile)
	check(err)

	header.blocksTotalCount, err = readShortLittleEnd(inFile)
	check(err)

	_, err = readShortLittleEnd(inFile) // Throw away the first block code
	check(err)

	header.lastBlockInUse, err = readShortLittleEnd(inFile)
	check(err)

	// Go to the field count
	_, err = inFile.Seek(0x0021, 0)
	check(err)

	header.fieldCount, err = readByteLittleEnd(inFile)
	check(err)

	header.keyFieldsCount, err = readByteLittleEnd(inFile)
	check(err)

	return header, err
}

func fetchBlockHeader(inFile *os.File) (blockHeader, error) {
	var err error
	var header blockHeader

	header.nextBlockNumber, err = readShortLittleEnd(inFile)
	check(err)

	header.prevBlockNumber, err = readShortLittleEnd(inFile)
	check(err)

	header.offsetLastRecord, err = readShortLittleEnd(inFile)
	check(err)

	return header, err

}
func pullFieldDescs(inFile *os.File, header databaseHeader) error {
	// Go to 0x78 to start file lengths

	_, err := inFile.Seek(120, 0)
	check(err)

	//fieldDescs := make([]fieldDescription, header.fieldCount)
	//fields := make(map[byte]fieldDescription)
	var fieldCounter byte
	fieldCounter = 0
	maxCount := header.fieldCount

	// Fetch the type and length

	var currentField fieldDescription

	for fieldCounter < maxCount {
		currentField = fields[fieldCounter]
		currentField.fieldType, err = readByteLittleEnd(inFile)
		//fieldDescs[fieldCounter].fieldType, err = readByteLittleEnd(inFile)
		check(err)

		currentField.length, err = readByteLittleEnd(inFile)
		check(err)
		fields[fieldCounter] = currentField
		fieldCounter++
	}

	// fetch the names
	var offset int64
	offset = int64(203) + int64(header.fieldCount*6)
	_, err = inFile.Seek(offset, 0)
	check(err)

	fieldCounter = 0
	var valueRead byte
	var fieldNameBytes []byte
	for fieldCounter < maxCount {
		currentField = fields[fieldCounter]
		for {
			valueRead, err = readByteLittleEnd(inFile)
			check(err)

			if valueRead == 0x00 {
				break
			} else {
				fieldNameBytes = append(fieldNameBytes, valueRead)
			}
		}
		currentField.name = string(fieldNameBytes)
		fieldNameBytes = fieldNameBytes[:0]
		fields[fieldCounter] = currentField
		fieldCounter++
	}

	return err
}

func printDatabaseHeaderInfo(header databaseHeader) {

	log.Println("Read and report")
	log.Printf("Total Blocks %d", header.blocksTotalCount)
	log.Printf("lastBlock in Use %d", header.lastBlockInUse)
	log.Printf("Fields in Use %d", header.fieldCount)
	log.Printf("Datablock Size Code %d", header.dataBlockSizeCode)

}

func printBlockHeaderInfo(header blockHeader) {
	log.Println("Next Block: ", header.nextBlockNumber)
	log.Println("Prev Block: ", header.prevBlockNumber)
	log.Println("Offset last Record: ", header.offsetLastRecord)

}

func fetchBlockRecords(maxOffset int64, inFile *os.File) (int64, error) {
	var fieldIndex byte
	var currentOffset int64
	var err error
	var recCount int64

	recCount = 0

	for currentOffset <= maxOffset {
		for fieldIndex < byte(len(fields)) {
			field := fields[fieldIndex]
			input := make([]byte, field.length)
			_, err = inFile.Read(input)
			check(err)

			// log.Printf("%d %s : %s", fieldIndex, field.name, input)

			fieldIndex++
		}
		recCount++

		fieldIndex = 0
		currentOffset, err = inFile.Seek(0, 1)
		check(err)

	}

	log.Printf("Record Count : %d", recCount)
	return currentOffset, err

}

func main() {
	log.Println("Opening File")

	inFile, err := os.Open(sampleFileName)
	check(err)

	defer inFile.Close()

	// Go get the database header
	dbDatabaseHead, err := setupDatabaseHeader(inFile)
	check(err)

	// Pull the Field Descriptions
	fields = make(map[byte]fieldDescription)
	err = pullFieldDescs(inFile, dbDatabaseHead)
	check(err)

	//printDatabaseHeaderInfo(dbDatabaseHead)

	var currentOffset int64

	currentOffset = int64(dbDatabaseHead.headerBlockSize)
	_, err = inFile.Seek(currentOffset, 0)
	check(err)

	//var strRead string

	var blockHead blockHeader
	blockHead, err = fetchBlockHeader(inFile)
	check(err)

	//printBlockHeaderInfo(blockHead)

	currentOffset, err = inFile.Seek(0, 1)
	var blockOffset int64
	blockOffset = currentOffset

	check(err)

	log.Printf("current offset : %d\n", currentOffset)
	log.Printf("offset last record : %d\n", blockHead.offsetLastRecord)

	//	for blockHead.nextBlockNumber > 0 {
	for {
		maxOffset := blockOffset + int64(blockHead.offsetLastRecord)
		currentOffset, err = fetchBlockRecords(maxOffset, inFile)
		log.Printf("Current Offset %x\n", currentOffset)
		check(err)

		log.Printf("Next block Number Test %d\n", blockHead.nextBlockNumber)
		log.Printf("Header block Size: %d", dbDatabaseHead.headerBlockSize)
		var totalBlockSize int64
		totalBlockSize = int64(dbDatabaseHead.dataBlockSizeCode) * 1024

		log.Printf("total block size: %d", totalBlockSize)

		currentOffset = int64(dbDatabaseHead.headerBlockSize) + (int64(blockHead.nextBlockNumber-1) * int64(totalBlockSize))
		blockOffset = currentOffset

		_, err = inFile.Seek(currentOffset, 0)
		check(err)

		log.Printf("Current Offset:  %x \n ", currentOffset)

		blockHead, err = fetchBlockHeader(inFile)
		check(err)

		printBlockHeaderInfo(blockHead)

		if blockHead.nextBlockNumber == 0 {
			maxOffset := blockOffset + int64(blockHead.offsetLastRecord)
			_, err = fetchBlockRecords(maxOffset, inFile)
			check(err)

			break
		}
	}

	q.Q(dbDatabaseHead)
	q.Q(blockHead)
	q.Q(fields)

}
