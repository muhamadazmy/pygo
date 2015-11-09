package pygo

type Pygo interface {
	Error() string
	Apply(fnc string, kwargs map[string]interface{}) (interface{}, error)
	Call(fnc string, args ...interface{}) (interface{}, error)
	Close()
}
