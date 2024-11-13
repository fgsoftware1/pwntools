package main

import (
	"bufio"
	"debug/elf"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"time"
)

// ELF represents an ELF binary
type ELF struct {
	Path string
	File *elf.File
}

// NewELF creates a new ELF instance
func NewELF(path string) (*ELF, error) {
	file, err := elf.Open(path)
	if err != nil {
		return nil, err
	}
	return &ELF{Path: path, File: file}, nil
}

// Remote represents a remote connection
type Remote struct {
	conn net.Conn
}

// NewRemote creates a new remote connection
func NewRemote(host string, port int) (*Remote, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, err
	}
	return &Remote{conn: conn}, nil
}

// SendLine sends a line of text to the remote connection
func (r *Remote) SendLine(data string) error {
	_, err := r.conn.Write([]byte(data + "\n"))
	return err
}

// Recv receives data from the remote connection
func (r *Remote) Recv(size int) ([]byte, error) {
	buf := make([]byte, size)
	n, err := r.conn.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

// Interactive starts an interactive shell with the remote connection
func (r *Remote) Interactive() error {
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			r.SendLine(scanner.Text())
		}
	}()

	for {
		buf := make([]byte, 1024)
		n, err := r.conn.Read(buf)
		if err != nil {
			return err
		}
		fmt.Print(string(buf[:n]))
	}
}

// Process represents a local process
type Process struct {
	cmd    *exec.Cmd
	stdin  *bufio.Writer
	stdout *bufio.Reader
}

// NewProcess creates a new local process
func NewProcess(path string, args ...string) (*Process, error) {
	cmd := exec.Command(path, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	return &Process{
		cmd:    cmd,
		stdin:  bufio.NewWriter(stdin),
		stdout: bufio.NewReader(stdout),
	}, nil
}

// SendLine sends a line of text to the process
func (p *Process) SendLine(data string) error {
	_, err := p.stdin.WriteString(data + "\n")
	if err != nil {
		return err
	}
	return p.stdin.Flush()
}

// Recv receives data from the process
func (p *Process) Recv(size int) ([]byte, error) {
	buf := make([]byte, size)
	n, err := p.stdout.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

// Interactive starts an interactive shell with the process
func (p *Process) Interactive() error {
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			p.SendLine(scanner.Text())
		}
	}()

	for {
		buf := make([]byte, 1024)
		n, err := p.stdout.Read(buf)
		if err != nil {
			return err
		}
		fmt.Print(string(buf[:n]))
	}
}
