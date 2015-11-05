package run 

import (
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"

)

var user string
func init(){
	if runtime.GOOS == "windows" {
		user = os.Getenv("USERNAME")
		os.Setenv("USER", user)
	} else {
		os.Setenv("USER", "user")
		user = os.Getenv("USER")
	}
}

func TestEnvInitialization(t *testing.T) {
	// set to os defaults
	Env = makeEnvMap(os.Environ(), false)
	Env.Set("USER=$USER:togo")
	assert.Equal(t, user+":togo", (*Env)["USER"], "Environment interpretation failed")
}
func TestEnvCombination(t *testing.T){
	// set to os defaults
	Env = makeEnvMap(os.Environ(), false)
	l := len(os.Environ())
	Env = Env.combine([]string{"USER=$USER:$USER:func"})
	
	assert.Equal(t, user+":"+user+":func", (*Env)["USER"], "Should have been overriden by func environment")
	
	assert.Equal(t, len(*Env), l, "Consolidated environment length changed.")

	Env = Env.combine([]string{"GOSU_NEW_VAR=foo"})
	assert.Equal(t, "foo", (*Env)["GOSU_NEW_VAR"] , "Should have conslidated Env set")
	assert.Equal(t,  l+1, len(*Env), "Consolidated environment length should have increased by 1")

}

func TestEnvQuote(t *testing.T) {

	Env = Env.combine([]string{`FOO="a=bar b=bah c=baz"`})
	if val, ok:= (*Env)["FOO"]; !ok{
		t.Error("Key insertion failed", Env)
	}else if val != `"a=bar b=bah c=baz"` {
		t.Errorf("Quoted var failed %q", val)
	}
}

func TestEnvInterpretation(t *testing.T) {

	// set back to default
	Env = makeEnvMap(os.Environ(), false)
	Env.Set(`USER1=$USER`,`USER2=$USER1`)

	Env = Env.combine([]string{"USER3=$USER2"})
	assert.Equal(t, (*Env)["USER1"],user , "Should have been evaluated")
	assert.Equal(t, (*Env)["USER3"],user , "Should have been evaluated during consolidation.")

	Env = Env.combine([]string{"PATH=foo::bar::bah"})
	assert.Equal(t, "foo"+string(os.PathListSeparator)+"bar"+string(os.PathListSeparator)+"bah", (*Env)["PATH"], "Should have replaced run.PathSeparator")

	// set back to defaults
	Env = makeEnvMap(os.Environ(),false)
	Env.Set(`FOO=foo`,`FAIL=$FOObar:togo`,`OK=${FOO}bar:togo`)

	assert.Equal(t, ":togo", (*Env)["FAIL"], "$FOObar should have been interpreted as empty string.")
	assert.Equal(t, "foobar:togo", (*Env)["OK"], "${FOO}bar should have been interpreted accordingly.")
}

func TestPromotion(t *testing.T) {
	assert.Equal(t, "", os.Getenv("_foo"))
	assert.Equal(t, "", os.Getenv("_test_bar"))
	assert.Equal(t, "", os.Getenv("_test_opts"))
	Env = makeEnvMap([]string{"_foo", "_test_bar=bah", `_test_opts="a=b,c=d,*="`}, false)
	for key,_:= range *Env{
		Env.promote(key)
	}
	assert.Equal(t, "", os.Getenv("_foo"))
	assert.Equal(t, "bah", os.Getenv("_test_bar"))
	assert.Equal(t, "\"a=b,c=d,*=\"", os.Getenv("_test_opts"))
}
