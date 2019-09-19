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
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"golang.org/x/tools/go/packages"
)

type Environment struct {
	*packages.Package
	pairs  *envMap
	goPath string
	goMod  string
	goBin  string
}

func (e *Environment) Clone() *Environment {
	en := &Environment{
		Package: e.Package,
		pairs:   &envMap{data: &sync.Map{}},
		goPath:  e.goPath,
		goMod:   e.goMod,
		goBin:   e.goBin,
	}

	e.pairs.Range(func(key, value string) bool {
		en.pairs.Store(key, value)
		return true
	})

	return en
}

func New() (*Environment, error) {
	cfg := &packages.Config{}
	pkgs, err := packages.Load(cfg)
	if err != nil {
		return nil, err
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("could not find any packages")
	}

	e := &Environment{}
	e.Package = pkgs[0]

	e.goBin = os.Getenv("GO_BIN")
	if len(e.goBin) == 0 {
		e.goBin = "go"
	}

	if err = e.loadEnv(); err != nil {
		return nil, err
	}

	if err = e.loadGoMod(); err != nil {
		return nil, err
	}
	return e, err
}

func (e *Environment) loadGoMod() error {
	out, err := exec.Command("go", "env", "GOMOD").Output()
	if err == nil {
		gp := strings.TrimSpace(string(out))
		e.goMod = gp
	}
	return nil
}

func (e *Environment) loadEnv() error {
	// set the GOPATH if using >= 1.8 and the GOPATH isn't set
	if os.Getenv("GOPATH") == "" {
		out, err := exec.Command("go", "env", "GOPATH").Output()
		if err == nil {
			gp := strings.TrimSpace(string(out))
			os.Setenv("GOPATH", gp)
			e.goPath = gp
		}
	}

	e.pairs = &envMap{data: &sync.Map{}}
	for _, en := range os.Environ() {
		pair := strings.Split(en, "=")

		e.pairs.Store(pair[0], pair[1])
	}
	return nil
}

func (e *Environment) Mods() bool {
	return len(e.goMod) > 0
}

// Reload the ENV variables. Useful if
// an external ENV manager has been used
func (e *Environment) Reload() error {
	en, err := New()
	if err != nil {
		return err
	}
	(*e) = *en
	return nil
}

// Load .env files. Files will be loaded in the same order that are received.
// Redefined vars will override previously existing values.
// IE: envy.Load(".env", "test_env/.env") will result in DIR=test_env
// If no arg passed, it will try to load a .env file.
func (e *Environment) Load(files ...string) error {
	// If no files received, load the default one
	if len(files) == 0 {
		err := godotenv.Overload()
		if err == nil {
			if err := e.Reload(); err != nil {
				return err
			}
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
		if err := e.Reload(); err != nil {
			return err
		}

	}
	return nil
}

// GetOr a value from the ENV. If it doesn't exist the
// default value will be returned.
func (e *Environment) GetOr(key string, value string) string {
	s, _ := e.pairs.LoadOrStore(key, value)
	return s
}

// MustGet a value from the ENV. If it doesn't exist
// an error will be returned
func (e *Environment) Get(key string) (string, error) {
	s, ok := e.pairs.Load(key)
	if !ok {
		return s, fmt.Errorf("could not find ENV var with %s", key)
	}
	return s, nil
}

// Set a value into the ENV. This is NOT permanent. It will
// only affect values accessed through envy.
func (e *Environment) Set(key string, value string) {
	switch key {
	case "GOPATH":
		e.goPath = value
	case "GOMOD":
		e.goMod = value
	case "GO_BIN":
		e.goBin = value
	case "GO111MODULE":
		if value == "off" {
			e.goMod = ""
		} else {
			e.loadGoMod()
		}
	}
	e.pairs.Store(key, value)
}

// ForceSet the value into the underlying ENV, as well as envy.
// This may return an error if there is a problem setting the
// underlying ENV value.
func (e *Environment) ForceSet(key string, value string) error {
	err := os.Setenv(key, value)
	if err != nil {
		return err
	}
	e.pairs.Store(key, value)

	return nil
}

// Map all of the keys/values set in envy.
func (e *Environment) Map() map[string]string {
	m := map[string]string{}
	e.pairs.Range(func(key, value string) bool {
		m[key] = value
		return true
	})
	return m
}

func (e *Environment) GoPath() string {
	return e.goPath
}

func (e *Environment) GoBin() string {
	return e.goBin
}

func (e *Environment) Environ() []string {
	var en []string
	e.pairs.Range(func(key, value string) bool {
		en = append(en, fmt.Sprintf("%s=%s", key, value))
		return true
	})
	return en
}

// GoPaths returns all possible GOPATHS that are set.
func (e *Environment) GoPaths() []string {
	gp := e.GoPath()
	if runtime.GOOS == "windows" {
		return strings.Split(gp, ";") // Windows uses a different separator
	}
	return strings.Split(gp, ":")
}
