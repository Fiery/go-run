package run 

import (
	"fmt"
	"os"
	"os/exec"
	"sync"

	"time"
)

// appMap stores all exec.Cmd invoked by user, 
var appMap = make(map[string]map[time.Time]*exec.Cmd)


type app struct{
	// extracted arguments
	arg []string
	// extracted binary 
	bin string
	// complete command line
	cmd string
	// working directory
	dir string
	// extracted command specific env
	env []string

}

type syncApp struct {
	app
}

type asyncApp struct{
	app
	wg *sync.WaitGroup
}

// getCmd returns exec.Cmd
// binary names will be evaluated with Env here since this is the last step before Run()
func (a *app) getCmd() (*exec.Cmd, error) {
	path, err := exec.LookPath(a.bin)
	if err != nil {
		if path , err= exec.LookPath(os.Expand(a.bin, func(key string)string{
			if v,ok:=(*Env)[key];ok{
					return v
			}else{
				return ""
			}
		})); err!=nil{
			return nil, fmt.Errorf("installing %v is in your future...", a.bin)
		}
	}
	cmd := exec.Command(path, a.arg...)
	if a.dir != "" {
		cmd.Dir = a.dir
	}

	cmd.Env = Env.combine(append(a.env, os.Environ()...)).list()
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr 


	if len(Env.list())>0  {
		logger.Printf("Env: %s\n", Env)
	}
	logger.Printf("Command loaded: %s\n", a.cmd)


	return cmd, nil
}


func (sa *syncApp) Run() error{
	if cmd, err:= sa.getCmd(); err!=nil{
		return err
	}else{
		return cmd.Run()
	}
}


func (aa *asyncApp) Run() (err error) {
	aa.wg.Add(1)
	defer aa.wg.Done()
	if cmd,err:= aa.getCmd(); err != nil {
		return err
	}else if err:= cmd.Start();err != nil {
		return err
	}else{
		if c, ok:= appMap[aa.cmd]; !ok{
			c= make(map[time.Time]*exec.Cmd)
			appMap[aa.cmd] = c
			c[time.Now()] = cmd
		}else{
			c[time.Now()] = cmd
		}

		logger.Println("Async application added [%q]\n", aa.cmd)
		//err = cmd.Wait()
		return nil
	}
}

// kills processes with the corresponding command, 
func Stop(cmd string, spans ...[2]time.Time) error{
	if cm,ok:= appMap[cmd]; ok {
	timecheck:
		for ts, c:=range cm{
			for _,span:=range spans {
				if ts.Before(span[0]) || ts.After(span[1]) {
					continue timecheck
				}
			}
			if p, err:= os.FindProcess(c.ProcessState.Pid());err!=nil{
				return fmt.Errorf("Specified process not found in current available processes!")
			}else{
				if !c.ProcessState.Exited(){
					if err = c.Process.Kill(); err!=nil {
						return fmt.Errorf("Could not kill process %+v\n%v\n", p, err)
					}
				}
			}
			delete(cm, ts)
			logger.Printf("Processes[%q:%v] killed\n", cmd, ts.Format(time.RFC850))
		}
		
	}else{
		return fmt.Errorf("Specified process not in maintanance list!\n")

	}
	return nil
}
