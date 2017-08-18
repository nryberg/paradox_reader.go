package main

import (
	"fmt"
	"log"
	"os"
)

const sampleFileName = "/Users/Nick/Dropbox/Develop/Upwork/Paradox/Related/Samples/AREA-PDX/AREACODE.DB"

func main() {
	log.Println("Opening File")

	inFile, err := os.Open(sampleFileName)
	if err != nil {
		log.Println("Failure to open file")
		log.Println(err)
	}
	defer inFile.Close()

	log.Println("Read and report")

	input := make([]byte, 5)
	numRead, err := inFile.Read(input)

	if err != nil {
		log.Println(err)
	}
	fmt.Printf("%d Bytes read: %s\n", numRead, string(input))
}
