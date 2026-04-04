package ipc

import (
	"fmt"
	"syscall"
	"unsafe"
)

const MapSize = 64 * 1024
const ShmEnvVar = "EXCALIBUR_SHM_ID"

type SharedMemory struct {
	ShmID int
	Addr  unsafe.Pointer
	Bitmap []byte
}

// CreateSharedMemory initializes a new SysV shared memory segment.
func CreateSharedMemory() (*SharedMemory, error) {
	// Create shared memory segment (IPC_PRIVATE)
	shmid, _, err := syscall.Syscall(syscall.SYS_SHMGET, 0, uintptr(MapSize), 0600|01000)
	if int(shmid) == -1 {
		return nil, fmt.Errorf("shmget failed: %w", err)
	}

	// Attach the segment to our process address space
	addr, _, err := syscall.Syscall(syscall.SYS_SHMAT, shmid, 0, 0)
	if int(addr) == -1 {
		return nil, fmt.Errorf("shmat failed: %w", err)
	}

	// Map it to a Go slice for easy manipulation
	bitmap := (*[MapSize]byte)(unsafe.Pointer(addr))[:]

	return &SharedMemory{
		ShmID:  int(shmid),
		Addr:   unsafe.Pointer(addr),
		Bitmap: bitmap,
	}, nil
}

// CleanUp detaches and removes the shared memory from the system.
const IPC_RMID = 0

// CleanUp detaches and removes the shared memory from the system.
// Crucial to prevent zombie memory segments (leaks) after the fuzzer halts.
func (shm *SharedMemory) CleanUp() {
	// 1. Detach the shared memory segment from our process address space
	syscall.Syscall(syscall.SYS_SHMDT, uintptr(shm.Addr), 0, 0)
	
	// 2. Mark the segment to be destroyed via IPC_RMID
	// By explicitly issuing IPC_RMID (0), the OS guarantees the memory is freed.
	syscall.Syscall(syscall.SYS_SHMCTL, uintptr(shm.ShmID), uintptr(IPC_RMID), 0)
}