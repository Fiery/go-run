package run 

import (
	"testing"
	"strings"
	"fmt"

	"os"
	"github.com/stretchr/testify/assert"

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

	w.WriteString("Sample\n")

	assert.Equal(t, strings.Trim(Prompt("Please input: "), " "), "Sample")

}


/* TODO: @Fiery: Fix below test, now fails on Mac

func TestHiddenPrompt(t *testing.T) {

	stdin:=os.Stdin
	defer func(){
		os.Stdin = stdin
	}()
	os.Stdin = r

	w.WriteString("Sample\n")

	assert.Equal(t, strings.Trim(PromptHidden("Please input: ")," "), "Sample")
}

func TestMaskedPrompt(t *testing.T) {
	stdin:=os.Stdin
	defer func(){
		os.Stdin = stdin
	}()
	os.Stdin = r

	w.WriteString("Sample\n")

	assert.Equal(t, strings.Trim(PromptMasked("Please input: ")," "), "Sample")
}
*/