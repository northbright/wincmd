package wincmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/northbright/byteorder"
	"github.com/northbright/uuid"
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

// GetTempPath gets the TEMP path on Windows.
func GetTempPath() string {
	return os.Getenv("TEMP")
}

// Run runs a command and returns its combined standard output and standard error in UTF-8.
func Run(name string, args ...string) (string, error) {
	var (
		err         error
		updatedArgs []string
	)

	// Use UUID as random log file name
	randFileName, _ := uuid.New()
	logFile := path.Join(GetTempPath(), fmt.Sprintf("%s.txt", randFileName))

	updatedArgs = append(updatedArgs, name)
	updatedArgs = append(updatedArgs, args...)
	updatedArgs = append(updatedArgs, ">", logFile, "2>&1")

	cmd := exec.Command("powershell.exe", updatedArgs...)

	cmd.Env = append(os.Environ())
	if _, err = cmd.CombinedOutput(); err != nil {
		return "", err
	}

	buf, err := ioutil.ReadFile(logFile)
	if err != nil {
		log.Printf("ReadFile() error: %v", err)
		return "", err
	}

	output, err := DecodeUTF16(buf)
	log.Printf("DecodeUTF16(): %s, error: %v", output, err)
	if err != nil {
		return "", err
	}

	if err = os.Remove(logFile); err != nil {
		return "", err
	}

	return output, nil
}
