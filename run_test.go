package run 

import (
	"runtime"
	"testing"
	"bytes"
	"strings"

	"os"
	"github.com/Fiery/testify/assert"

)

var cat = "cat"
func init() {
	logger.SetOutput(os.Stdout)
	if runtime.GOOS == "windows" {
		cat = "type"
	}
}


func TestSingleCommand(t *testing.T) {

    var output bytes.Buffer
	Call(`cat test/foo`).Pipe(Stdout, &output).Run()
	assert.Equal(t, "text from foo\n", output.String())
}

func TestChain(t *testing.T){
    var output bytes.Buffer
	Call(`cat test/foo`).Call(`echo hello`).Call(`echo world`).Pipe(Stdout, &output).Run()
	assert.Equal(t, "text from foo\nhello\nworld\n", output.String())
}



func TestMultiCommand(t *testing.T) {

    var output bytes.Buffer
	Call(cat+" test/foo\n"+cat+" test/bar").Pipe(Stdout, &output).Run()
	assert.Equal(t, "text from foo\nthis is content of bar\n", output.String())
}



func TestAt(t *testing.T) {
    var output bytes.Buffer
	if runtime.GOOS == "windows"{
		Call("foo.cmd").At("test").Pipe(Stdout, &output).Run()
	} else {
		Call("bash foo.sh").At("test").Pipe(Stdout, &output).Run()
	}

	assert.Equal(t, "FOOBAR", strings.Trim(output.String()," "),)
}

func TestWithError(t *testing.T) {
    var output bytes.Buffer
    var error bytes.Buffer
	status:=Call(`
		$cat $t/doesnotexist
		$cat $t/bar
		`).With("t=test","cat=cat").Pipe(Stdout, &output).Pipe(Stderr, &error).Run()

	assert.Error(t, status)
	assert.Contains(t, "line=1", status.Error())
	assert.Contains(t, "exit", status.Error())
	assert.Contains(t, "doesnotexist", error.String())
	assert.Equal(t, "", output.String())
}


func TestIn(t *testing.T) {
    var output bytes.Buffer
    old, _ := os.Getwd()
	if runtime.GOOS == "windows"{
		Call("foo.cmd").In("test").Pipe(Stdout, &output).Run()
	} else {
		Call("bash foo.sh").In("test").Pipe(Stdout, &output).Run()
	}

	assert.Equal(t, "FOOBAR", strings.Trim(output.String()," "),)
	
	now,_ := os.Getwd()
	assert.Equal(t, now, old, "In failed to reset work directory")
}

func TestShell(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}
    var output bytes.Buffer
    // use sh as default shell, which doesn't suport `-n`
	Shell(`echo foobar`).Pipe(Stdout, &output).Run()
	assert.Equal(t,"foobar\n" , output.String(), "Simple bash failed.")

	output.Reset()
	Shell(`bash`, `echo -n foobar`).Pipe(Stdout, &output).Run()
	assert.Equal(t,"foobar" , output.String(), "Simple bash failed.")

	output.Reset()
	Shell(`bash`,`
		echo -n foobar
		echo -n bahbaz
	`).Pipe(Stdout, &output).Run()
	assert.Equal(t,"foobarbahbaz" , output.String(), "Multiline bash failed.")

	output.Reset()
	Shell(`bash`, `
		echo -n \
		foobar
	`).Pipe(Stdout, &output).Run()
	assert.Equal(t,"foobar" , output.String(), "Bash line continuation failed.")

	output.Reset()
	Shell(`bash`,`
		echo -n "foobar"
	`).Pipe(Stdout, &output).Run()
	assert.Equal(t,"foobar" , output.String(), "Bash quotes failed.")

	output.Reset()
	Shell(`bash`,`
		echo -n "fo\"obar"
	`).Pipe(Stdout, &output).Run()
	assert.Equal(t,"fo\"obar" , output.String(), "Bash quoted command failed.")
}

func TestStart(t *testing.T){
	
}


func TestEnvOperation(t *testing.T) {
	// Equivalent to os.Setenv("TEST_RUN_ENV", "fubar")
	/*if runtime.GOOS == "windows" {
		status:=Call(`set TEST_RUN_ENV=fubar`).Run()
		assert.Equal(t, "",status.Error())
	} else {
		status:=Call(`bash -c "export TEST_RUN_ENV=fubar"`).Run()
		assert.Equal(t, nil,status)
	}*/
	os.Setenv("TEST_RUN_ENV", "fubar")
	Env = makeEnvMap(os.Environ(), true)

	assert.Equal(t, "fubar", (*Env)["TEST_RUN_ENV"], "set/export env failed")	
	output:= bytes.NewBuffer(nil)
	if runtime.GOOS == "windows" {
		Call(`FOO=bar BAH=baz cmd /C 'echo %TEST_RUN_ENV% %FOO%'`).Pipe(Stdout, output).Run()
	} else {
		Call(`FOO=bar BAH=baz bash -c "echo -n $TEST_RUN_ENV $FOO"`).Pipe(Stdout, output).Run()
	}

	assert.Equal(t, "fubar bar",strings.Trim((*output).String()," "))

	os.Unsetenv("TEST_RUN_ENV")
	/*
	if runtime.GOOS == "windows" {
		Call(`set TEST_RUN_ENV=`).Run()
	} else {
		Call(`unset TEST_RUN_ENV`).Run()
	}
	*/
}