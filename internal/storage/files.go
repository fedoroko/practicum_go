package storage

import (
	"bufio"
	"encoding/json"
	"os"
)

type producer struct {
	file   *os.File
	writer *bufio.Writer
}

func newProducer(fileName string, flag int) (*producer, error) {
	flags := os.O_WRONLY | os.O_CREATE
	if flag != 0 {
		flags |= flag
	}

	file, err := os.OpenFile(fileName, flags, 0777)
	if err != nil {
		return nil, err
	}

	return &producer{
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}

func (p *producer) write(r *repo) error {
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	if _, err = p.writer.Write(data); err != nil {
		return err
	}
	if err = p.writer.WriteByte('\n'); err != nil {
		return err
	}
	// Очистка файла перед записью
	if err = p.file.Truncate(0); err != nil {
		return err
	}
	// Перевод каретки в начало файла
	if _, err = p.file.Seek(0, 0); err != nil {
		return err
	}

	return p.writer.Flush()
}
func (p *producer) close() error {
	return p.file.Close()
}

type consumer struct {
	file   *os.File
	reader *bufio.Reader
}

func newConsumer(fileName string) (*consumer, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}
	return &consumer{
		file:   file,
		reader: bufio.NewReader(file),
	}, nil
}

func (c *consumer) read(r *repo) error {
	data, err := c.reader.ReadBytes('\n')
	if err != nil {
		return err
	}

	return json.Unmarshal(data, r)

}
func (c *consumer) close() error {
	return c.file.Close()
}
