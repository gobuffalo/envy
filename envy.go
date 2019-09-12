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
	"fmt"
	"os"
	"strings"
)

var env = func() *Environment {
	e := &Environment{}
	if err := e.Load(); err != nil {
		fmt.Println(">>>TODO envy.go:101: err ", err)
	}
	return e
}()

// Mods returns true if module support is enabled, false otherwise
// See https://github.com/golang/go/wiki/Modules#how-to-install-and-activate-module-support for details
func Mods() bool {
	return env.Mods()
}

// Reload the ENV variables. Useful if
// an external ENV manager has been used
func Reload() {
	env.Reload()
}

// Load .env files. Files will be loaded in the same order that are received.
// Redefined vars will override previously existing values.
// IE: envy.Load(".env", "test_env/.env") will result in DIR=test_env
// If no arg passed, it will try to load a .env file.
func Load(files ...string) error {
	return env.Load(files...)
}

// Get a value from the ENV. If it doesn't exist the
// default value will be returned.
func Get(key string, value string) string {
	return env.GetOr(key, value)
}

// Get a value from the ENV. If it doesn't exist
// an error will be returned
func MustGet(key string) (string, error) {
	return env.Get(key)
}

// Set a value into the ENV. This is NOT permanent. It will
// only affect values accessed through envy.
func Set(key string, value string) {
	env.Set(key, value)
}

// MustSet the value into the underlying ENV, as well as envy.
// This may return an error if there is a problem setting the
// underlying ENV value.
func MustSet(key string, value string) error {
	return env.ForceSet(key, value)
}

// Map all of the keys/values set in envy.
func Map() map[string]string {
	return env.Map()
}

// Temp makes a copy of the values and allows operation on
// those values temporarily during the run of the function.
// At the end of the function run the copy is discarded and
// the original values are replaced. This is useful for testing.
// Warning: This function is NOT safe to use from a goroutine or
// from code which may access any Get or Set function from a goroutine
func Temp(f func()) {
	oenv := env
	env = oenv.clone()
	defer func() { env = oenv }()
	f()
}

func GoPath() string {
	return env.GoPath()
}

func GoBin() string {
	return env.GoBin()
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
	return env.GoPaths()
}

// CurrentModule will attempt to return the module name from `go.mod` if
// modules are enabled.
// If modules are not enabled it will fallback to using CurrentPackage instead.
func CurrentModule() (string, error) {
	return env.PkgPath, nil
}

// CurrentPackage attempts to figure out the current package name from the PWD
// Use CurrentModule for a more accurate package name.
func CurrentPackage() string {
	return env.PkgPath
}

func Environ() []string {
	return env.Environ()
}
