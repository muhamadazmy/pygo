package pygo

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

const (
	code = "import pygo; pygo.run('%s')"
)

var (
	PythonBinary string = "python2.7"
	defaultOpts  PyOpts
)

type response struct {
	value interface{}
	err   error
}

type call struct {
	function string
	args     []interface{}
	kwargs   map[string]interface{}
	response chan *response
}

type pygoImpl struct {
	binPath string
	module  string
	opts    *PyOpts
	ps      *os.Process

	stream  Stream
	chanerr *os.File

	//only filled if process exited.
	stderr string
	state  *os.ProcessState

	channel chan *call
}

type PyOpts struct {
	PythonPath string
}

func NewPy(module string, opts *PyOpts) (Pygo, error) {
	if opts == nil {
		opts = &defaultOpts
	}

	path, err := exec.LookPath(PythonBinary)

	if err != nil {
		return nil, err
	}

	py := &pygoImpl{
		binPath: path,
		opts:    opts,
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

	var env []string = nil

	if py.opts.PythonPath != "" {
		env = []string{fmt.Sprintf("PYTHONPATH=%s", py.opts.PythonPath)}
	}

	attr := &os.ProcAttr{
		Files: []*os.File{nil, nil, stderrWriter, pyIn, pyOut},
		Env:   env,
	}

	ps, err := os.StartProcess(py.binPath, []string{
		PythonBinary,
		"-c",
		fmt.Sprintf(code, py.module)},
		attr)

	if err != nil {
		return err
	}

	py.ps = ps
	py.stream = NewStream(goOut, goIn)
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
		"kwargs":   c.kwargs,
	}

	err := py.stream.Write(data)
	if err != nil {
		response.err = err
		return
	}
	//read response.
	value, err := py.stream.Read()

	response.value = value
	response.err = err
}

func (py *pygoImpl) process() {
	for {
		py.processSingle()
	}
}

func (py *pygoImpl) call(function string, args []interface{}, kwargs map[string]interface{}) (interface{}, error) {
	if py.state != nil {
		return nil, fmt.Errorf("Can't execute python code, python process has exited", py.stderr)
	}

	responseChan := make(chan *response)
	defer close(responseChan)

	call := call{
		function: function,
		args:     args,
		kwargs:   kwargs,
		response: responseChan,
	}
	py.channel <- &call
	response := <-responseChan
	if response.err != nil {
		return nil, response.err
	}
	responseMap := response.value.(map[string]interface{})

	if state, ok := responseMap["state"]; ok {
		if state.(string) == "ERROR" {
			return nil, fmt.Errorf("%v", responseMap["return"])
		}
	}

	return responseMap["return"], nil
}

func (py *pygoImpl) Apply(function string, kwargs map[string]interface{}) (interface{}, error) {
	return py.call(function, nil, kwargs)
}

func (py *pygoImpl) Call(function string, args ...interface{}) (interface{}, error) {
	return py.call(function, args, nil)
}
