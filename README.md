# envy

[![Actions Status](https://github.com/gobuffalo/envy/workflows/Tests/badge.svg)](https://github.com/gobuffalo/envy/workflows/actions)


Envy makes working with ENV variables in Go trivial.

* Get ENV variables with default values.
* Set ENV variables safely without affecting the underlying system.
* Temporarily change ENV vars; useful for testing.
* Map all of the key/values in the ENV.
* Loads .env files (by using [godotenv](https://github.com/joho/godotenv/))
* More!

## Installation

```text
$ go get -u github.com/gobuffalo/envy
```

## Usage

```go
func Test_Get(t *testing.T) {
	r := require.New(t)
	r.NotZero(os.Getenv("GOPATH"))
	r.Equal(os.Getenv("GOPATH"), envy.Get("GOPATH", "foo"))
	r.Equal("bar", envy.Get("IDONTEXIST", "bar"))
}

func Test_MustGet(t *testing.T) {
	r := require.New(t)
	r.NotZero(os.Getenv("GOPATH"))
	v, err := envy.MustGet("GOPATH")
	r.NoError(err)
	r.Equal(os.Getenv("GOPATH"), v)

	_, err = envy.MustGet("IDONTEXIST")
	r.Error(err)
}

func Test_Set(t *testing.T) {
	r := require.New(t)
	_, err := envy.MustGet("FOO")
	r.Error(err)

	envy.Set("FOO", "foo")
	r.Equal("foo", envy.Get("FOO", "bar"))
}

func Test_Temp(t *testing.T) {
	r := require.New(t)

	_, err := envy.MustGet("BAR")
	r.Error(err)

	envy.Temp(func() {
		envy.Set("BAR", "foo")
		r.Equal("foo", envy.Get("BAR", "bar"))
		_, err = envy.MustGet("BAR")
		r.NoError(err)
	})

	_, err = envy.MustGet("BAR")
	r.Error(err)
}
```
## .env files support

NOTE: the behavior of `.env` support was changed in `v2`.
Previously, `envy.Load()` overwrote all pre-existing environment variables
with the values in the `.env` file but now pre-existing variables have higher
priority and will remain as is. (It will help you to configure your runtime
environment in the modern platforms including cloud computing environments)

Envy now supports loading `.env` files by using the
[godotenv library](https://github.com/joho/godotenv/).
That means one can use and define multiple `.env` files which will be loaded
on-demand.
By default, Envy loads `.env` file in the working directory if the file exists.
To load additional one or more, you need to call the `envy.Load` function in
one of the following ways:

```go
envy.Load() // 1

envy.Load("MY_ENV_FILE") // 2

envy.Load(".env.prod", ".env") // 3

envy.Load(".env", "NON_EXISTING_FILE") // 4

envy.Load(".env.prod", "NON_EXISTING_FILE", ".env") // 5

// 6
envy.Load(".env.prod")
envy.Load("NON_EXISTING_FILE")
envy.Load(".env")

```

1. Will load the default `.env` file from the current working directory.
2. Will load the file `MY_ENV_FILE`, **but not** `.env`
3. Will load the file `.env.prod` first, then will load the `.env` file.
   If any variable is redefined in `.env`, they will be ignored.
4. Will load the `.env` file and return an error as the second file does not
   exist. The values in `.env` will be loaded and available.
5. Will load the `.env.prod` file and return an error as the second file does
   not exist. The values in `.env.prod` will be loaded and available,
   **but the ones in** `.env` **won't**.
5. The result of this will be the same as 3 as you expected.
