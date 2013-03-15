// The maildir package provides an interface to mailboxes in the Maildir format.
package maildir

import (
	"bufio"
	"bytes"
	"io"
	"net/mail"
	"net/textproto"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// The Separator separates a messages unique key from its flags in the filename.
// This should only be changed on operating systems where the colon isn't
// allowed in filenames.
var Separator rune = ':'

// A KeyError occurs when a key matches more or less than one message.
type KeyError struct {
	Key string // the (invalid) key
	N   int    // number of matches (!= 1)
}

func (e *KeyError) Error() string {
	return "maildir: key " + e.Key + " matches " + strconv.Itoa(e.N) + " files."
}

// A Dir represents a single directory in a Maildir mailbox.
type Dir string

// Unseen moves messages from new to cur (they are now "seen") and returns their keys.
// This is the only function in this package that operates on any subdirectory
// other than "cur".
func (d Dir) Unseen() ([]string, error) {
	f, err := os.Open(filepath.Join(string(d), "new"))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	names, err := f.Readdirnames(0)
	if err != nil {
		return nil, err
	}
	var keys []string
	for _, n := range names {
		if n[0] != '.' {
			split := strings.FieldsFunc(n, func(r rune) bool {
				return r == Separator
			})
			keys = append(keys, split[0])
			os.Rename(filepath.Join(string(d), "new", n),
				filepath.Join(string(d), "cur", n+string(Separator)+"2,S"))
		}
	}
	return keys, nil
}

// Keys returns a slice of valid keys to access messages by.
func (d Dir) Keys() ([]string, error) {
	f, err := os.Open(filepath.Join(string(d), "cur/"))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	names, err := f.Readdirnames(0)
	if err != nil {
		return nil, err
	}
	var keys []string
	for _, n := range names {
		if n[0] != '.' {
			split := strings.FieldsFunc(n, func(r rune) bool {
				return r == ':'
			})
			keys = append(keys, split[0])
		}
	}
	return keys, nil
}

// Filename returns the path to the file corresponding to the key.
func (d Dir) Filename(key string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(string(d), "cur", key+"*"))
	if err != nil {
		return "", err
	}
	if n := len(matches); n != 1 {
		return "", &KeyError{key, n}
	}
	return matches[0], nil
}

// Header returns the corresponding mail header to a key.
func (d Dir) Header(key string) (header mail.Header, err error) {
	filename, err := d.Filename(key)
	if err != nil {
		return
	}
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()
	tp := textproto.NewReader(bufio.NewReader(file))
	hdr, err := tp.ReadMIMEHeader()
	if err != nil {
		return
	}
	header = mail.Header(hdr)
	return
}

// Message returns a Message by key.
func (d Dir) Message(key string) (*mail.Message, error) {
	filename, err := d.Filename(key)
	if err != nil {
		return &mail.Message{}, err
	}
	r, err := os.Open(filename)
	if err != nil {
		return &mail.Message{}, err
	}
	defer r.Close()
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, r)
	if err != nil {
		return &mail.Message{}, err
	}
	msg, err := mail.ReadMessage(buf)
	if err != nil {
		return msg, err
	}
	return msg, nil
}