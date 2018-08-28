package sonm

import (
	"bytes"
	"io"
)

type logReader struct {
	cli      TaskManagement_LogsClient
	buf      bytes.Buffer
	finished bool
}

func NewLogReader(client TaskManagement_LogsClient) io.Reader {
	return &logReader{cli: client}
}

func (m *logReader) Read(p []byte) (n int, err error) {
	if len(p) > m.buf.Len() && !m.finished {
		chunk, err := m.cli.Recv()
		if err == io.EOF {
			m.finished = true
		} else if err != nil {
			return 0, err
		}
		if chunk != nil && chunk.Data != nil {
			m.buf.Write(chunk.Data)
		}
	}
	return m.buf.Read(p)
}
