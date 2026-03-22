package ipc

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
)

// ForkServer manages the lifecycle and high-speed IPC with the C-based Execution Engine.
type ForkServer struct {
	cmd        *exec.Cmd
	ctrlPipe   *os.File // Go writes here, C reads (mapped to FD 3)
	statusPipe *os.File // C writes here, Go reads (mapped to FD 4)
	isRunning  bool
}

// NewForkServer initializes the pipes and prepares the C Execution Engine for execution.
func NewForkServer(enginePath, targetPath string) (*ForkServer, error) {
	// 1. Create Control Pipe (Go -> C)
	ctrlReader, ctrlWriter, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create control pipe: %w", err)
	}

	// 2. Create Status Pipe (C -> Go)
	statusReader, statusWriter, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create status pipe: %w", err)
	}

	// 3. Prepare the execution command
	cmd := exec.Command(enginePath, targetPath)

	// Route C's stdout/stderr to our Go console so we can see its printf logs
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Map the pipes to the child process.
	// ExtraFiles appends to standard FDs. 
	// ctrlReader becomes FD 3, statusWriter becomes FD 4.
	cmd.ExtraFiles = []*os.File{ctrlReader, statusWriter}

	return &ForkServer{
		cmd:        cmd,
		ctrlPipe:   ctrlWriter,
		statusPipe: statusReader,
		isRunning:  false,
	}, nil
}

// Start boots the C Execution Engine and waits for the readiness handshake.
func (fs *ForkServer) Start() error {
	if err := fs.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start C engine: %w", err)
	}

	// Await the 4-byte readiness signal (0xDEADBEEF) from C
	var readySignal uint32
	err := binary.Read(fs.statusPipe, binary.LittleEndian, &readySignal)
	if err != nil {
		return fmt.Errorf("handshake failed, C engine unresponsive: %w", err)
	}

	if readySignal != 0xDEADBEEF {
		return fmt.Errorf("invalid handshake signal received: %X", readySignal)
	}

	fs.isRunning = true
	fmt.Println("[+] IPC: Go Orchestrator successfully connected to C Fork Server.")
	return nil
}

// TriggerFuzz sends a command to the C engine to fork and execute the target, 
// then returns the target's exit status.
func (fs *ForkServer) TriggerFuzz() (int32, error) {
	if !fs.isRunning {
		return 0, fmt.Errorf("fork server is not running")
	}

	// 1. Send 4-byte trigger command to C
	triggerCmd := uint32(1)
	if err := binary.Write(fs.ctrlPipe, binary.LittleEndian, triggerCmd); err != nil {
		return 0, fmt.Errorf("failed to send trigger: %w", err)
	}

	// 2. Read the child PID that the C engine just spawned
	var childPid uint32
	if err := binary.Read(fs.statusPipe, binary.LittleEndian, &childPid); err != nil {
		return 0, fmt.Errorf("failed to read child PID: %w", err)
	}

	// 3. Wait for the execution status (Crash? Normal exit?)
	var exitStatus int32
	if err := binary.Read(fs.statusPipe, binary.LittleEndian, &exitStatus); err != nil {
		return 0, fmt.Errorf("failed to read exit status: %w", err)
	}

	return exitStatus, nil
}