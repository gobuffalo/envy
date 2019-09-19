package envy

import (
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

var _ = func() error {
	b, err := ioutil.ReadFile("env")
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.Create(".env")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	f.Write(b)
	return nil
}()

func Test_Environment_Get(t *testing.T) {
	r := require.New(t)

	e, err := New()
	r.NoError(err)

	r.NotZero(os.Getenv("GOPATH"))
	r.Equal(os.Getenv("GOPATH"), e.GetOr("GOPATH", "foo"))
	r.Equal("bar", e.GetOr("IDONTEXIST", "bar"))
}

func Test_Environment_MustGet(t *testing.T) {
	r := require.New(t)

	e, err := New()
	r.NoError(err)

	r.NotZero(os.Getenv("GOPATH"))
	v, err := e.Get("GOPATH")
	r.NoError(err)
	r.Equal(os.Getenv("GOPATH"), v)

	_, err = e.Get("IDONTEXIST")
	r.Error(err)
}

func Test_Environment_Set(t *testing.T) {
	r := require.New(t)

	e, err := New()
	r.NoError(err)

	_, err = e.Get("FOO")
	r.Error(err)

	e.Set("FOO", "foo")
	r.Equal("foo", e.GetOr("FOO", "bar"))
}

func Test_Environment_ForceSet(t *testing.T) {
	r := require.New(t)

	e, err := New()
	r.NoError(err)

	r.Zero(os.Getenv("FOO"))

	err = e.ForceSet("FOO", "BAR")
	r.NoError(err)

	r.Equal("BAR", os.Getenv("FOO"))
}

func Test_Environment_GoPath(t *testing.T) {
	r := require.New(t)

	e, err := New()
	r.NoError(err)

	e.Set("GOPATH", "/foo")
	r.Equal("/foo", e.GoPath())
}

func Test_Environment_GoPaths(t *testing.T) {
	r := require.New(t)

	e, err := New()
	r.NoError(err)

	if runtime.GOOS == "windows" {
		e.Set("GOPATH", "/foo;/bar")
	} else {
		e.Set("GOPATH", "/foo:/bar")
	}
	r.Equal([]string{"/foo", "/bar"}, e.GoPaths())
}

func Test_Environment_CurrentPackage(t *testing.T) {
	r := require.New(t)
	r.Equal("github.com/gobuffalo/envy", CurrentPackage())
}

// Env files loading
func Test_Environment_LoadEnvLoadsEnvFile(t *testing.T) {
	r := require.New(t)

	e, err := New()
	r.NoError(err)

	e.Load()
	r.Equal("root", e.GetOr("DIR", ""))
	r.Equal("none", e.GetOr("FLAVOUR", ""))
	r.Equal("false", e.GetOr("INSIDE_FOLDER", ""))
}

func Test_Environment_LoadDefaultEnvWhenNoArgsPassed(t *testing.T) {
	r := require.New(t)

	e, err := New()
	r.NoError(err)

	err = e.Load()
	r.NoError(err)

	r.Equal("root", e.GetOr("DIR", ""))
	r.Equal("none", e.GetOr("FLAVOUR", ""))
	r.Equal("false", e.GetOr("INSIDE_FOLDER", ""))
}

func Test_Environment_DoNotLoadDefaultEnvWhenArgsPassed(t *testing.T) {
	r := require.New(t)

	e, err := New()
	r.NoError(err)

	err = e.Load("test_env/.env")
	r.NoError(err)

	r.Equal("test_env", e.GetOr("DIR", ""))
	r.Equal("none", e.GetOr("FLAVOUR", ""))
	r.Equal("true", e.GetOr("INSIDE_FOLDER", ""))
}

func Test_Environment_OverloadParams(t *testing.T) {
	r := require.New(t)

	e, err := New()
	r.NoError(err)

	err = e.Load("test_env/.env.test", "test_env/.env.prod")
	r.NoError(err)

	r.Equal("production", e.GetOr("FLAVOUR", ""))
}

func Test_Environment_ErrorWhenSingleFileLoadDoesNotExist(t *testing.T) {
	r := require.New(t)

	e, err := New()
	r.NoError(err)

	e.pairs.Delete("FLAVOUR")
	err = e.Load(".env.fake")

	r.Error(err)
	r.Equal("FAILED", e.GetOr("FLAVOUR", "FAILED"))
}

func Test_Environment_KeepEnvWhenFileInListFails(t *testing.T) {
	r := require.New(t)

	e, err := New()
	r.NoError(err)

	err = e.Load(".env", ".env.FAKE")
	r.Error(err)
	r.Equal("none", e.GetOr("FLAVOUR", "FAILED"))
	r.Equal("root", e.GetOr("DIR", "FAILED"))
}

func Test_Environment_KeepEnvWhenSecondLoadFails(t *testing.T) {
	r := require.New(t)

	e, err := New()
	r.NoError(err)

	err = e.Load(".env")
	r.NoError(err)
	r.Equal("none", e.GetOr("FLAVOUR", "FAILED"))
	r.Equal("root", e.GetOr("DIR", "FAILED"))

	err = e.Load(".env.FAKE")

	r.Equal("none", e.GetOr("FLAVOUR", "FAILED"))
	r.Equal("root", e.GetOr("DIR", "FAILED"))
}

func Test_Environment_StopLoadingWhenFileInListFails(t *testing.T) {
	r := require.New(t)

	e, err := New()
	r.NoError(err)

	err = e.Load(".env", ".env.FAKE", "test_env/.env.prod")
	r.Error(err)
	r.Equal("none", e.GetOr("FLAVOUR", "FAILED"))
	r.Equal("root", e.GetOr("DIR", "FAILED"))
}
