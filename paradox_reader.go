package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

//const sampleFileName = "/Users/Nick/Dropbox/Develop/Upwork/Paradox/Related/Samples/AREA-PDX/AREACODE.DB"

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

	log.Println("-- Database Header --")
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

func writeDataToFile(record string, outFileName string) error {
	//TODO: Move to a more agnostic form of output.  This could be
	// 			a CSV file (for now) or it could be STDOUT.  Or ??

	outFile, err := os.OpenFile(outFileName, os.O_APPEND|os.O_WRONLY, 0600)

	check(err)

	defer outFile.Close()
	_, err = outFile.WriteString(record)
	check(err)

	return err
}

func fetchBlockRecords(maxOffset int64, inFile *os.File, outFileName string) (int64, error) {
	var fieldIndex byte
	var currentOffset int64
	var err error
	var recCount int64
	var record string

	recCount = 0

	// TODO:  The last record in each block is getting dropped
	// 				NEED TO FIX THIS FIRST

	for currentOffset <= maxOffset {
		for fieldIndex < byte(len(fields)) {
			field := fields[fieldIndex]
			input := make([]byte, field.length)
			_, err = inFile.Read(input)
			check(err)
			input = bytes.Trim(input, "\x00") // trim nulls

			// TODO : pick a more agnostic output format for this stage of the game
			// 				and then flip to correct CSV handling after that.

			if strings.Contains(string(input), ",") { // wraps it in quotes if comma already in text.
				record = record + strconv.Quote(string(input)) + "," // append field data
			} else {
				record = record + string(input) + "," // append field data
			}

			fieldIndex++
		}

		record = record[0 : len(record)-1] // trim the last communicate
		record = record + "\n"             // Add a new line

		log.Println(record)
		err = writeDataToFile(record, outFileName)
		check(err)

		record = ""
		recCount++

		fieldIndex = 0
		currentOffset, err = inFile.Seek(0, 1)
		check(err)

	}

	return currentOffset, err

}

func sendFieldNamesToFile(fields map[byte]fieldDescription, outputFileName string) error {
	var outString string
	numFields := len(fields)

	for i := 0; i < numFields; i++ {
		field := fields[byte(i)]
		outString = outString + field.name + ","
	}

	outString = outString[0 : len(outString)-1]
	outString = outString + "\n"
	err := ioutil.WriteFile(outputFileName, []byte(outString), 0644)
	check(err)

	return err
}

func main() {
	inputFileNamePtf := flag.String("i", "", "The input file name for processing")
	outputFileNamePtf := flag.String("o", "./output.csv", "The output file name to push the data to")

	var currentOffset int64

	flag.Parse()

	log.Printf("Opening File : %s", *inputFileNamePtf)

	inFile, err := os.Open(*inputFileNamePtf)
	check(err)

	defer inFile.Close()

	dbDatabaseHead, err := setupDatabaseHeader(inFile) // Go get the database header
	check(err)

	fields = make(map[byte]fieldDescription) // Pull the Field Descriptions
	err = pullFieldDescs(inFile, dbDatabaseHead)
	check(err)

	// printDatabaseHeaderInfo(dbDatabaseHead)

	err = sendFieldNamesToFile(fields, *outputFileNamePtf)
	check(err)

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

	for {
		maxOffset := blockOffset + int64(blockHead.offsetLastRecord)
		_, err = fetchBlockRecords(maxOffset, inFile, *outputFileNamePtf)
		check(err)

		var totalBlockSize int64
		totalBlockSize = int64(dbDatabaseHead.dataBlockSizeCode) * 1024

		currentOffset = int64(dbDatabaseHead.headerBlockSize) + (int64(blockHead.nextBlockNumber-1) * int64(totalBlockSize))
		blockOffset = currentOffset

		_, err = inFile.Seek(currentOffset, 0)
		check(err)

		blockHead, err = fetchBlockHeader(inFile)
		check(err)

		if blockHead.nextBlockNumber == 0 {
			maxOffset := blockOffset + int64(blockHead.offsetLastRecord)
			_, err = fetchBlockRecords(maxOffset, inFile, *outputFileNamePtf)
			check(err)

			break
		}
	}
}
