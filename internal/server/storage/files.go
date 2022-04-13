package storage

import (
	"bufio"
	"encoding/json"
	"os"
)

type producer struct {
	file *os.File
	//encoder *json.Encoder
	writer *bufio.Writer
}

func newProducer(fileName string, flag int) (*producer, error) {
	flags := os.O_WRONLY | os.O_CREATE
	//fmt.Println(config)
	if flag != 0 {
		flags |= flag
	}
	//fmt.Println(config)
	file, err := os.OpenFile(fileName, flags, 0777)
	if err != nil {
		return nil, err
	}

	return &producer{
		file: file,
		//encoder: json.NewEncoder(file),
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
	err = p.file.Truncate(0)
	p.file.Seek(0, 0)
	if err != nil {
		return err
	}

	return p.writer.Flush()
	//return p.encoder.Encode(&r)
}
func (p *producer) close() error {
	return p.file.Close()
}

type consumer struct {
	file *os.File
	//decoder *json.Decoder
	reader *bufio.Reader
}

func newConsumer(fileName string) (*consumer, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}
	return &consumer{
		file: file,
		//decoder: json.NewDecoder(file),
		reader: bufio.NewReader(file),
	}, nil
}

func (c *consumer) read(r *repo) error {
	//err := c.writer.Decode(&r)
	//fmt.Println(err, r)
	//return err
	data, err := c.reader.ReadBytes('\n')
	if err != nil {
		return err
	}

	return json.Unmarshal(data, r)

}
func (c *consumer) close() error {
	return c.file.Close()
}
