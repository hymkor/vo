package peinfo

import (
	"errors"
	"unsafe"

	"golang.org/x/sys/windows"
)

var versionDll = windows.NewLazyDLL("version")
var procGetFileVersionInfoSize = versionDll.NewProc("GetFileVersionInfoSizeW")
var procGetFileVersionInfo = versionDll.NewProc("GetFileVersionInfoW")
var procVerQueryValue = versionDll.NewProc("VerQueryValueW")

type vsFixedFileInfo struct {
	Signature        uint32
	StrucVersion     uint32
	FileVersionMS    uint32
	FileVersionLS    uint32
	ProductVersionMS uint32
	ProductVersionLS uint32
	FileFlagsmask    uint32
	FileFlags        uint32
	FileOs           uint32
	FileType         uint32
	FileSubtype      uint32
	FileDateMS       uint32
	FileDateLS       uint32
}

func lower16bit(n uint32) uint {
	return uint(n & 0xFFFF)
}

func upper16bit(n uint32) uint {
	return uint(n>>16) & 0xFFFF
}

type VersionInfo struct {
	buffer []byte
	size   uintptr
	fname  *uint16
}

func GetVersionInfo(fname string) (*VersionInfo, error) {
	_fname, err := windows.UTF16PtrFromString(fname)
	if err != nil {
		return nil, err
	}
	size, _, err := procGetFileVersionInfoSize.Call(
		uintptr(unsafe.Pointer(_fname)),
		0)
	if size == 0 {
		if err != nil {
			return nil, err
		}
		return nil, errors.New("GetFileVersionInfoSize failed.")
	}

	buffer := make([]byte, size)

	rc, _, err := procGetFileVersionInfo.Call(
		uintptr(unsafe.Pointer(_fname)),
		0,
		size,
		uintptr(unsafe.Pointer(&buffer[0])))

	if rc == 0 {
		if err != nil {
			return nil, err
		}
		return nil, errors.New("GetFileVersioninfo failed.")
	}

	return &VersionInfo{
		buffer: buffer,
		size:   size,
		fname:  _fname,
	}, nil
}

func (vi *VersionInfo) query(key string, f uintptr) (uintptr, error) {
	subBlock, err := windows.UTF16PtrFromString(key)
	if err != nil {
		return 0, err
	}
	var queryLen uintptr
	procVerQueryValue.Call(
		uintptr(unsafe.Pointer(&vi.buffer[0])),
		uintptr(unsafe.Pointer(subBlock)),
		f,
		uintptr(unsafe.Pointer(&queryLen)))

	return queryLen, nil
}

// Number returns executable's File-Version(slice of 4-integers)
// and Product-Version(slice of 4-integers).
func (vi *VersionInfo) Number() (file []uint, product []uint, err error) {
	var f *vsFixedFileInfo

	_, err = vi.query(`\`, uintptr(unsafe.Pointer(&f)))
	if err != nil {
		return nil, nil, err
	}
	return []uint{
			upper16bit(f.FileVersionMS),
			lower16bit(f.FileVersionMS),
			upper16bit(f.FileVersionLS),
			lower16bit(f.FileVersionLS),
		},
		[]uint{
			upper16bit(f.ProductVersionMS),
			lower16bit(f.ProductVersionMS),
			upper16bit(f.ProductVersionLS),
			lower16bit(f.ProductVersionLS),
		},
		nil
}
