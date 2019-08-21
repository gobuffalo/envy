package envy

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
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

func Test_Mods(t *testing.T) {
	r := require.New(t)
	Temp(func() {
		Set(GO111MODULE, "off")
		r.False(Mods())
		Set(GO111MODULE, "on")
		r.True(Mods())
		Set(GO111MODULE, "auto")
		r.Equal(!InGoPath(), Mods())
		Set(GO111MODULE, "")
		r.Equal(!InGoPath(), Mods())
	})
}

// envy should detect when running as a unit test and return GO_ENV=test if otherwise undefined
func Test_GO_ENVUnitTest(t *testing.T) {
	r := require.New(t)
	r.Zero(os.Getenv("GO_ENV"))
	r.Equal("test", Get("GO_ENV", "foo"))
}

func Test_Get(t *testing.T) {
	r := require.New(t)
	r.NotZero(os.Getenv("GOPATH"))
	r.Equal(os.Getenv("GOPATH"), Get("GOPATH", "foo"))
	r.Equal("bar", Get("IDONTEXIST", "bar"))
}

func Test_MustGet(t *testing.T) {
	r := require.New(t)
	r.NotZero(os.Getenv("GOPATH"))
	v, err := MustGet("GOPATH")
	r.NoError(err)
	r.Equal(os.Getenv("GOPATH"), v)

	_, err = MustGet("IDONTEXIST")
	r.Error(err)
}

func Test_Set(t *testing.T) {
	r := require.New(t)
	_, err := MustGet("FOO")
	r.Error(err)

	Set("FOO", "foo")
	r.Equal("foo", Get("FOO", "bar"))
}

func Test_MustSet(t *testing.T) {
	r := require.New(t)

	r.Zero(os.Getenv("FOO"))

	err := MustSet("FOO", "BAR")
	r.NoError(err)

	r.Equal("BAR", os.Getenv("FOO"))
}

func Test_Temp(t *testing.T) {
	r := require.New(t)

	_, err := MustGet("BAR")
	r.Error(err)

	Temp(func() {
		Set("BAR", "foo")
		r.Equal("foo", Get("BAR", "bar"))
		_, err = MustGet("BAR")
		r.NoError(err)
	})

	_, err = MustGet("BAR")
	r.Error(err)
}

func Test_GoPath(t *testing.T) {
	r := require.New(t)
	Temp(func() {
		Set("GOPATH", "/foo")
		r.Equal("/foo", GoPath())
	})
}

func Test_GoPaths(t *testing.T) {
	r := require.New(t)
	Temp(func() {
		if runtime.GOOS == "windows" {
			Set("GOPATH", "/foo;/bar")
		} else {
			Set("GOPATH", "/foo:/bar")
		}
		r.Equal([]string{"/foo", "/bar"}, GoPaths())
	})
}

func Test_CurrentPackage(t *testing.T) {
	r := require.New(t)
	r.Equal("github.com/gobuffalo/envy", CurrentPackage())
}

// Env files loading
func Test_LoadEnvLoadsEnvFile(t *testing.T) {
	r := require.New(t)
	Temp(func() {
		r.Equal("root", Get("DIR", ""))
		r.Equal("none", Get("FLAVOUR", ""))
		r.Equal("false", Get("INSIDE_FOLDER", ""))
	})
}

func Test_LoadDefaultEnvWhenNoArgsPassed(t *testing.T) {
	r := require.New(t)
	Temp(func() {
		err := Load()
		r.NoError(err)

		r.Equal("root", Get("DIR", ""))
		r.Equal("none", Get("FLAVOUR", ""))
		r.Equal("false", Get("INSIDE_FOLDER", ""))
	})
}

func Test_DoNotLoadDefaultEnvWhenArgsPassed(t *testing.T) {
	r := require.New(t)
	Temp(func() {
		err := Load("test_env/.env")
		r.NoError(err)

		r.Equal("test_env", Get("DIR", ""))
		r.Equal("none", Get("FLAVOUR", ""))
		r.Equal("true", Get("INSIDE_FOLDER", ""))
	})
}

func Test_OverloadParams(t *testing.T) {
	r := require.New(t)
	Temp(func() {
		err := Load("test_env/.env.test", "test_env/.env.prod")
		r.NoError(err)

		r.Equal("production", Get("FLAVOUR", ""))
	})
}

func Test_ErrorWhenSingleFileLoadDoesNotExist(t *testing.T) {
	r := require.New(t)
	Temp(func() {
		delete(env, "FLAVOUR")
		err := Load(".env.fake")

		r.Error(err)
		r.Equal("FAILED", Get("FLAVOUR", "FAILED"))
	})
}

func Test_KeepEnvWhenFileInListFails(t *testing.T) {
	r := require.New(t)
	Temp(func() {
		err := Load(".env", ".env.FAKE")
		r.Error(err)
		r.Equal("none", Get("FLAVOUR", "FAILED"))
		r.Equal("root", Get("DIR", "FAILED"))
	})
}

func Test_KeepEnvWhenSecondLoadFails(t *testing.T) {
	r := require.New(t)
	Temp(func() {
		err := Load(".env")
		r.NoError(err)
		r.Equal("none", Get("FLAVOUR", "FAILED"))
		r.Equal("root", Get("DIR", "FAILED"))

		err = Load(".env.FAKE")

		r.Equal("none", Get("FLAVOUR", "FAILED"))
		r.Equal("root", Get("DIR", "FAILED"))
	})
}

func Test_StopLoadingWhenFileInListFails(t *testing.T) {
	r := require.New(t)
	Temp(func() {
		err := Load(".env", ".env.FAKE", "test_env/.env.prod")
		r.Error(err)
		r.Equal("none", Get("FLAVOUR", "FAILED"))
		r.Equal("root", Get("DIR", "FAILED"))
	})
}

func Test_GOPATH_Not_Set(t *testing.T) {
	r := require.New(t)

	Temp(func() {
		MustSet("GOPATH", "/go")
		loadEnv()
		r.Equal("/go", Get("GOPATH", "notset"))
	})

	r.Equal("github.com/gobuffalo/envy", CurrentPackage())
}

func Test_CurrentModule(t *testing.T) {
	r := require.New(t)

	Temp(func() {
		MustSet("GO111MODULE", "on")
		loadEnv()

		// Make a new dir for test
		testDirPath := "submodule"
		err := os.Mkdir(testDirPath, os.ModeDir|os.ModePerm)
		r.Nil(err)

		// Make sure the dir will delete after test
		// And change back to the old directory
		pwd, err := os.Getwd()
		r.Nil(err)
		defer func() {
			os.RemoveAll(filepath.Join(pwd, testDirPath))
			os.Chdir(pwd)
		}()

		// Change dir to a subdirectory
		err = os.Chdir(testDirPath)
		r.Nil(err)

		if gomod, err := CurrentModule(); err == nil {
			r.Equal("github.com/gobuffalo/envy", gomod)
		} else {
			r.Contains([]string{
				"run `go env GOMOD` command return error while go module is enabled",
				"run `go env GOMOD` command return empty while go module is enabled",
			}, err.Error())
		}
	})
}

func Test_CurrentModule_NoGoModFile(t *testing.T) {
	r := require.New(t)

	Temp(func() {
		MustSet("GO111MODULE", "on")
		loadEnv()

		// Make sure the dir will delete after test
		// And change back to the old directory
		pwd, err := os.Getwd()
		r.Nil(err)
		defer func() {
			os.Chdir(pwd)
		}()

		// Change dir to a location where no go.mod defined
		err = os.Chdir(os.TempDir())
		r.Nil(err)

		_, err = CurrentModule()
		r.NotNil(err)
		// go version less than 1.11 not support the command; go above 1.11 will report the go.mod is malformed (output the /dev/null)
		r.Contains([]string{
			"run `go env GOMOD` command return error while go module is enabled",
			"run `go env GOMOD` command return empty while go module is enabled",
			"go.mod is malformed",
		}, err.Error())
	})
}
