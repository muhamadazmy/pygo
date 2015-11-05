package pygo

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

const (
	binary = "python2.7"
	//code   = "import pygo; pygo.wrap('%s')"
	code = "import os; f = os.fdopen(4, 'w').write('%s')"
)

type response struct {
	value interface{}
	err   error
}

type call struct {
	function string
	args     map[string]interface{}
	response chan *response
}

type pygoImpl struct {
	binPath string
	module  string
	ps      *os.Process

	chanin  *os.File
	chanout *os.File
	chanerr *os.File

	//only filled if process exited.
	stderr string
	state  *os.ProcessState

	channel chan *call
}

func NewPy(module string) (Pygo, error) {
	path, err := exec.LookPath(binary)
	if err != nil {
		return nil, err
	}

	py := &pygoImpl{
		binPath: path,
		module:  module,
		channel: make(chan *call),
	}

	err = py.init()
	if err != nil {
		return nil, err
	}

	go py.wait()
	go py.process()

	return py, nil
}

func (py *pygoImpl) wait() {
	data, err := ioutil.ReadAll(py.chanerr)
	if err != nil {
		log.Println(err)
	}

	py.stderr = string(data)

	state, _ := py.ps.Wait()
	py.state = state
}

//init opes the pipes and start the python process.
func (py *pygoImpl) init() error {
	stderrReader, stderrWriter, err := os.Pipe()

	if err != nil {
		return err
	}

	pyIn, goOut, err := os.Pipe()
	if err != nil {
		return err
	}

	goIn, pyOut, err := os.Pipe()
	if err != nil {
		return err
	}

	attr := &os.ProcAttr{
		Files: []*os.File{nil, nil, stderrWriter, pyIn, pyOut},
	}

	ps, err := os.StartProcess(py.binPath, []string{
		binary,
		"-c",
		fmt.Sprintf(code, py.module)},
		attr)

	if err != nil {
		return err
	}

	py.ps = ps
	py.chanin = goOut
	py.chanout = goIn
	py.chanerr = stderrReader

	return nil
}

func (py *pygoImpl) Error() string {
	return py.stderr
}

func (py *pygoImpl) processSingle() {
	c := <-py.channel

	var response response

	defer func() {
		c.response <- &response
	}()

	data := map[string]interface{}{
		"function": c.function,
		"args":     c.args,
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		response.err = err
		return
	}

	n, err := py.chanin.Write(bytes)
	if err != nil {
		response.err = err
		return
	}

	//read response.
	buffer := make([]byte, 1000)
	n, err = py.chanout.Read(buffer)

	if err != nil && err != io.EOF {
		response.err = err
		return
	}

	response.value = buffer[:n]
}

func (py *pygoImpl) process() {
	for {
		py.processSingle()
	}
}

func (py *pygoImpl) Do(function string, args map[string]interface{}) (interface{}, error) {
	if py.state != nil {
		return nil, fmt.Errorf("Can't execute python code, python process has exited", py.stderr)
	}

	responseChan := make(chan *response)
	call := call{
		function: function,
		args:     args,
		response: responseChan,
	}
	py.channel <- &call
	response := <-responseChan
	return response.value, response.err
}
