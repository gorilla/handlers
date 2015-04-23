# Description
A light weight logging library designed to take a logging threshold and write all entries above the threshold to a writer interface.

# Features
* Log levels
* Log level thresholds
* Outputs to any writer interface
* Convenience fmt functions
* Custom formatting

# Usage
## Default
```Go
package main

import (
	"os"
   "github.com/monsooncommerce/log"
)

func main() {
	logger := log.New(os.Stdout, log.Debug)
	logger.SetTag("myapp")

   logger.Error("This is an error message")

   err := FetchError()
   logger.Errorf("This is a formatted error message: %v", err.Error())
}
```

Output:

```
2013-10-02T13:41:43Z host.monsooncommerce.com myapp[25702]: ERROR [This is an error message]
2013-10-02T13:41:44Z host.monsooncommerce.com myapp[25702]: ERROR [This is a formatted error message: [fetched an error message]]
```

## Custom Formatter
```Go
package main

import (
	"os"
   "github.com/monsooncommerce/log"
)

type CustomFormat struct {
	customMsg string
}

func main() {
	logger := log.New(os.Stdout, log.Debug)
	f := &CustomFormat{
		customMsg: "Awesome Log Line",
	}
	logger.SetFormatter(f)

   logger.Error("This is an error message")

   err := FetchError()
   logger.Errorf("This is a formatted error message: %v", err.Error())
}

func (f *CustomFormat) Format(l log.Level, args ...interface{}) string {
	return fmt.Sprintf("%s: %s -- %s", f.customMsg, l, args)
}

```

Output:

```
Awesome Log Line: ERROR -- [This is an error message]
Awesome Log Line: ERROR -- [This is a formatted error message: [fetched an error message]]
```


# Default Formatting
```
TIMESTAMP HOSTNAME TAG[PID]: LEVEL MESSAGE
```
* **Timestamp** - RFC3339 formated
* **Hostname** - the FQDN of the host writing the message
* **Tag** - represents the application name
* **PID** - the Unix PID of the running application
* **Level** - one of DEBUG, ERROR, INFO, NOTICE, WARNING, or FATAL
