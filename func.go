package migrate

import "os"

func envOrDef(name, def string) (val string) {
	if val = os.Getenv(name); val != "" {
		return val
	}
	return def
}
