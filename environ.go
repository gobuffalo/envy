package envy

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gobuffalo/syncx"
	"github.com/joho/godotenv"
	"github.com/rogpeppe/go-internal/modfile"
)

type Env struct {
	env syncx.StringMap
}

func New() Env {
	e := Env{
		env: syncx.StringMap{},
	}
	e.Load()
	e.loadEnv()
	return e
}

func (e *Env) loadEnv() {
	if os.Getenv("GO_ENV") == "" {
		// if the flag "test.v" is *defined*, we're running as a unit test. Note that we don't care
		// about v.Value (verbose test mode); we just want to know if the test environment has defined
		// it. It's also possible that the flags are not yet fully parsed (i.e. flag.Parsed() == false),
		// so we could not depend on v.Value anyway.
		//
		if v := flag.Lookup("test.v"); v != nil {
			e.env.Store("GO_ENV", "test")
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

	for _, x := range os.Environ() {
		pair := strings.Split(x, "=")
		e.env.Store(pair[0], os.Getenv(pair[0]))
	}
}

func (e Env) Mods() bool {
	x, ok := e.env.Load(GO111MODULE)
	if !ok {
		return false
	}
	return x == "on"
}

// Reload the ENV variables. Useful if
// an external ENV manager has been used
func (e *Env) Reload() {
	e.env = syncx.StringMap{}
	e.loadEnv()
}

// Load .env files. Files will be loaded in the same order that are received.
// Redefined vars will override previously existing values.
// IE: envy.Load(".env", "test_env/.env") will result in DIR=test_env
// If no arg passed, it will try to load a .env file.
func (e *Env) Load(files ...string) error {

	// If no files received, load the default one
	if len(files) == 0 {
		err := godotenv.Overload()
		if err == nil {
			e.Reload()
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
		e.Reload()

	}
	return nil
}

// Get a value from the ENV. If it doesn't exist the
// default value will be returned.
func (e Env) Get(key string, value string) string {
	if v, ok := e.env.Load(key); ok {
		return v
	}
	return value
}

// Get a value from the ENV. If it doesn't exist
// an error will be returned
func (e Env) MustGet(key string) (string, error) {
	if v, ok := e.env.Load(key); ok {
		return v, nil
	}
	return "", fmt.Errorf("could not find ENV var with %s", key)
}

// Set a value into the ENV. This is NOT permanent. It will
// only affect values accessed through envy.
func (e *Env) Set(key string, value string) {
	e.env.Store(key, value)
}

// MustSet the value into the underlying ENV, as well as envy.
// This may return an error if there is a problem setting the
// underlying ENV value.
func (e *Env) MustSet(key string, value string) error {
	if err := os.Setenv(key, value); err != nil {
		return err
	}
	e.env.Store(key, value)
	return nil
}

// Map all of the keys/values set in envy.
func (e Env) Map() map[string]string {
	cp := map[string]string{}
	e.env.Range(func(k, v string) bool {
		cp[k] = v
		return true
	})
	return cp
}

// Temp makes a copy of the values and allows operation on
// those values temporarily during the run of the function.
// At the end of the function run the copy is discarded and
// the original values are replaced. This is useful for testing.
// Warning: This function is NOT safe to use from a goroutine or
// from code which may access any Get or Set function from a goroutine
func (e *Env) Temp(f func()) {
	oenv := e.env
	fmt.Println("### oenv ->", oenv)
	defer func() { e.env = oenv }()
	e.env = syncx.StringMap{}
	oenv.Range(func(k, v string) bool {
		e.env.Store(k, v)
		return true
	})
	f()
}

func (e Env) GoPath() string {
	return e.Get("GOPATH", "")
}

func (e Env) GoBin() string {
	return e.Get("GO_BIN", "go")
}

func (e Env) InGoPath() bool {
	pwd, _ := os.Getwd()
	for _, p := range e.GoPaths() {
		if strings.HasPrefix(pwd, p) {
			return true
		}
	}
	return false
}

// GoPaths returns all possible GOPATHS that are set.
func (e Env) GoPaths() []string {
	gp := e.Get("GOPATH", "")
	if runtime.GOOS == "windows" {
		return strings.Split(gp, ";") // Windows uses a different separator
	}
	return strings.Split(gp, ":")
}

func (e Env) importPath(path string) string {
	path = strings.TrimPrefix(path, "/private")
	for _, gopath := range e.GoPaths() {
		srcpath := filepath.Join(gopath, "src")
		rel, err := filepath.Rel(srcpath, path)
		if err == nil {
			return filepath.ToSlash(rel)
		}
	}

	// fallback to trim
	rel := strings.TrimPrefix(path, filepath.Join(GoPath(), "src"))
	rel = strings.TrimPrefix(rel, string(filepath.Separator))
	return filepath.ToSlash(rel)
}

// CurrentModule will attempt to return the module name from `go.mod` if
// modules are enabled.
// If modules are not enabled it will fallback to using CurrentPackage instead.
func (e Env) CurrentModule() (string, error) {
	if !e.Mods() {
		return e.CurrentPackage(), nil
	}
	moddata, err := ioutil.ReadFile("go.mod")
	if err != nil {
		return "", errors.New("go.mod cannot be read or does not exist while go module is enabled")
	}
	packagePath := modfile.ModulePath(moddata)
	if packagePath == "" {
		return "", errors.New("go.mod is malformed")
	}
	return packagePath, nil
}

// CurrentPackage attempts to figure out the current package name from the PWD
// Use CurrentModule for a more accurate package name.
func (e Env) CurrentPackage() string {
	pwd, _ := os.Getwd()
	return e.importPath(pwd)
}

func (e Env) Environ() []string {
	var x []string
	e.env.Range(func(k, v string) bool {
		x = append(x, fmt.Sprintf("%s=%s", k, v))
		return true
	})
	return x
}
