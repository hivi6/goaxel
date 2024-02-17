package goaxel

import (
	"encoding/binary"
	"fmt"
	"os"
)

type MetadataRange struct {
	start, stop, current uint64
}

type MetadataInfo struct {
	connections uint64

	ranges []MetadataRange
}

func CreateMetadata(metadataFilename string) error {
	file, err := os.OpenFile(metadataFilename, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("something went wrong while creating metadata file '%v'", metadataFilename)
	}
	defer file.Close()
	return nil
}

func DestroyMetadata(metadataFilename string) error {
	if err := os.Remove(metadataFilename); err != nil {
		return fmt.Errorf("something went wrong while destroying metadata file '%v'", metadataFilename)
	}
	return nil
}

func NewMetadata(conn uint64) MetadataInfo {
	var metadata MetadataInfo
	metadata.connections = conn
	metadata.ranges = make([]MetadataRange, conn)
	return metadata
}

func ReadMetadata(metadataFilename string) (MetadataInfo, error) {
	file, err := os.OpenFile(metadataFilename, os.O_RDONLY, 0666)
	if err != nil {
		return MetadataInfo{}, fmt.Errorf("something went wrong while opening metadata file '%v'", metadataFilename)
	}
	defer file.Close()

	resMeta := MetadataInfo{}

	// read the number of connections
	buffer := make([]byte, 8)
	if _, err := file.Read(buffer); err != nil {
		return MetadataInfo{}, fmt.Errorf("something went wrong while reading metadata file '%v'", metadataFilename)
	}
	resMeta.connections = binary.LittleEndian.Uint64(buffer)
	resMeta.ranges = make([]MetadataRange, resMeta.connections)

	// read all start stop and current
	for i := 0; i < int(resMeta.connections); i++ {
		for j := 0; j < 3; j++ {
			if _, err := file.Read(buffer); err != nil {
				return MetadataInfo{}, fmt.Errorf("something went wrong while reading metadata file '%v'", metadataFilename)
			}
			switch j {
			case 0:
				resMeta.ranges[i].start = binary.LittleEndian.Uint64(buffer)
			case 1:
				resMeta.ranges[i].stop = binary.LittleEndian.Uint64(buffer)
			case 2:
				resMeta.ranges[i].current = binary.LittleEndian.Uint64(buffer)
			}
		}
	}

	return resMeta, nil
}

func WriteMetadata(metadataFilename string, metadata MetadataInfo) error {
	file, err := os.OpenFile(metadataFilename, os.O_WRONLY, 0666)
	if err != nil {
		return fmt.Errorf("something went wrong while opening metadata file '%v'", metadataFilename)
	}
	defer file.Close()

	// convert metadata to binary
	buffer := []byte{}
	buffer = binary.LittleEndian.AppendUint64(buffer, metadata.connections)
	for i := 0; i < int(metadata.connections); i++ {
		buffer = binary.LittleEndian.AppendUint64(buffer, metadata.ranges[i].start)
		buffer = binary.LittleEndian.AppendUint64(buffer, metadata.ranges[i].stop)
		buffer = binary.LittleEndian.AppendUint64(buffer, metadata.ranges[i].current)
	}

	// write the metadata
	if _, err = file.Write(buffer); err != nil {
		return fmt.Errorf("something went wrong while writing metadata file '%v'", metadataFilename)
	}

	// rename the metadata file
	os.Rename(file.Name(), metadataFilename)

	return nil
}
