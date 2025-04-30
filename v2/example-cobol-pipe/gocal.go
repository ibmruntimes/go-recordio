package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

type SpawnWork struct {
	Cmd        *exec.Cmd
	Stdin      chan []byte
	Stdout     chan []byte
	Stderr     chan []byte
	StdinPipe  io.WriteCloser
	StdoutPipe io.ReadCloser
	StderrPipe io.ReadCloser
}

func NewSpawn(cmd *exec.Cmd) *SpawnWork {
	return &SpawnWork{
		Cmd:    cmd,
		Stdin:  make(chan []byte),
		Stdout: make(chan []byte),
		Stderr: make(chan []byte),
	}
}
func (cs *SpawnWork) Start() error {
	var err error

	cs.StdinPipe, err = cs.Cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	cs.StdoutPipe, err = cs.Cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	cs.StderrPipe, err = cs.Cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	go func() {
		for data := range cs.Stdin {
			_, err := cs.StdinPipe.Write(data)
			if err != nil {
				fmt.Printf("Error writing to stdin: %v\n", err)
				break
			}
		}
		cs.StdinPipe.Close()
	}()

	go func() {
		scanner := bufio.NewScanner(cs.StdoutPipe)
		for scanner.Scan() {
			cs.Stdout <- []byte(scanner.Text() + "\n")
		}
		close(cs.Stdout)
	}()

	go func() {
		scanner := bufio.NewScanner(cs.StderrPipe)
		for scanner.Scan() {
			cs.Stderr <- []byte(scanner.Text() + "\n")
		}
		close(cs.Stderr)
	}()

	if err := cs.Cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	return nil
}

func (cs *SpawnWork) Wait() error {
	return cs.Cmd.Wait()
}

type Token struct {
	Type  string
	Value string
}

func Lexer(input string) ([]Token, error) {
	var tokens []Token
	var buffer strings.Builder

	for i := 0; i < len(input); i++ {
		char := rune(input[i])

		if unicode.IsSpace(char) {
			continue
		}

		if unicode.IsDigit(char) || char == '.' {
			buffer.WriteRune(char)
			for i+1 < len(input) && (unicode.IsDigit(rune(input[i+1])) || input[i+1] == '.') {
				i++
				buffer.WriteByte(input[i])
			}
			tokens = append(tokens, Token{Type: "NUMBER", Value: buffer.String()})
			buffer.Reset()
			continue
		}

		if char == '+' || char == '-' || char == '*' || char == '/' || char == '(' || char == ')' {
			tokens = append(tokens, Token{Type: "OPERATOR", Value: string(char)})
			continue
		}

		return nil, fmt.Errorf("invalid character: %c", char)
	}

	return tokens, nil
}

func Parser(tokens []Token) error {
	parenCount := 0

	for i, token := range tokens {
		switch token.Type {
		case "NUMBER":
			if _, err := strconv.ParseFloat(token.Value, 64); err != nil {
				return fmt.Errorf("invalid number: %s", token.Value)
			}
		case "OPERATOR":
			if token.Value == "(" {
				parenCount++
			} else if token.Value == ")" {
				parenCount--
				if parenCount < 0 {
					return fmt.Errorf("mismatched parentheses")
				}
			} else if i == 0 || i == len(tokens)-1 {
				return fmt.Errorf("operator at invalid position")
			}
		}
	}

	if parenCount != 0 {
		return fmt.Errorf("mismatched parentheses")
	}

	return nil
}

func precedence(op string) int {
	switch op {
	case "+", "-":
		return 1
	case "*", "/":
		return 2
	default:
		return 0
	}
}

func InfixToPostfix(tokens []Token) []Token {
	var postfix []Token
	var stack []Token

	for _, token := range tokens {
		switch token.Type {
		case "NUMBER":
			postfix = append(postfix, token)
		case "OPERATOR":
			if token.Value == "(" {
				stack = append(stack, token)
			} else if token.Value == ")" {
				for len(stack) > 0 && stack[len(stack)-1].Value != "(" {
					postfix = append(postfix, stack[len(stack)-1])
					stack = stack[:len(stack)-1]
				}
				if len(stack) > 0 {
					stack = stack[:len(stack)-1]
				}
			} else {
				for len(stack) > 0 && precedence(stack[len(stack)-1].Value) >= precedence(token.Value) {
					postfix = append(postfix, stack[len(stack)-1])
					stack = stack[:len(stack)-1]
				}
				stack = append(stack, token)
			}
		}
	}

	for len(stack) > 0 {
		postfix = append(postfix, stack[len(stack)-1])
		stack = stack[:len(stack)-1]
	}

	return postfix
}

func main() {
	me, err := os.Executable()
	if err != nil {
		log.Fatalf("Error getting executable path: %v", err)
	}
	task1 := NewSpawn(exec.Command(filepath.Join(filepath.Dir(me), "pfix")))
	err = task1.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fail to start %v\n", err)
		return
	}
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter arithmetic expressions, e.g.\n 200.6 + (232.34/3.1) <enter>\n(Ctrl+D to exit):")
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := scanner.Text()

		tokens, err := Lexer(input)
		if err != nil {
			fmt.Println("Syntax error:", err)
			continue
		}

		err = Parser(tokens)
		if err != nil {
			fmt.Println("Syntax error:", err)
			continue
		}

		postfix := InfixToPostfix(tokens)

		task1.Stdin <- []byte("clear\n")
		for _, token := range postfix {
			task1.Stdin <- []byte(token.Value + "\n")
		}
		task1.Stdin <- []byte("stacktop\n")
		msg, ok := <-task1.Stdout
		if ok {
			fmt.Printf("Result from COBOL decimal calculator: %v\n", string(msg))
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading input:", err)
	}
	task1.Stdin <- []byte("quit\n")
	task1.Wait()
	fmt.Printf("Bye\n")

}
