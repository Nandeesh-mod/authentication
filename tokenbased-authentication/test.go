package main

import "fmt"

func Formula(a int, b int, c func(a int, b int) (int, error)) int {
	res := a + b
	val, err := c(a, b)
	if err != nil {
		fmt.Println("Error")
	}
	res += val
	return res
}

func main() {
	res := Formula(10, 20, func(a int, b int) (int, error) {
		return (a + b), nil
	})

	fmt.Println(res)
}
