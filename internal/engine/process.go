package engine

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

// Process wraps the Stockfish executable to handle UCI communication.
type Process struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
}

// Start spawns the Stockfish process and establishes communication.
func Start(path string) (*Process, error) {
	cmd := exec.Command(path)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	stdout := bufio.NewReader(stdoutPipe)

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start engine: %w", err)
	}

	p := &Process{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
	}

	// 3.2 Implement UCI Handshake
	if err := p.handshake(); err != nil {
		p.Close()
		return nil, err
	}

	return p, nil
}

// Close kills the engine process to prevent memory leaks.
func (p *Process) Close() {
	p.SendCommand("quit")
	p.cmd.Process.Kill()
	p.cmd.Wait()
}

// SendCommand writes a command to the engine's stdin.
func (p *Process) SendCommand(cmd string) error {
	_, err := fmt.Fprintln(p.stdin, cmd)
	return err
}

// ReadUntil reads from the engine's stdout until it finds a line containing the expected string.
func (p *Process) ReadUntil(expected string, timeout time.Duration) (string, error) {
	resultCh := make(chan string)
	errCh := make(chan error)

	go func() {
		for {
			line, err := p.stdout.ReadString('\n')
			if err != nil {
				errCh <- err
				return
			}
			line = strings.TrimSpace(line)
			if strings.Contains(line, expected) {
				resultCh <- line
				return
			}
		}
	}()

	select {
	case line := <-resultCh:
		return line, nil
	case err := <-errCh:
		return "", err
	case <-time.After(timeout):
		return "", fmt.Errorf("timeout waiting for engine to respond with '%s'", expected)
	}
}

// handshake performs the standard UCI initialization.
func (p *Process) handshake() error {
	if err := p.SendCommand("uci"); err != nil {
		return err
	}
	if _, err := p.ReadUntil("uciok", 5*time.Second); err != nil {
		return err
	}

	if err := p.SendCommand("isready"); err != nil {
		return err
	}
	if _, err := p.ReadUntil("readyok", 5*time.Second); err != nil {
		return err
	}

	return nil
}

// Configure sets the engine's difficulty and resource limits.
func (p *Process) Configure(skillLevel int) error {
	// Limit footprint
	p.SendCommand("setoption name Threads value 1")
	p.SendCommand("setoption name Hash value 16")

	// Set difficulty (0-20)
	p.SendCommand(fmt.Sprintf("setoption name Skill Level value %d", skillLevel))

	return nil
}

// GetBestMove asks the engine for the best move in the current position.
func (p *Process) GetBestMove(fen string) (string, error) {
	// Sync state
	if err := p.SendCommand(fmt.Sprintf("position fen %s", fen)); err != nil {
		return "", err
	}

	// Ask for move, giving it up to 1000ms to think
	if err := p.SendCommand("go movetime 1000"); err != nil {
		return "", err
	}

	// Read until "bestmove"
	line, err := p.ReadUntil("bestmove", 3*time.Second)
	if err != nil {
		return "", err
	}

	// Extract the move (e.g., "bestmove e2e4 ponder e7e5")
	parts := strings.Split(line, " ")
	if len(parts) >= 2 {
		return parts[1], nil
	}

	return "", fmt.Errorf("could not parse bestmove from line: %s", line)
}
