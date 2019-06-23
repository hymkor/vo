package main

import (
	"io"
	"time"
)

const FILE_ADDRESS_OF_NEW_EXE_HEADER = 60

func bytes2dword(array []byte) int64 {
	return int64(array[0]) +
		int64(array[1])*256 +
		int64(array[2])*256*256 +
		int64(array[3])*256*256*256
}

func GetPeHeaderPos(fd io.ReadSeeker) (int64, error) {
	var array [4]byte

	_, err := fd.Seek(FILE_ADDRESS_OF_NEW_EXE_HEADER, io.SeekStart)
	if err != nil {
		return 0, err
	}
	_, err = fd.Read(array[:])
	if err != nil {
		return 0, err
	}
	return bytes2dword(array[:]), nil
}

func GetPEStamp(fd io.ReadSeeker) (time.Time, error) {
	var array [4]byte

	peHeaderPos, err := GetPeHeaderPos(fd)
	if err != nil {
		return time.Time{}, err
	}

	_, err = fd.Seek(peHeaderPos+8, io.SeekStart)
	if err != nil {
		return time.Time{}, err
	}

	_, err = fd.Read(array[:])
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(bytes2dword(array[:]), 0), nil
}
