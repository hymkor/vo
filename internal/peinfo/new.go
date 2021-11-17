package peinfo

import (
	"crypto/md5"
	"debug/pe"
	"fmt"
	"io"
	"os"
	"strings"
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

type ExeSpec struct {
	Name           string
	Md5Sum         string
	FileVersion    string
	ProductVersion string
	Size           int64
	Stamp          time.Time
	Is64bit        bool
}

func New(fname string) *ExeSpec {
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

	stamp, _ := ReadTimeStamp(fd)

	is64bitFlag, _ := is64bit(fd)

	return &ExeSpec{
		Name:           fname,
		Md5Sum:         fmt.Sprintf("%x", h.Sum(nil)),
		FileVersion:    fileVer,
		ProductVersion: prodVer,
		Size:           size,
		Stamp:          stamp,
		Is64bit:        is64bitFlag,
	}
}

func (spec *ExeSpec) WriteTo(w io.Writer) (int64, error) {
	n1, err := fmt.Fprintln(w, spec.Name)
	if err != nil {
		return int64(n1), err
	}

	var second strings.Builder

	if spec.FileVersion != "" || spec.ProductVersion != "" {
		fmt.Fprintf(&second, "%-18s%-18s",
			spec.FileVersion,
			spec.ProductVersion)
	}
	if !spec.Stamp.IsZero() && spec.Stamp != time.Unix(0, 0) {
		fmt.Fprintf(&second, "%-18s", spec.Stamp.Format("2006-01-02 15:04:05"))
	}
	if spec.Is64bit {
		io.WriteString(&second, " (64)")
	}
	n2 := 0
	if second.Len() > 0 {
		n2, err = fmt.Fprintf(w, "\t%s\n", second.String())
		if err != nil {
			return int64(n1 + n2), err
		}
	}
	n3, err := fmt.Fprintf(w, "\t%d bytes  md5sum:%s\n", spec.Size, spec.Md5Sum)
	return int64(n1 + n2 + n3), err
}
