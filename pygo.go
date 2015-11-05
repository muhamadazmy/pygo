package pygo

type Pygo interface {
	Error() string
	Do(fnc string, kwargs map[string]interface{}) (interface{}, error)
}
