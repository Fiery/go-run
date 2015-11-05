package run

import (
	"log"
	"sync"
	"strings"
	"fmt"
	"io/ioutil"
	"unicode"
	"strconv"

	"github.com/howeyc/gopass"
)



var logger = log.New(ioutil.Discard, "[togo-run] ", log.LstdFlags)

var waitGroup sync.WaitGroup
var workingDir string


func getQuoteSplitter(sep rune) func(rune) bool{
	lastQuote := rune(0)
	return func(c rune) bool {
		switch {
		// quote end
		case c == lastQuote:
			lastQuote = rune(0)
			return false
		// quoted text
		case lastQuote != rune(0):
			return false
		// quote start
		case unicode.In(c, unicode.Quotation_Mark):
			lastQuote = c
			return false
		// unquoted text, need split
		default:
			if sep!=rune(0){ 
				return c==sep
			}else{
				// use space as default separator
				return unicode.IsSpace(c)
			}

		}
	}
}

func parseCommand(command string) (exe string, arg, env []string) {


	for _, a:= range strings.FieldsFunc(command, getQuoteSplitter(rune(0))){
		if unq,err:= strconv.Unquote(a);err!=nil{
			arg = append(arg, a)
		}else{
			arg = append(arg, unq)
		}
	}

	for i, item := range arg {
		if strings.Contains(item, "=") {
			if env == nil {
				env = []string{item}
				continue
			}
			env = append(env, item)
		} else {
			// end env prefix, return tralling command
			exe = item
			arg = arg[i+1:]
			return
		}
	}

	exe = arg[0]
	arg = arg[1:]


	return
}


// Prompt prompts user for input with plain value.
func Prompt(prompt string) string {
    fmt.Print(prompt)
    var input string
    fmt.Scanln(&input)
    fmt.Println(input)
    return input
}

// PromptHidden prompts user for hidden terminal input.
func PromptHidden(prompt string) string {
	fmt.Print(prompt)
	return string(gopass.GetPasswd())
}


// PromptMasked prompts user for masked terminal input.
func PromptMasked(prompt string) string {
	fmt.Print(prompt)
	return string(gopass.GetPasswdMasked())
}