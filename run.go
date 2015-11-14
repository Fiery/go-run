// Package run is a command runner with deliberately designed chainable interface, inspired by gulp.js
package run

import (
	"os"
	"strings"
	"fmt"
	"bytes"
	"io"
	"log"
	"io/ioutil"
	"syscall"
)

const (
	// Pipe bitmask options
	Stdin = 0x01<<iota
	Stdout
	Stderr
	
	// Shorthand version
	I = Stdin
	O = Stdout
	E = Stderr

	None = 0 
)


// Shell defines a shell command Runnable object, which evaluates sh/bash command line. 
// Use backquote for multiline commands.
// To specify a shell, insert shell name as the first string argument
// To run as shell script, use run.Call("sh/bash script.sh")
func Shell(c ...string) Runnable{
	return shell(c, nil)
}
// Call defines a system call Runnable object, which calls the command with arguments
func Call(c string) Runnable{
	return call(c, nil)
}
// Start starts an async Runnable object, which does basically same as Call, returns immediately
func Start(o string) Runnable{
	return start(o, nil)
}

// Wait waits till all preceded async apps finish
func Wait(){
	daemonProc.Wait()
}

// Runnable expose APIs for chainable call structure   
type Runnable interface{
	Run() error

	// Runnable functions can be chained arbitrarily 

	Call(string) Runnable
	Start(string) Runnable
	Shell(...string) Runnable

	With(...string) Runnable
	Pipe(int, *bytes.Buffer) Runnable

	At(string) Runnable
	In(string) Runnable
}

// runner is Runnable's underlying implementation
type runner func() error

// Run implements Runnable interface
func (r runner) Run() error{
	return r()
}

// Pipe implements Runnable interface
func (r runner) Pipe(p int, b *bytes.Buffer) Runnable{
	return pipe(p,b, r)
}

// In implements Runnable interface
func (r runner) In(p string) Runnable{
	return in(p, r)
}

// With implements Runnable interface
func (r runner) With(o ...string) Runnable{
	return with(o, r)
}
// Shell implements Runnable interface
func (r runner) Shell(c ...string) Runnable{
	return shell(c, r)
}
// Call implements Runnable interface
func (r runner) Call(c string) Runnable{
	return call(c, r)
}
// Start implements Runnable interface
func (r runner) Start(o string) Runnable{
	return start(o, r)
}

// Wd implements Runnable interface
func (r runner) At(p string) Runnable{
	return at(p,r)
}


var logger = log.New(ioutil.Discard, "[run] ", log.LstdFlags)

func with(vars []string, run Runnable) Runnable {
	return runner(func() error {
		// Naaive implementation, need cmdean up
		env := Env
		Env = Env.combine(vars)
		if run !=nil {
			if err :=  run.Run(); err!=nil{
				return err
			}
		}
		Env = env
		return nil
	})
}

func at(path string, run Runnable) Runnable {

	return runner(func() error {
		pwd := workingDir
		defer func(){
			workingDir = pwd
		}()
		workingDir= path
		if run!=nil{
			return run.Run()
		}else{
			return nil
		}
	})



}

func pipe(pipe int, buf *bytes.Buffer, run Runnable) Runnable {
	return runner(func() error {

		if pipe>0{
			if r, w, err := os.Pipe(); err!=nil{
				fmt.Fprintln(os.Stderr, err)
				return err
			}else{

				if pipe&Stderr > 0{
					if stderr:= os.Stderr; stderr.Fd()!= uintptr(syscall.Stderr) {
						return fmt.Errorf("Stderr already piped out!")
					}else{
						os.Stderr = w
						defer func(){
							os.Stderr = stderr
						}()
					}
				}
				if pipe&Stdout > 0{
					if stdout:= os.Stdout; stdout.Fd()!= uintptr(syscall.Stdout){
						return fmt.Errorf("Stdout already piped out!")
					}else{
						os.Stdout = w
						defer func(){
							os.Stdout = stdout
						}()
					}
				}
				
				ec := make(chan error)
				go func() {
					_, err := io.Copy(buf, r)
					r.Close()
					if err != nil {
						//fmt.Fprintf(os.Stderr, "testing: copying pipe: %v\n", err)
						ec <- fmt.Errorf("Error when copying stream %v\n", err)
					}
					ec <- nil
				}()
				
				if run!=nil{
					if err := run.Run();  err != nil {
						return fmt.Errorf("Error when running: %v\n", err)
					}
				}

				
				w.Close()

				return <-ec 
				
			}
		}else{
			return fmt.Errorf("Not valid pipe option! %d", pipe) 
		}
	})

}

func mock_pipe(pipe int, buf *bytes.Buffer, run Runnable) Runnable {
	return runner(func() error {
		r, w,_ := os.Pipe()
		stdout:= os.Stdout
		
		os.Stdout = w
		defer func(){
			os.Stdout = stdout
		}()


		if run!=nil{
			if err:=run.Run();err!=nil{
				return err
			}
		}

		
		w.Close()

		_, err := io.Copy(buf, r)
		r.Close()
		if err != nil {
			return err 
		}

		return nil
	

	})

}


func in(path string, run Runnable) Runnable {
	return runner(func() error{
		if dir, err := os.Getwd(); err != nil {
			return err
		}else{	
			defer func() {
				os.Chdir(dir)
			}()
		}

		
		if err := os.Chdir(path); err != nil {
			return err
		}

		if run!=nil{
			return run.Run()
		}else{
			return nil
		}
	})
}


func start(command string, run Runnable) Runnable {
	return runner(func() error{

		if run!=nil{
			if err:=run.Run(); err!=nil{
				return err
			}
		}

		for _, line := range strings.Split(command, "\n"){
			line = strings.Trim(line, " \t")
			if line == "" {
				continue
			}
			bin, arg, env := parseCommand(line)

			return (&asyncApp{
				app{
					bin: bin,
					dir: workingDir,
					arg: arg,
					cmd: line,
					env: env,
				},
				nil,
				&daemonProc,
			}).Run()
		}
		return nil

	})
}

func shell(command []string, run Runnable) Runnable {

	return runner(func() error{
		if run!=nil{
			if err:=run.Run(); err!=nil{
				return err
			}
		}
		if len(command)==2{
			return (&syncApp{
				app{
					bin: command[0],
					arg: []string{"-c", command[1]},
					dir: workingDir ,
					cmd: strings.Join(command, " "),
				},
			}).Run()
		}else if len(command)==1{
			return (&syncApp{
				app{
					// default shell
					bin: "sh",
					arg: []string{"-c", command[0]},
					dir: workingDir ,
					cmd: command[0],
				},
			}).Run()
		}else{
			return fmt.Errorf("Shell command options not valid!")
		}

	})

}

func call(command string, run Runnable) Runnable {

	return runner(func() error{
		if run!=nil{
			if err:=run.Run(); err!=nil{
				return err
			}
		}

		for i, line := range strings.Split(command, "\n"){
			line = strings.Trim(line, " \t")
			if line == "" {
				continue
			}
			bin, arg, env := parseCommand(line)

			var prog *syncApp
			prog = &syncApp{
				app{
					bin: bin,
					dir: workingDir,
					arg: arg,
					cmd: line,
					env: env,
				},
			}
			
			err := prog.Run()
			if err != nil {
				return fmt.Errorf(err.Error()+"\nline=%d", i)
			}
		}
		return nil 
	})
}
