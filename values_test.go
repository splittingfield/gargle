package gargle

import "fmt"

func ExampleNegatedBoolVar() {
	var value bool
	NegatedBoolVar(&value).Set("false")
	fmt.Println(value)

	// Output: true
}
