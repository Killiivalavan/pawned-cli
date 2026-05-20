package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

func readUntilMatch(reader *bufio.Reader, pattern *regexp.Regexp, timeout time.Duration) (string, error) {
	ch := make(chan string, 1)
	errCh := make(chan error, 1)
	go func() {
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					errCh <- err
				}
				return
			}
			// fmt.Print(">> ", line) // Debug
			if match := pattern.FindStringSubmatch(line); match != nil {
				ch <- match[1]
				return
			}
		}
	}()

	select {
	case res := <-ch:
		return res, nil
	case err := <-errCh:
		return "", err
	case <-time.After(timeout):
		return "", fmt.Errorf("timeout waiting for match: %s", pattern.String())
	}
}

func waitForPrompt(reader *bufio.Reader, prompt string, timeout time.Duration) error {
	ch := make(chan struct{}, 1)
	go func() {
		buf := make([]byte, 1024)
		var output string
		for {
			n, err := reader.Read(buf)
			if n > 0 {
				output += string(buf[:n])
				// fmt.Print(string(buf[:n])) // Debug
				if strings.Contains(output, prompt) {
					ch <- struct{}{}
					return
				}
			}
			if err != nil {
				return
			}
		}
	}()

	select {
	case <-ch:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("timeout waiting for prompt: %s", prompt)
	}
}

func main() {
	fmt.Println("Starting Player 1 (Creator)...")
	p1 := exec.Command("./chesshell-cli", "play", "--multiplayer")
	
	p1In, _ := p1.StdinPipe()
	p1Out, _ := p1.StdoutPipe()
	
	if err := p1.Start(); err != nil {
		fmt.Printf("Failed to start Player 1: %v\n", err)
		os.Exit(1)
	}
	defer p1.Process.Kill()

	p1Reader := bufio.NewReader(p1Out)
	
	// Wait for code
	codeRegex := regexp.MustCompile(`Your game code is: (CH-[A-Z0-9]{4})`)
	code, err := readUntilMatch(p1Reader, codeRegex, 5*time.Second)
	if err != nil {
		fmt.Printf("Failed to get game code from Player 1: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Got code: %s\n", code)

	fmt.Println("Starting Player 2 (Joiner)...")
	p2 := exec.Command("./chesshell-cli", "play", "--join", code)
	p2In, _ := p2.StdinPipe()
	p2Out, _ := p2.StdoutPipe()

	if err := p2.Start(); err != nil {
		fmt.Printf("Failed to start Player 2: %v\n", err)
		os.Exit(1)
	}
	defer p2.Process.Kill()
	p2Reader := bufio.NewReader(p2Out)

	// Wait for Player 1 to see the prompt
	fmt.Println("Waiting for Player 1 prompt...")
	if err := waitForPrompt(p1Reader, "Your move (UCI) ->", 10*time.Second); err != nil {
		fmt.Printf("Player 1 failed to reach prompt: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Player 1 (White) is ready to move.")

	// Player 1 plays e2e4
	fmt.Println("Player 1 plays e2e4")
	fmt.Fprintln(p1In, "e2e4")

	// Wait for Player 2 to see the prompt
	fmt.Println("Waiting for Player 2 prompt...")
	if err := waitForPrompt(p2Reader, "Your move (UCI) ->", 10*time.Second); err != nil {
		fmt.Printf("Player 2 failed to reach prompt: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Player 2 (Black) received move and is ready to reply.")

	// Player 2 plays e7e5
	fmt.Println("Player 2 plays e7e5")
	fmt.Fprintln(p2In, "e7e5")

	// Wait for Player 1 to see the prompt again
	fmt.Println("Waiting for Player 1 prompt...")
	if err := waitForPrompt(p1Reader, "Your move (UCI) ->", 10*time.Second); err != nil {
		fmt.Printf("Player 1 failed to receive reply: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✅ MULTIPLAYER TEST PASSED!")
	fmt.Println("Both players successfully connected to the room and exchanged moves via the Cloudflare Worker relay.")
	
	fmt.Fprintln(p1In, "quit")
	fmt.Fprintln(p2In, "quit")
	
	time.Sleep(1 * time.Second)
}
