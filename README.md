## go-run

Running system program with calls chained together, inspired by gulp.js.

Inspired by [godo](https://godoc.org/github.com/mgutz/godo/v2) and created as a plugin for [belt](https://github.com/Fiery/belt) task runner

[![GoDoc](https://godoc.org/github.com/Fiery/go-run?status.svg)](https://godoc.org/github.com/Fiery/go-run)

### Quick Start

As a quick overview of usages, below shows an example of how to use `run` in general.

*go-run runs independently, means that you can import it and use it anywhere you want*

```go
package main

import (
    "fmt"
    "github.com/Fiery/go-run"
)

func main(){
    // set local env, use "::" as cross-platform path list separator
    run.Set(`GOPATH=./lib/::$GOPATH`)
    // simply executes shell command with local and global envs
    run.Shell("echo Hello $USER!").Run()


    // You could also specify local envs in command line strings, just as how you type in shell
    run.Call("GOOS=linux GOARCH=amd64 go build").Run()

    // In allows you to temporarily change to the specified folder and run
    run.Shell("bash","script.sh").In("project").Run()

    // Pipe the Stdout (and/or Stderr) to evaluate later
    var output = bytes.Buffer
    run.Call("echo Just run it!").Pipe(&output, run.Stdout|run.Stderr).Run()
    fmt.Println(output.String())


    // Asynchronously runs the command
    run.Start("node server.js").Run()
}

```
### API
#### run.Runnable

Runnable is the universal interface for all chainable runners available in `run`.
Fill in your own type with below functions to implement run.Runnable.
```
  Run() error

  // Runnable functions can be chained in any order

  Call(string) Runnable
  Start(string) Runnable
  Shell(...string) Runnable

  With(...string) Runnable
  Pipe(int, *bytes.Buffer) Runnable

  At(string) Runnable
  In(string) Runnable
```


#### run.Shell

Shell returns an object implements `Runnable`, which allows you to chain another `run` function.
It uses `sh` as default shell and can be specified with optional configuration.

```go
run.Shell(`
    echo -n $USER
    echo some really long \
        command
`).Run()
```
above snippet will run multiline commands.

Environment variables are acceptable anywhere in command line
If the binary itself is a env var, it will be expanded before execution

```go
run.Env.Set("name=run")
run.Shell(`echo -n $name`).Run()
```

Can be chained by Pipe( ) to capture Stdout and Stderr or to bind Stdin.

```go
var output bytes.Buffer
run.Shell(`bash`, `echo -n $USER`).Pipe(&output, run.Stdout).Run()
```


#### run.Call

Similarly, `run.Call()` returns `Runnable` which can be chained by Pipe( ), With( ), In( ), or another Call( ) if needed

```go
  run.Call("echo Hello").Call("echo Hello again").Call("echo Hello again and again").Run()
  // Output:
  // Hello
  // Hello again
  // Hello again and again

```



#### run.Start

Start an async command. returns immediately.

```go
run.Start("main").Run()
```

#### run.Runnable.In

Temporarily `cd` to specified path and run the Runnables

```go
run.Call("...").In("path/to/run").Run()
run.Shell("...").In("path/to/run").Run()
```

#### run.Runnable.At

Runs the command with working directory set to be the input

```go
run.Call("...").At("path/to/run").Run()
run.Shell("...").At("path/to/run").Run()
```


#### run.Runnable.Pipe

Any Runnable can use Pipe to direct its Stdin|Stdout|Stderr to a predefined io.Reader|io.Writer. Again it returns Runnable which can be chained with other functions.

```go
var output = bytes.NewBuffer(nil)
run.Call(`GOOS=linux GOARCH=amd64 go build`).Pipe(output, Stdout|Stderr)
```


#### run.Runnable.With

Set command specific variables, only valid within the calling Runnable chain

```go
run.Call("$c $t/txt").With("c=cat","t=test").Run()
```


### Runtime Environment

#### Set from command line 

Environment variables may be set via key-value assignments as prefix in command line.

Process environment will be set as the combination of available system environment variables and `run.Env`


```sh
  USER=run PATH=$PATH::./lib/ ./run/main
```

#### Set using run.Env

Use `run.Env.Set` To specify runtime environment for the command.
Assignment expression to be separated with any type of spaces.
All `$VAR` or `${VAR}` will be expanded when loading at runtime

```go
run.Env.Set(
    // Path list should use "::" as cross-platform path list separator.
    // On Windows "::" will be replaced with ";".
    // On Mac and linux "::" will be replaced with ":".
`
    PATH=./lib/::$PATH
    USER=user
`)
```

#### Use run.Runnable.With

More fine-grained environment configuration can be achieved by using With( )

```go
// only valid in current Call
run.Call("$c $t/txt").With("c=cat","t=test").Run()
```


TIP: Set the `Env` when using a dependency manager like `godep`

```go
wd, _ := os.Getwd()
run.Env.Set(fmt.Sprintf("GOPATH=%s::$GOPATH", path.Join(wd, "Godeps/_workspace")))
```




### Tools

* To get plain string user input

  ```go
user := Prompt("user: ")
```

* To get hidden/masked user input

  ```go
password := PromptHidden("password: ")
password := PromptMasked("password: ")
```
