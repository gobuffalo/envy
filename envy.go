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

var Default = New()

// GO111MODULE is ENV for turning mods on/off
const GO111MODULE = "GO111MODULE"

var Mods = Default.Mods

// Reload the ENV variables. Useful if
// an external ENV manager has been used
var Reload = Default.Reload

// Load .env files. Files will be loaded in the same order that are received.
// Redefined vars will override previously existing values.
// IE: envy.Load(".env", "test_env/.env") will result in DIR=test_env
// If no arg passed, it will try to load a .env file.
var Load = Default.Load

// Get a value from the ENV. If it doesn't exist the
// default value will be returned.
var Get = Default.Get

// Get a value from the ENV. If it doesn't exist
// an error will be returned
var MustGet = Default.MustGet

// Set a value into the ENV. This is NOT permanent. It will
// only affect values accessed through envy.
var Set = Default.Set

// MustSet the value into the underlying ENV, as well as envy.
// This may return an error if there is a problem setting the
// underlying ENV value.
var MustSet = Default.MustSet

// Map all of the keys/values set in envy.
var Map = Default.Map

// Temp makes a copy of the values and allows operation on
// those values temporarily during the run of the function.
// At the end of the function run the copy is discarded and
// the original values are replaced. This is useful for testing.
// Warning: This function is NOT safe to use from a goroutine or
// from code which may access any Get or Set function from a goroutine
var Temp = Default.Temp

var GoPath = Default.GoPath

var GoBin = Default.GoBin

var InGoPath = Default.InGoPath

// GoPaths returns all possible GOPATHS that are set.
var GoPaths = Default.GoPaths

// CurrentModule will attempt to return the module name from `go.mod` if
// modules are enabled.
// If modules are not enabled it will fallback to using CurrentPackage instead.
var CurrentModule = Default.CurrentModule

// CurrentPackage attempts to figure out the current package name from the PWD
// Use CurrentModule for a more accurate package name.

var CurrentPackage = Default.CurrentPackage

var Environ = Default.Environ
