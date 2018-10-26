package wincmd

import (
	//"encoding/binary"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	//"regexp"
	//"strings"
	"syscall"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/northbright/byteorder"
)

var (
	ErrInvalidBytesLen = fmt.Errorf("invalid length of bytes(len % 2 != 0)")
)

func BytesToUTF16(b []byte) ([]uint16, error) {
	var (
		buf   []uint16
		order = byteorder.Get()
	)

	l := len(b)
	if l%2 != 0 {
		return nil, ErrInvalidBytesLen
	}

	for i := 0; i < l; i += 2 {
		u16 := order.Uint16(b[i:])
		buf = append(buf, u16)
	}

	return buf, nil
}

func DecodeUTF16(buf []byte) (string, error) {
	log.Printf("buf: %X", buf)

	l := len(buf)
	if l%2 != 0 {
		return "", fmt.Errorf("Invalid UTF-16 bytes")
	}

	utf8Buf := make([]uint8, 4)
	b := &bytes.Buffer{}

	utf16Buf, err := BytesToUTF16(buf)
	if err != nil {
		return "", err
	}

	runes := utf16.Decode(utf16Buf)
	for _, r := range runes {
		n := utf8.EncodeRune(utf8Buf, r)
		b.Write(utf8Buf[:n])
	}

	return b.String(), nil
}

func GetTempPath() string {
	return os.Getenv("TEMP")
}

// RunCmd runs a command in powershell and return exit code and combined output in UTF-8.
func RunCmd(name string, args ...string) (int, []byte, error) {
	var (
		ws          syscall.WaitStatus
		updatedArgs []string
	)

	logFile := path.Join(GetTempPath(), "wincmd-log.txt")

	updatedArgs = append(updatedArgs, name)
	updatedArgs = append(updatedArgs, args...)
	updatedArgs = append(updatedArgs, ">", logFile)

	cmd := exec.Command("powershell.exe", updatedArgs...)

	buf, err := ioutil.ReadFile(logFile)
	if err != nil {
		log.Printf("ReadFile() error: %v", err)
		return 0, nil, err
	}

	str, err := DecodeUTF16(buf)
	log.Printf("DecodeUTF16(): %s, error: %v", str, err)

	cmd.Env = append(os.Environ())
	output, err := cmd.CombinedOutput()

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			ws = exitError.Sys().(syscall.WaitStatus)
			return ws.ExitStatus(), output, nil
		}
		return 0, output, err
	}
	ws = cmd.ProcessState.Sys().(syscall.WaitStatus)
	return ws.ExitStatus(), output, nil
}
