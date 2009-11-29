all: yaml.8 test

yaml.8: tokenize.go
	8g -o yaml.8 tokenize.go
	
test: test.go
	8g test.go
	8l -o test test.8