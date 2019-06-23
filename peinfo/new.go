package peinfo

import (
	"crypto/md5"
	"debug/pe"
	"fmt"
	"io"
	"os"
	"time"
)

func is64bit(r io.ReaderAt) (bool, error) {
	file, err := pe.NewFile(r)
	if err != nil {
		return false, err
	}
	_, ok := file.OptionalHeader.(*pe.OptionalHeader64)
	file.Close()
	return ok, nil
}

type exeSpec struct {
	Name           string
	Md5Sum         string
	FileVersion    string
	ProductVersion string
	Size           int64
	Stamp          time.Time
	Is64bit        bool
}

func New(fname string) *exeSpec {
	fd, err := os.Open(fname)
	if err != nil {
		return nil
	}
	defer fd.Close()

	var size int64

	if stat, err := fd.Stat(); err == nil {
		size = stat.Size()
	}

	var fileVer string
	var prodVer string

	if v, err := GetVersionInfo(fname); err == nil {
		if fv, pv, err := v.Number(); err == nil {
			fileVer = fmt.Sprintf("%d.%d.%d.%d", fv[0], fv[1], fv[2], fv[3])
			prodVer = fmt.Sprintf("%d.%d.%d.%d", pv[0], pv[1], pv[2], pv[3])
		}
	}

	h := md5.New()
	if _, err := io.Copy(h, fd); err != nil {
		return nil
	}

	stamp, _ := GetPEStamp(fd)

	is64bitFlag, _ := is64bit(fd)

	return &exeSpec{
		Name:           fname,
		Md5Sum:         fmt.Sprintf("%x", h.Sum(nil)),
		FileVersion:    fileVer,
		ProductVersion: prodVer,
		Size:           size,
		Stamp:          stamp,
		Is64bit:        is64bitFlag,
	}
}
