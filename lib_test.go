package run 

import (
	"testing"
	"strings"
	"fmt"

	"os"
	"github.com/Fiery/testify/assert"

)

var r, w *os.File
var err error
func init(){
	r,w,err = os.Pipe()
	if err!=nil{
		fmt.Println(err)
		os.Exit(1)
	}
}

func TestPrompt(t *testing.T) {
	stdin:=os.Stdin
	defer func(){
		os.Stdin = stdin
	}()
	os.Stdin = r

	w.WriteString("username\n")

	assert.Equal(t, "username", strings.Trim(Prompt("Please input: "), " "))

}