package pygo

type Pygo interface {
	Error() string
	Do(fnc string, args map[string]interface{}) (interface{}, error)
}
