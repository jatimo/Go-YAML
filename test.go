package main

import (
	"./yaml";
	"os";
	"fmt";
)

func main() {
	file, err := os.Open("sample.yml", os.O_RDONLY, 0666);
	if err != nil {
		//panic
	}
	fmt.Printf("%s\n ", yaml.Tokenize(file));
}	