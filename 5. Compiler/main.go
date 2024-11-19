package main

import "fmt"

func main() {
	input := `
    let x = 10;
    let y = 20;
    print(x + y);
    `

	lexer := NewLexer(input)

	fmt.Println("Start Tokenizing...")

	for {
		tok := lexer.NextToken()
		if tok.Type == TokenEOF { // EOF가 나오면 종료
			break
		}
		fmt.Printf("Token Type: %-10s Literal: %s\n", tok.Type, tok.Literal)
	}

	fmt.Println("Tokenizing Complete!")
}
