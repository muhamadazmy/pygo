# pygo
Call python functions from GO

# Install
Make sure you have the pygo module installed

```bash
pip2 install pygo
```

## How to use
create your python module `test.py` to look like this

```py
def add(a, b):
    return a + b
```

Then your GO code.

```go
package main

import (
    "github.com/muhamadazmy/pygo"
    "log"
)

func main() {

    py, err := pygo.NewPy("test", nil)
    if err != nil {
        log.Fatal(err)
    }

    res, err := py.Do("add", map[string]interface{}{
        "a": 2,
        "b": 2,
    })

    if err != nil {
        log.Fatal("Call failed ", err)
    }

    log.Println("Result", res)

    res, err = py.Do("add", map[string]interface{}{
        "a": 3,
        "x": 6.3,
    })

    if err != nil {
        log.Fatal("Call failed ", err)
    }

    log.Println("Result", res)
}
```
