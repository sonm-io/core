package miner

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

type spoolResponseProtocol struct {
	Error  string `json:"error"`
	Status string `json:"status"`
}

// decodeImagePull detects Error of an image pulling process
// by decoding reply from Docker
// Although Docker should reply with JSON Encoded items
// one per line, in different versions it could vary.
// This decoders can detect error even in mixed replies:
// {"Status": "OK"}\n{"Status": "OK"}
// {"Status": "OK"}{"Error": "error"}
func decodeImagePull(r io.Reader) error {
	more := true

	rd := bufio.NewReader(r)
	for more {
		line, err := rd.ReadBytes('\n')
		switch err {
		case nil:
			// pass
		case io.EOF:
			if len(line) == 0 {
				return nil
			}
			more = false
		default:
			return err
		}

		if len(line) == 0 {
			return fmt.Errorf("Empty response line")
		}

		if line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
		}

		if err = decodePullLine(line); err != nil {
			return err
		}
	}
	return nil
}

func decodePullLine(line []byte) error {
	var resp spoolResponseProtocol
	decoder := json.NewDecoder(bytes.NewReader(line))
	for {
		if err := decoder.Decode(&resp); err != nil {
			if err == io.EOF {
				return nil
			}

			return err
		}

		if len(resp.Error) != 0 {
			return fmt.Errorf(resp.Error)
		}
	}
}
