package main

import (
	"./yaml";
	"os";
)

func main() {
	fmt.Printf("%s\n", Tokenize(os.Open("sample.yml", os.O_RDONLY, 0666)));
}	