package envy

import (
	"io/ioutil"
	"log"
	"os"
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

// envy should detect when running as a unit test and return GO_ENV=test if otherwise undefined
// func Test_GO_ENVUnitTest(t *testing.T) {
// 	r := require.New(t)
// 	r.Zero(os.Getenv("GO_ENV"))
// 	r.Equal("test", Get("GO_ENV", "foo"))
// }

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
	r.Zero(os.Getenv("FOO"))
	_, err := MustGet("FOO")
	r.Error(err)

	Set("FOO", "foo")
	r.Equal("foo", Get("FOO", "bar"))
	// but Set should not touch the os envrionment
	r.NotEqual(os.Getenv("FOO"), "foo")
	r.Error(err)
}

func Test_MustSet(t *testing.T) {
	r := require.New(t)
	r.Zero(os.Getenv("FOO"))

	err := MustSet("FOO", "BAR")
	r.NoError(err)
	// MustSet also set underlying os environment
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

func Test_CurrentModule(t *testing.T) {
	r := require.New(t)

	mod, err := CurrentModule()
	r.NoError(err)
	r.Equal("github.com/gobuffalo/envy/v2", mod)
}

// Env files loading: automatically loaded by init()
func Test_LoadEnvLoadsEnvFile(t *testing.T) {
	r := require.New(t)
	r.Equal("root", Get("ENVY_DIR", ""))
	r.Equal("none", Get("ENVY_FLAVOUR", ""))
}

func Test_LoadDefaultEnvWhenNoArgsPassed(t *testing.T) {
	r := require.New(t)

	Temp(func() {
		// using Load() within Temp() is not safe. use it with caution!
		err := Load()
		r.NoError(err)

		r.Equal("root", Get("ENVY_DIR", ""))
		r.Equal("none", Get("ENVY_FLAVOUR", ""))
	})
}

func Test_LoadOnlyAdditionalVariables(t *testing.T) {
	r := require.New(t)

	Temp(func() {
		// using Load() within Temp() is not safe. use it with caution!
		err := Load("test_env/.env")
		r.NoError(err)

		r.Equal("root", Get("ENVY_DIR", ""))                    // from root
		r.Equal("none", Get("ENVY_FLAVOUR", ""))                // from root
		r.Equal("test_env/.env", Get("ENVY_TEST_ENV_PATH", "")) // additional
	})
}

func Test_PrioritizeFirstArg1(t *testing.T) {
	r := require.New(t)
	Temp(func() {
		// using Load() within Temp() is not safe. use it with caution!
		err := Load("test_env/.env.test", "test_env/.env.prod")
		r.NoError(err)

		r.Equal("none", Get("ENVY_FLAVOUR", ""))  // from root
		r.Equal("test", Get("ENVY_TEST_ENV", "")) // from .env.test
	})
}

func Test_PrioritizeFirstArg2(t *testing.T) {
	r := require.New(t)
	os.Unsetenv("ENVY_FLAVOUR")
	os.Unsetenv("ENVY_TEST_ENV")

	Temp(func() {
		// using Load() within Temp() is not safe. use it with caution!
		err := Load("test_env/.env.prod", "test_env/.env.test")
		r.NoError(err)

		r.Equal("production", Get("ENVY_FLAVOUR", "")) // from .env.prod
		r.Equal("prod", Get("ENVY_TEST_ENV", ""))      // from .env.prod
	})
}

func Test_ErrorWhenSingleFileLoadDoesNotExist(t *testing.T) {
	r := require.New(t)
	os.Unsetenv("ENVY_FLAVOUR")

	Temp(func() {
		delete(env, "ENVY_FLAVOUR")
		// using Load() within Temp() is not safe. use it with caution!
		err := Load(".env.fake")

		r.Error(err)
		r.Equal("FAILED", Get("ENVY_FLAVOUR", "FAILED"))
	})
}

func Test_KeepEnvWhenFileInListFails(t *testing.T) {
	r := require.New(t)
	os.Unsetenv("ENVY_FLAVOUR")
	os.Unsetenv("ENVY_DIR")

	Temp(func() {
		// using Load() within Temp() is not safe. use it with caution!
		err := Load(".env", ".env.FAKE")
		r.Error(err)
		r.Equal("none", Get("ENVY_FLAVOUR", "FAILED"))
		r.Equal("root", Get("ENVY_DIR", "FAILED"))
	})
}

func Test_KeepEnvWhenSecondLoadFails(t *testing.T) {
	r := require.New(t)
	os.Unsetenv("ENVY_FLAVOUR")
	os.Unsetenv("ENVY_DIR")

	Temp(func() {
		// using Load() within Temp() is not safe. use it with caution!
		err := Load(".env")
		r.NoError(err)
		r.Equal("none", Get("ENVY_FLAVOUR", "FAILED"))
		r.Equal("root", Get("ENVY_DIR", "FAILED"))

		err = Load(".env.FAKE")
		r.Error(err)

		r.Equal("none", Get("ENVY_FLAVOUR", "FAILED"))
		r.Equal("root", Get("ENVY_DIR", "FAILED"))
	})
}

func Test_StopLoadingWhenFileInListFails(t *testing.T) {
	r := require.New(t)
	os.Unsetenv("ENVY_FLAVOUR")
	os.Unsetenv("ENVY_DIR")

	Temp(func() {
		delete(env, "ENVY_FLAVOUR")
		// using Load() within Temp() is not safe. use it with caution!
		err := Load(".env.FAKE", "test_env/.env.prod")
		r.Error(err)
		r.Equal("FAILED", Get("ENVY_FLAVOUR", "FAILED"))
		r.Equal("root", Get("ENVY_DIR", "FAILED"))
	})
}
