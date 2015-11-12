package run 

import (
	"os"
	"strings"
	"fmt"
	flag "github.com/Fiery/pflag"
)

func init(){
	//flag.Parse()
 	//Env = makeEnvMap(flag.Args(), false)
}


// Uses map to easily update/change/fetch variables
type envMap map[string]string

// Env is the default environment to use for all commands.
// os.exec.Cmd takes environment from the merged set
// of global environment and this Env set. 
var Env = makeEnvMap(flag.Args(), false)


// PathListSeparator is a cross-platform path list separator template. 
// Will be replaced according to different go runtim (Windows: ";", Unix/BSD: ":")
var PathListSeparator = "::"

// Set takes arbitrary number of env assigment and update the mapping in-place
func (e *envMap) Set(set ...string){
	for key, val:= range *makeEnvMap(set, false, e){
		(*e)[key] = val
	}
}



// list returns `os` friendly env assignment list
func (e *envMap) list() (r []string){

	for key, val:= range *e{
		if strings.Contains(val, PathListSeparator) {
			val = strings.Replace(val, PathListSeparator, string(os.PathListSeparator), -1)
		}
		r = append(r, strings.Join([]string{key, val},"="))
	}
	return

}

func (e *envMap) promote(key string) error{
	if val, ok:=(*e)[key]; !ok{
		return fmt.Errorf("Cannot escalate env that are not included in current env map!")
	}else{
		return os.Setenv(key,val)
	}

}

// consolidate consolidates input environment variables in given set and returns a new envMap 
func (e *envMap) combine(envset []string) *envMap {
	final := *makeEnvMap(envset, false, e)
	for key,val:= range *e{
		if _, ok:=final[key];!ok{
			final[key] = val
		}
	}
	return &final
}


// makeEnvMap parse the given env string set and returns a new envMap
func makeEnvMap(set []string, inherit bool, ref ...*envMap) *envMap {
	env := make(envMap)
	for _, kv := range set {
		if strings.Contains(kv, PathListSeparator){
			kv = strings.Replace(kv, PathListSeparator, string(os.PathListSeparator), -1)
		}

		kv = os.Expand(kv, func(key string) string{
			for _, r := range append(ref, &env){
				if v,ok:=(*r)[key];ok{
					return v
				}
			}
			if v:=os.Getenv(key); inherit && v!=""{ 
				return v
			}else{
				return ""//key 
			}

		})

		if pair:= strings.Split(kv,"=");len(pair)<2{
			env[pair[0]] = ""
		}else if len(pair)>2{
			pair = strings.FieldsFunc(kv, getQuoteSplitter('='))
			env[pair[0]] = pair[1]		
		}else{			
			env[pair[0]] = pair[1]
		
		}

	}
	return &env
}

