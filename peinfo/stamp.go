package peinfo

import (
	"encoding/binary"
	"io"
	"time"
)

const FILE_ADDRESS_OF_NEW_EXE_HEADER = 60

func getPeHeaderPos(fd io.ReaderAt) (uint32, error) {
	var array [4]byte

	_, err := fd.ReadAt(array[:], FILE_ADDRESS_OF_NEW_EXE_HEADER)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(array[:]), nil
}

func GetTimeStamp(fd io.ReaderAt) (time.Time, error) {
	var array [4]byte

	peHeaderPos, err := getPeHeaderPos(fd)
	if err != nil {
		return time.Time{}, err
	}

	_, err = fd.ReadAt(array[:], int64(peHeaderPos+8))
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(int64(binary.LittleEndian.Uint32(array[:])), 0), nil
}
