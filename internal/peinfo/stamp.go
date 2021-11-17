package peinfo

import (
	"encoding/binary"
	"io"
	"os"
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

// ReadTimeStamp gets the timestamp, which was written in the binary by compiler when the executable was built from io.ReaderAt.
func ReadTimeStamp(fd io.ReaderAt) (time.Time, error) {
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

// GetTimeStamp gets the timestamp, which was written in the binary by compiler when the executable was built by filename.
func GetTimeStamp(fname string) (time.Time, error) {
	fd, err := os.Open(fname)
	if err != nil {
		return time.Time{}, err
	}
	defer fd.Close()
	return ReadTimeStamp(fd)
}
