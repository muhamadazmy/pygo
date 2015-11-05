package pygo

import (
	"encoding/binary"
	"encoding/json"
	"os"
)

type Stream interface {
	Write(interface{}) error
	Read() (interface{}, error)
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
	var read uint32 = 0
	for read < length {
		count, err := stream.chanout.Read(bytes[read:])
		read += uint32(count)
		if err != nil {
			return nil, err
		}
	}

	var value interface{}
	err = json.Unmarshal(bytes, &value)

	return value, err
}
