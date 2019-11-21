/*
package envy makes working with ENV variables in Go trivial.

* Get ENV variables with default values.
* Set ENV variables safely without affecting the underlying system.
* Temporarily change ENV vars; useful for testing.
* Map all of the key/values in the ENV.
* Loads .env files (by using [godotenv](https://github.com/joho/godotenv/))
* More!
*/
package envy

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"github.com/gobuffalo/here"
	"github.com/joho/godotenv"
)

var gil = &sync.RWMutex{}
var env = map[string]string{}
var her = here.New()

// GO111MODULE is ENV for turning mods on/off
const GO111MODULE = "GO111MODULE"

func init() {
	Load()
	loadEnv()
}

// Load the ENV variables to the env map
func loadEnv() {
	gil.Lock()
	defer gil.Unlock()

	if os.Getenv("GO_ENV") == "" {
		// if the flag "test.v" is *defined*, we're running as a unit test. Note that we don't care
		// about v.Value (verbose test mode); we just want to know if the test environment has defined
		// it. It's also possible that the flags are not yet fully parsed (i.e. flag.Parsed() == false),
		// so we could not depend on v.Value anyway.
		//
		if v := flag.Lookup("test.v"); v != nil {
			env["GO_ENV"] = "test"
		}
	}

	// set the GOPATH if using >= 1.8 and the GOPATH isn't set
	if os.Getenv("GOPATH") == "" {
		out, err := exec.Command("go", "env", "GOPATH").Output()
		if err == nil {
			gp := strings.TrimSpace(string(out))
			os.Setenv("GOPATH", gp)
		}
	}

	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		env[pair[0]] = os.Getenv(pair[0])
	}
}

// Mods returns true if module support is enabled, false otherwise
// See https://github.com/golang/go/wiki/Modules#how-to-install-and-activate-module-support for details
func Mods() bool {
	go111 := Get(GO111MODULE, "")
	if go111 == "off" {
		return false
	}

	info, _ := her.Current()
	return !info.Module.IsZero()
}

// Reload the ENV variables. Useful if
// an external ENV manager has been used
func Reload() {
	env = map[string]string{}
	loadEnv()
}

// Load .env files. Files will be loaded in the same order that are received.
// Redefined vars will override previously existing values.
// IE: envy.Load(".env", "test_env/.env") will result in DIR=test_env
// If no arg passed, it will try to load a .env file.
func Load(files ...string) error {

	// If no files received, load the default one
	if len(files) == 0 {
		err := godotenv.Overload()
		if err == nil {
			Reload()
		}
		return err
	}

	// We received a list of files
	for _, file := range files {

		// Check if it exists or we can access
		if _, err := os.Stat(file); err != nil {
			// It does not exist or we can not access.
			// Return and stop loading
			return err
		}

		// It exists and we have permission. Load it
		if err := godotenv.Overload(file); err != nil {
			return err
		}

		// Reload the env so all new changes are noticed
		Reload()

	}
	return nil
}

// Get a value from the ENV. If it doesn't exist the
// default value will be returned.
func Get(key string, value string) string {
	gil.RLock()
	defer gil.RUnlock()
	if v, ok := env[key]; ok {
		return v
	}
	return value
}

// Get a value from the ENV. If it doesn't exist
// an error will be returned
func MustGet(key string) (string, error) {
	gil.RLock()
	defer gil.RUnlock()
	if v, ok := env[key]; ok {
		return v, nil
	}
	return "", fmt.Errorf("could not find ENV var with %s", key)
}

// Set a value into the ENV. This is NOT permanent. It will
// only affect values accessed through envy.
func Set(key string, value string) {
	gil.Lock()
	defer gil.Unlock()
	env[key] = value
}

// MustSet the value into the underlying ENV, as well as envy.
// This may return an error if there is a problem setting the
// underlying ENV value.
func MustSet(key string, value string) error {
	gil.Lock()
	defer gil.Unlock()
	err := os.Setenv(key, value)
	if err != nil {
		return err
	}
	env[key] = value
	return nil
}

// Map all of the keys/values set in envy.
func Map() map[string]string {
	gil.RLock()
	defer gil.RUnlock()
	cp := map[string]string{}
	for k, v := range env {
		cp[k] = v
	}
	return cp
}

// Temp makes a copy of the values and allows operation on
// those values temporarily during the run of the function.
// At the end of the function run the copy is discarded and
// the original values are replaced. This is useful for testing.
// Warning: This function is NOT safe to use from a goroutine or
// from code which may access any Get or Set function from a goroutine
func Temp(f func()) {
	oenv := env
	env = map[string]string{}
	for k, v := range oenv {
		env[k] = v
	}
	defer func() { env = oenv }()
	f()
}

func GoPath() string {
	return Get("GOPATH", "")
}

func GoBin() string {
	return Get("GO_BIN", "go")
}

func InGoPath() bool {
	pwd, _ := os.Getwd()
	for _, p := range GoPaths() {
		if strings.HasPrefix(pwd, p) {
			return true
		}
	}
	return false
}

// GoPaths returns all possible GOPATHS that are set.
func GoPaths() []string {
	gp := Get("GOPATH", "")
	if runtime.GOOS == "windows" {
		return strings.Split(gp, ";") // Windows uses a different separator
	}
	return strings.Split(gp, ":")
}

func CurrentModule() (string, error) {
	info, err := her.Current()
	if err != nil {
		return "", err
	}
	if info.Module.IsZero() {
		return info.ImportPath, nil
	}
	return info.Module.Path, nil
}

func CurrentPackage() string {
	info, err := her.Current()
	if err != nil {
		panic(err)
	}
	return info.ImportPath
}

func Environ() []string {
	gil.RLock()
	defer gil.RUnlock()
	var e []string
	for k, v := range env {
		e = append(e, fmt.Sprintf("%s=%s", k, v))
	}
	return e
}
