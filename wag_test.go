package wag

import (
	"io/ioutil"
	"testing"
)

func TestLowerIfElse(t *testing.T) {
	test(t, "testsuite/lower-if-else.wasm")
}

func TestHelloWorld(t *testing.T) {
	test(t, "testsuite/hello_world.wasm")
}

func test(t *testing.T, filename string) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	module := loadModule(data)
	t.Logf("module = %v", module)

	function := &module.Functions[0]
	t.Logf("function = %v", function)

	result := function.execute([]int64{1, 2})
	t.Logf("result = %v", result)
}
