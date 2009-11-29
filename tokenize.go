package yaml

import (
	"io";
	"strings";
	"regexp";
	"fmt";
);

//TOKENIZE.GO

/*
Input character stream

invoice: 34843
date   : 2001-01-23
bill-to: &id001
    given  : Chris
    family : Dumars
    address:
        lines: |
            458 Walkman Dr.
            Suite #292
        city    : Royal Oak
        state   : MI
        postal  : 48046
ship-to: *id001
product:
    - sku         : BL394D
      quantity    : 4
      description : Basketball
      price       : 450.00
    - sku         : BL4438H
      quantity    : 1
      description : Super Hoop
      price       : 2392.00
tax  : 251.42
total: 4443.52
comments: >
    Late afternoon is best.
    Backup contact is Nancy
    Billsmer @ 338-4338.

Tokenized

STRING invoice
MAPVALUE
INT 34843
STRING date
MAPVALUE
DATE 2001-01-23
STRING bill-to
MAPVALUE
ANCHOR id001
 STRING given
 MAPVALUE
 Chris
 STRING family
 MAPVALUE
 Dumars
 STRING address
 MAPVALUE
  STRING lines
  MAPVALUE
  LITERAL
   458 Walkman Dr.
   Suite #292
  STRING city
  MAPVALUE
  STRING Royal Oak
  STRING state
  MAPVALUE
  String MI
  String postal
  MAPVALUE
  INT 48046
STRING ship-to
MAPVALUE
ALIAS id001
STRING product
MAPVALUE
 ENTRY
  STRING sku
  MAPVALUE
  STRING BL394D
  STRING quantity
  MAPVALUE
  INT 4
  STRING description
  MAPVALUE
  String Basketball
  STRING price
  MAPVALUE
  FLOAT 450.00
 ENTRY
  STRING sku
  MAPVALUE
  STRING BL4438H
  STRING quantity
  MAPVALUE
  INT 1
  STRING description
  MAPVALUE 
  STRING Super Hoop
  STRING price
  MAPVALUE
  FLOAT 2392.00
STRING tax
MAPVALUE
FLOAT 251.42
STRING total
MAPVALUE
FLOAT 4443.52
STRING comments
MAPVALUE
 FOLDED
  Late afternoon is best.
  Backup contact is Nancy
  Billsmer @ 338-4338.
  
*/

/*
First iteration schema support:
bool
int (base 10)
float
string
(just what is needed to tokenize the above example)
*/

type LexicalMatch struct {
	exp string; //regular expression
	out string; //formatted string, where %s is the lexeme that replaces the lexeme
}

matches := []LexicalMatch { //In order of precedence!
	LexicalMatch { ">", "FOLDED" },
	LexicalMatch { "|", "LITERAL" },
	LexicalMatch { "&", "ANCHOR" },
	LexicalMatch { "*", "ALIAS" },
	LexicalMatch { "- ", "ENTRY" },
	LexicalMatch { ": ", "MAPVALUE"}, //Map value
	LexicalMatch { "[\\-\\+]?(\\.[0-9]+|[0-9]+(\\.[0-9]*)?)([eE][\\-\\+]?[0-9]+)?", "FLOAT %s" }, //FLOAT
	LexicalMatch { "[\\-\\+]?[0-9]+", "INT %s" }, //INT (base 10)
	LexicalMatch { "[^:^&^*^ ^\\|]+( *[^:^&^*^ ^\\|])*", "STRING %s" }, //STRING
}

func doesMatch(s string) bool {
	for _, match := range matches {
		rx := regexp.MustCompile(match.exp);
		res := rx.ExecuteString(s);
		if len(res) == 0 {
			continue
		}
		for i := 0; i < len(res)/2; i += 2 {
			if res[i] == 0 && res[i + 1] {
				return true
			}
		}
	}
}

func tokenizeLexeme(s string) string {
	for _, match := range matches {
		rx := regexp.MustCompile(match.exp);
		res := rx.ExecuteString(s);
		if len(res) == 0 {
			continue
		}
		for i := 0; i < len(res)/2; i += 2 {
			if res[i] == 0 && res[i + 1] {
				return fmt.Sprintf(match.out, s);
			}
		}
	}
}

type Scanner struct {
	data string;
	index int;
}

//Step returns the next character in the character stream. Returns "" if no characters left
func (scan *Scanner)Step() string {
	scan.index += 1;
	
	for i, char := range scan.data {
		if i == index {
			return string([]int{char});
		}
	}
	
	return "";
}

func (scan *Scanner)StepBack() string {
	scan.index -= 1;
}

// //Double step returns the next two characters
// func (scan *Scanner)DoubleStep() string {
// 	return fmt.Sprintf("%s%s", scan.Step(), scan.Step());
// }

func Tokenize(input io.Reader) string {
	data := string(io.ReadAll(input));
	
	
	//step through each. As soon as a regexp is recognized, turn on a flag. Keep stepping until no regexp's are recognized, then take a step back and tokenize that lexeme
	matches := false;
	tokenized := ""; //where all the final tokens get concatonized together
	lexeme := "";
	
	lines := strings.Split(data, "\n", 0);
	
	for _, line range lines {
		//do something about indentation here
	
		scan := new(Scanner).data = line;
		lexeme.cat(scan.Step());
		if doesMatch(lexeme) { //it matches a regexp
			matches = true;
			fmt.Printf("match...");
		}
		else { //no match
			if matches == true {
				fmt.Printf("woop! lost it. Take a step back...\n");
				matches = false;
				scan.StepBack();
				lexeme.trim(1);
				tokenized.cat(fmt.Sprintf("%s\n", tokenizeLexeme(lexeme)));
			}
		}
		
		tokenized.cat("\n");
	}
	
	return tokenized;
}

func (s *string)cat(str string) {
	dat := make([]byte, len(str)+len(*s));
	
	for i := 0; i < len(*s); i++ {
		dat[i] = (*s)[i];
	}
	for i := len(*s); i < len(dat); i++ {
		dat[i] = str[i-len(*s)];
	}
	
	return &string(dat)
}

func (s *string)trim(num int) { //trims num characters off the end of s
	dat := make([]int, utf8.RuneCountInString(*s)-num);
	
	for i, char := range *s {
		dat[i] = char;
		if i == len(dat)-1 { //get outa here before it segfaults because s has more values than dat has room for
			break
		}
	}
	return string(dat)
}