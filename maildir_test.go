package maildir

import (
	"os"
	"testing"

	"io/ioutil"
)

// cleanup removes a Dir's directory structure
func cleanup(t *testing.T, d Dir) {

	err := os.RemoveAll(string(d))
	if err != nil {
		t.Error(err)
	}
}

// exists checks if the given path exists
func exists(path string) bool {

	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	}

	panic(err)
}

// cat returns the content of a file as a string
func cat(t *testing.T, path string) string {

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	c, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	return string(c)
}

// makeDelivery creates a new message
func makeDelivery(t *testing.T, d Dir, msg string) {

	del, err := d.NewDelivery()
	if err != nil {
		t.Fatal(err)
	}

	err = del.Write([]byte(msg))
	if err != nil {
		t.Fatal(err)
	}

	err = del.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreate(t *testing.T) {

	t.Parallel()

	var d Dir = "test_create"
	err := d.Create()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup(t, d)
}

func TestDelivery(t *testing.T) {

	t.Parallel()

	var d Dir = "test_delivery"
	msgs := []string{
		"this is the first message",
		"a second message follows",
		"why not have three messages?",
	}

	err := d.Create()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup(t, d)

	// Deliver all prepared messages.
	for _, msg := range msgs {
		makeDelivery(t, d, msg)
	}

	keys, err := d.Unseen()
	if err != nil {
		t.Fatal(err)
	}

	// Check if we see all delivered messages in
	// cur directory after calling d.Unseen().
	if len(keys) != len(msgs) {
		t.Fatal("Amount of unseen messages does not concur with delivered messages")
	}

	for i, msg := range msgs {

		path, err := d.Filename(keys[i])
		if err != nil {
			t.Fatal(err)
		}

		if !exists(path) {
			t.Fatal("File doesn't exist")
		}

		if cat(t, path) != msg {
			t.Fatal("Content doesn't match")
		}
	}
}

func TestPurge(t *testing.T) {

	t.Parallel()

	var d Dir = "test_purge"

	err := d.Create()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup(t, d)

	makeDelivery(t, d, "foo")

	keys, err := d.Unseen()
	if err != nil {
		t.Fatal(err)
	}

	path, err := d.Filename(keys[0])
	if err != nil {
		t.Fatal(err)
	}

	err = d.Purge(keys[0])
	if err != nil {
		t.Fatal(err)
	}

	if exists(path) {
		t.Fatal("File still exists")
	}
}

func TestMove(t *testing.T) {

	t.Parallel()

	var d1 Dir = "test_move1"
	var d2 Dir = "test_move2"
	const msg = "a moving message"

	err := d1.Create()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup(t, d1)

	err = d2.Create()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup(t, d2)

	makeDelivery(t, d1, msg)

	keys, err := d1.Unseen()
	if err != nil {
		t.Fatal(err)
	}

	err = d1.Move(d2, keys[0])
	if err != nil {
		t.Fatal(err)
	}

	keys, err = d2.Keys()
	if err != nil {
		t.Fatal(err)
	}

	path, err := d2.Filename(keys[0])
	if err != nil {
		t.Fatal(err)
	}

	if cat(t, path) != msg {
		t.Fatal("Content doesn't match")
	}

}
