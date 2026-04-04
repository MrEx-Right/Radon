package main

import (
	"flag"
	"fmt"
	"fuzzer/ipc"
	"fuzzer/orchestrator/corpus"
	"fuzzer/orchestrator/mutator"
	"log"
	"os"
	"time"
)

// FuzzerStats holds the real-time telemetry data for the TUI dashboard
type FuzzerStats struct {
	StartTime  time.Time
	Executions uint64
	Crashes    uint64
	Paths      uint64 // The number of unique execution paths discovered
	LastCrash  time.Time
}

// drawUI is a background Goroutine that renders the AFL-style matrix dashboard
func drawUI(stats *FuzzerStats, manager *corpus.Manager) {
	// Clear the entire screen initially
	fmt.Print("\033[H\033[2J")
	
	ticker := time.NewTicker(500 * time.Millisecond)
	for range ticker.C {
		// Move cursor to top-left instead of clearing to prevent screen flickering
		fmt.Print("\033[H")
		
		elapsed := time.Since(stats.StartTime)
		execsPerSec := float64(stats.Executions) / elapsed.Seconds()
		if elapsed.Seconds() < 1 {
			execsPerSec = 0
		}

		lastCrashStr := "none seen yet"
		if !stats.LastCrash.IsZero() {
			lastCrashStr = time.Since(stats.LastCrash).Round(time.Second).String() + " ago"
		}

		// The highly optimized, professional TUI layout
		ui := fmt.Sprintf(`
  ======================================================
                 ☢️  RADON FUZZER v1.0  ☢️
  ======================================================
  
  [ Process Timing ]
    Run time    : %s
    Last crash  : %s
  
  [ Overall Results ]
    Total execs : %d
    Crashes     : %d
    Paths found : %d
  
  [ Corpus Stats ]
    Queue size  : %d seeds
  
  [ Engine Speed ]
    Execs / sec : %.0f execs/sec
  
  ======================================================
`, 
			elapsed.Round(time.Second), 
			lastCrashStr, 
			stats.Executions, 
			stats.Crashes, 
			stats.Paths, 
			manager.QueueSize(), 
			execsPerSec)

		fmt.Print(ui)
	}
}

func main() {
	// 1. Define all CLI arguments before parsing to prevent runtime panics
	targetPtr := flag.String("target", "", "Path to the vulnerable target binary")
	inputDirPtr := flag.String("in", "input", "Directory containing initial seed files")
	outputDirPtr := flag.String("out", "fuzzer_workspace", "Directory to store crashes")
	enginePath := flag.String("engine", "execution-engine/fork-server.out", "Path to the fork server executable") 

	flag.Usage = func() {
		fmt.Printf(`
  ======================================================
                 ☢️  RADON FUZZER v1.0-beta  ☢️
  ======================================================
  An autonomous, coverage-guided white-box fuzzer.

  USAGE:
    ./radon --target <path_to_binary> [OPTIONS]

  OPTIONS:
    --target <path>   (REQUIRED) Path to the instrumented target binary.
    --in <dir>        Directory containing initial seed files (Default: "input").
    --out <dir>       Workspace directory for crashes/queue (Default: "fuzzer_workspace").
    --engine <path>   Path to the Fork Server engine (Default: "execution-engine/fork-server.out").

  EXAMPLES:
    ./radon --target ./test-targets/kurban.out
    ./radon --target ./test-targets/kurban.out --in custom_seeds/
  ======================================================
`)
	}
	
	// 2. Parse flags once all definitions are complete
	flag.Parse()

	if *targetPtr == "" {
		fmt.Println("[-] FATAL: Target binary not specified.")
		fmt.Println("[*] Usage: ./radon --target <path_to_binary> [--in <input_dir>] [--engine <path_to_fork_server>]")
		os.Exit(1)
	}

	// Initialize the Corpus Manager and load initial seeds
	manager := corpus.NewManager(*outputDirPtr)
	if err := manager.LoadSeeds(*inputDirPtr); err != nil {
		log.Fatalf("[-] FATAL: Failed to load seeds: %v", err)
	}

	// Allocate and attach the 64KB Shared Memory for coverage tracking
	shm, err := ipc.CreateSharedMemory()
	if err != nil {
		log.Fatalf("[-] FATAL: Failed to initialize Shared Memory: %v", err)
	}
	defer shm.CleanUp()

	os.Setenv(ipc.ShmEnvVar, fmt.Sprintf("%d", shm.ShmID))

	// Boot up the Execution Engine (Fork Server)
	// (KANKA DİKKAT: Buradaki kaçak enginePath satırını sildik, zaten yukarıda tanımlı!)
	server, err := ipc.NewForkServer(*enginePath, *targetPtr)
	if err != nil {
		log.Fatalf("[-] FATAL: Failed to initialize IPC bridge: %v", err)
	}

	if err := server.Start(); err != nil {
		log.Fatalf("[-] FATAL: Fork Server failed to start: %v", err)
	}

	// Initialize telemetry stats
	stats := &FuzzerStats{
		StartTime: time.Now(),
		Paths:     1, // Base execution path is assumed as 1
	}

	// Ignite the UI thread
	go drawUI(stats, manager)

	// COVERAGE MAPS
	// virginMap: Tracks globally seen edges to identify novel paths
	// zeroMap: Used to quickly zero out the shared memory bitmap via copy()
	virginMap := make([]byte, ipc.MapSize)
	zeroMap := make([]byte, ipc.MapSize)

	// MAIN FUZZING LOOP
	for {
		stats.Executions++
		
		// 1. Clear previous execution traces to avoid trace collisions
		copy(shm.Bitmap, zeroMap)
		
		// Fetch the next payload from the circular queue
		basePayload, err := manager.GetNext()
		if err != nil {
			log.Fatalf("\n[-] ERROR: Queue error: %v", err)
		}

		// Mutate the payload and write it to disk for the target to consume
		mutatedPayload := mutator.Mutate(basePayload)
		os.WriteFile("/dev/shm/radon_cur_input", mutatedPayload, 0644)
		
		// Trigger the Fork Server to execute the target with the mutated payload
		status, err := server.TriggerFuzz()
			if err != nil {
			shm.CleanUp() 
			log.Fatalf("\n[-] ERROR: Fuzz execution failed: %v", err)
		}
		
		// 2. COVERAGE FEEDBACK ANALYSIS - The Fuzzer's "Brain"
		hasNewPath := false
		for i := 0; i < ipc.MapSize; i++ {
			// If an edge was hit during this execution AND it's globally novel:
			if shm.Bitmap[i] > 0 && virginMap[i] == 0 {
				virginMap[i] = 1 // Burn it into the global memory
				hasNewPath = true
			}
		}

		// 3. EVOLUTION: If a novel path was discovered, save the payload as a new seed!
		if hasNewPath {
			stats.Paths++
			manager.SaveSeed(mutatedPayload)
		}
		
		// Handle crashes based on POSIX signal conventions passed from the Fork Server
		// SIGSEGV (11) -> 139, SIGABRT (6) -> 134
		if status == 139 || status == 134 {
			stats.Crashes++
			stats.LastCrash = time.Now()
			crashID := fmt.Sprintf("%d", time.Now().UnixNano())
			manager.SaveCrash(mutatedPayload, crashID)
		}
	}
}