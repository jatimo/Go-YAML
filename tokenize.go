package yaml

import (
	"io";
	"strings";
	"regexp";
	"fmt";
	"utf8";
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
	needsString bool; //tells whether or not to run (S)Printf(exp, in), or just (S)Printf(exp)
}

var matches = []LexicalMatch { //In order of precedence!
	LexicalMatch { ">", "FOLDED", false },
	LexicalMatch { "|", "LITERAL", false },
	LexicalMatch { "&", "ANCHOR", false },
	LexicalMatch { "\\*", "ALIAS", false },
	LexicalMatch { "-", "ENTRY", false },
	LexicalMatch { ":", "MAPVALUE", false}, //Map value
	//LexicalMatch { "[\\-\\+]?(\\.[0-9]+|[0-9]+(\\.[0-9]*)?)([eE][-\\+]?[0-9]+)?", "FLOAT %s" }, //FLOAT
	//LexicalMatch { "[\\-\\+]?[0-9]+", "INT %s" }, //INT (base 10)
	LexicalMatch { "[^:^&^*^ ^\\-]+( *[^:^&^*^ ^\\-])*", "STRING %s", true }, //STRING
}

func doesMatch(s string) bool {
	for _, match := range matches {
		rx := regexp.MustCompile(match.exp);
		res := rx.ExecuteString(s);
		if len(res) == 0 {
			continue
		}
		for i := 0; i < len(res)/2; i += 2 {
			if res[i] == 0 && res[i + 1] == utf8.RuneCountInString(s) {
				return true
			}
		}
	}
	return false
}

func tokenizeLexeme(s string) string {
	for _, match := range matches {
		rx := regexp.MustCompile(match.exp);
		res := rx.ExecuteString(s);
		if len(res) == 0 {
			continue
		}
		for i := 0; i < len(res)/2; i += 2 {
			if res[i] == 0 && res[i + 1] == utf8.RuneCountInString(s) {
				if match.needsString {
					return fmt.Sprintf(match.out, s);
				}
				else {
					return fmt.Sprintf(match.out);
				}
			}
		}
	}
	return s; //this is bad
}

type Scanner struct {
	data string;
	index int;
}

//Step returns the next character in the character stream. Returns "" if no characters left
func (scan *Scanner)Step() (string, bool) {
	
	for i, char := range scan.data {
		if i == scan.index {
			scan.index += 1;
			return string([]int{char}), true;
		}
	}
	
	scan.index += 1;
	
	return "", false;
}

func (scan *Scanner)StepBack() {
	scan.index -= 1;
}

// //Double step returns the next two characters
// func (scan *Scanner)DoubleStep() string {
// 	return fmt.Sprintf("%s%s", scan.Step(), scan.Step());
// }

func Tokenize(input io.Reader) string {
	bytedata, err := io.ReadAll(input);
	if err != nil {
		//panic
	}
	data := string(bytedata);
	
	
	//step through each. As soon as a regexp is recognized, turn on a flag. Keep stepping until no regexp's are recognized, then take a step back and tokenize that lexeme
	matches := false;
	tokenized := ""; //where all the final tokens get concatonized together
	lexeme := "";
	
	lines := strings.Split(data, "\n", 0);
	
// 	for _, line := range lines {
// 		fmt.Printf("%s\n", line);
// 	}

	var indentDefines = make([]int, 2);
	indentDefines[0] = 0;
	indentDefines[0] = 0;
	
	for _, line := range lines {
		//do something about indentation here
		spaces := numLeadingSpaces(line);
		var indent string;
		indent, indentDefines = calcIndent(spaces, indentDefines);
		
		
		scan := new(Scanner);
		scan.data = line;
		
		for  {			
			trail, still := scan.Step();
			
			if !still { //at the last character, do something with it or forget it forever :)
				tokenized = cat(tokenized, fmt.Sprintf("%s ", tokenizeLexeme(lexeme)));
				//fmt.Printf("adding %s\n", fmt.Sprintf("%s ", tokenizeLexeme(lexeme)));
				lexeme = "";
				break
			}
			
			lexeme = cat(lexeme, trail);
			
			//TODO: this is a crude solution: make it better?
			if lexeme == " " {
				lexeme = ""; 
				continue
			}

			//fmt.Printf("lexeme='%s'\n", lexeme);
			if doesMatch(lexeme) { //it matches a regexp
				matches = true;
				//fmt.Printf("%s match...\n", lexeme);
			}
			else { //no match
				if matches == true {
					//fmt.Printf("woop! lost it. Take a step back...\n");
					matches = false;
					scan.StepBack();
					lexeme = trim(lexeme, 1);
					tokenized = cat(tokenized, fmt.Sprintf("%s ", tokenizeLexeme(lexeme)));
					//fmt.Printf("adding %s\n", fmt.Sprintf("%s ", tokenizeLexeme(lexeme)));
					lexeme = "";
				}
			}
		}
		fmt.Printf("%d leading spaces.\n", spaces);
		fmt.Printf("'%s' new spaces.\n", indent);
		tokenized = fmt.Sprintf("%s%s\n", indent, tokenized);
		//break
	}
	
	return tokenized;
}

func calcIndent(numSpaces int, indents []int) (string, []int) { //returns a string with the appropriate number of spaces
	
	
	for i := 0; i < len(indents)/2; i += 2 {
		fmt.Printf("numSpaces=%d\n", numSpaces);
		if (indents)[i] == numSpaces {
			fmt.Printf("Found indent match ; num=%d\n", indents[i+1]);
			return repeatString(" ", indents[i+1]), indents;
		}
	}
	fmt.Printf("No match, so make it\n");
	biggest := 0;
	for i := 0; i < len(indents)/2; i += 2 {
		if indents[i+1] > biggest {
			biggest = indents[i+1];
			fmt.Printf("Biggest = %d\n", biggest);
		}
	}
	//not in there
	bigger := make([]int, len(indents)+2);
	
	for i := 0; i < len(indents); i++ {
		bigger[i] = (indents)[i];
	}
	
	//going to assume that the yml is consistent among indents, and that any new ones are going to be BIGGER
	//this is probably not a safe assumption
	//TODO FIX THIS!
	//(the reason this is non-trivial is that interpolating indents may require a bit more of an intelligent indent determining algorithm)
	
	bigger[0] = numSpaces;
	bigger[1] = biggest+1;
	indent, bigger := calcIndent(numSpaces, bigger);
	return indent, bigger;
}

func repeatString(s string, times int) string {
	orig := s;
	for i := 0; i < times-1; i++ {
		s = cat(s, orig);
	}
	
	if times == 0 {
		return "";
	}
	
	return s;
}

func numLeadingSpaces(s string) int {
	for i, char := range s {
		if char != int(' ') {
			return i;
		}
	}
	
	return utf8.RuneCountInString(s);
}

func cat(s string, str string) string{
	dat := make([]byte, len(str)+len(s));
	
	for i := 0; i < len(s); i++ {
		dat[i] = s[i];
	}
	for i := len(s); i < len(dat); i++ {
		dat[i] = str[i-len(s)];
	}
	
	return string(dat)
}

func trim(s string, num int) string { //trims num characters off the end of s
	dat := make([]int, utf8.RuneCountInString(s)-num);
	
	for i, char := range s {
		dat[i] = char;
		if i == len(dat)-1 { //get outa here before it segfaults because s has more values than dat has room for
			break
		}
	}
	return string(dat)
}