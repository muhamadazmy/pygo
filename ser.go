package pygo

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"os"
)

type Stream interface {
	Write(interface{}) error
	Read() (interface{}, error)
	Close()
}

type streamImpl struct {
	chanin  *os.File
	chanout *os.File
}

func NewStream(in *os.File, out *os.File) Stream {
	return &streamImpl{
		chanin:  in,
		chanout: out,
	}
}

func (stream *streamImpl) Write(data interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	length := uint32(len(bytes))

	err = binary.Write(stream.chanin, binary.BigEndian, length)
	if err != nil {
		return err
	}
	_, err = stream.chanin.Write(bytes)
	return err
}

func (stream *streamImpl) Read() (interface{}, error) {
	//read length
	var length uint32
	err := binary.Read(stream.chanout, binary.BigEndian, &length)
	if err != nil {
		return nil, err
	}

	bytes := make([]byte, length)
	_, err = io.ReadFull(stream.chanout, bytes)
	if err != nil {
		return nil, err
	}

	var value interface{}
	err = json.Unmarshal(bytes, &value)

	return value, err
}

func (stream *streamImpl) Close() {
	stream.chanin.Close()
	stream.chanout.Close()
}
