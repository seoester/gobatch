package gobatch

import (
	"errors"
	"testing"
)

func TestMust(t *testing.T) {
	batch := &MockBatch{}
	if Must(batch, nil) != batch {
		t.Error("Must(batch, nil) != batch")
	}

	var panics bool
	func() {
		defer func() {
			if p := recover(); p != nil {
				panics = true
			}
		}()
		_ = Must(&MockBatch{}, errors.New("error"))
	}()

	if !panics {
		t.Error("Must(batch, err) doesn't panic")
	}
}