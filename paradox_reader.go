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

func setupHeaderData(inFile *os.File) (databaseHeader, error) {
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

func main() {
	log.Println("Opening File")

	inFile, err := os.Open(sampleFileName)
	check(err)

	defer inFile.Close()

	// Go get the database header
	header, err := setupHeaderData(inFile)
	check(err)

	// Pull the Field Descriptions
	fields = make(map[byte]fieldDescription)
	err = pullFieldDescs(inFile, header)
	check(err)

	log.Println("Read and report")
	log.Printf("Total Blocks %d", header.blocksTotalCount)
	log.Printf("lastBlock in Use %d", header.lastBlockInUse)
	log.Printf("Fields in Use %d", header.fieldCount)
	log.Printf("Datablock Size Code %d", header.dataBlockSizeCode)

	var currentOffset int64

	currentOffset = int64(header.headerBlockSize)
	_, err = inFile.Seek(currentOffset, 0)
	check(err)

	nextBlockNumber, err := readShortLittleEnd(inFile)
	check(err)

	prevBlockNumber, err := readShortLittleEnd(inFile)
	check(err)

	offsetLastRecord, err := readShortLittleEnd(inFile)
	check(err)

	log.Println("Next Block: ", nextBlockNumber)
	log.Println("Prev Block: ", prevBlockNumber)
	log.Println("Offset last Record: ", offsetLastRecord)

	//var strRead string

	currentOffset, err = inFile.Seek(0, 1)
	var blockOffset int64
	blockOffset = currentOffset

	check(err)
	var fieldIndex byte

	log.Printf("current offset : %d\n", currentOffset)
	log.Printf("offset last record : %d\n", offsetLastRecord)

	for nextBlockNumber > 0 {
		for currentOffset <= (blockOffset + int64(offsetLastRecord)) {
			for fieldIndex < byte(len(fields)) {
				field := fields[fieldIndex]
				input := make([]byte, field.length)
				_, err = inFile.Read(input)
				check(err)

				log.Printf("%d %s : %s", fieldIndex, field.name, input)

				fieldIndex++
			}
			fieldIndex = 0
			currentOffset, err = inFile.Seek(0, 1)
			check(err)

		}

		log.Printf("Next block Number Test %d\n", nextBlockNumber)
		log.Printf("Header block Size: %d", header.headerBlockSize)
		var totalBlockSize int64
		totalBlockSize = int64(header.dataBlockSizeCode) * 1024

		log.Printf("total block size: %d", totalBlockSize)

		currentOffset = int64(header.headerBlockSize) + (int64(nextBlockNumber-1) * int64(totalBlockSize))
		blockOffset = currentOffset

		_, err = inFile.Seek(currentOffset, 0)
		check(err)

		log.Printf("Current Offset:  %x \n ", currentOffset)

		nextBlockNumber, err = readShortLittleEnd(inFile)
		check(err)

		prevBlockNumber, err = readShortLittleEnd(inFile)
		check(err)

		offsetLastRecord, err = readShortLittleEnd(inFile)
		check(err)
		log.Println("Next Block: ", nextBlockNumber)
		log.Println("Prev Block: ", prevBlockNumber)
		log.Println("Offset last Record: ", offsetLastRecord)
		check(err)
	}

	q.Q(header)
	q.Q(fields)

}
