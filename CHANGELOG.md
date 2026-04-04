# Changelog

All notable changes to the **Radon Fuzzer** project will be documented in this file.

## [v1.0.3-beta] - 2026-04-05
### ☢️ The POSIX Purge & IPC Lockdown Update

- **Professional CLI Experience:** Eradicated the default, unpolished command-line help interfaces. Both the Go Orchestrator (`radon`) and the C Compiler Wrapper (`radon-cc`) now feature custom, highly detailed terminal usage menus (`-h` / `--help`) that clearly define arguments, default paths, and practical examples.
- **POSIX Signal Translation:** Fixed a critical flaw where the Go Orchestrator misinterpreted target crashes. The Fork Server now correctly extracts signals via standard POSIX macros (`WIFSIGNALED`, `WTERMSIG`) and forwards standardized exit codes (e.g., `139` for `SIGSEGV`), ensuring 100% accurate crash triage.
- **IPC Pipe Leak / Zombie FD Lockdown:** Patched a severe stability vulnerability in the Fork Server. The fuzzer's communication pipes (`FORKSRV_CTRL_FD`, `FORKSRV_STATUS_FD`) are now strictly closed within the child process *before* `execv` is called. This prevents vulnerable target applications from hijacking or crashing the Fuzzer's internal nervous system.
- **Trampoline False-Positive Elimination:** Toughened the `radon-cc` assembly injection logic to prevent catastrophic `SIGILL` crashes. The tracer now strictly requires a colon (`:`) to confirm label definitions and utilizes an expanded blacklist to ignore non-executable labels like `.LFE` (Function End) and `.LC` (String Constants).
- **Zombie SHM Leak (IPC_RMID):** Resolved a stealthy memory leak where detached 64KB shared memory segments were not being destroyed by the OS. Explicitly defined and enforced the `IPC_RMID` command during the teardown sequence to guarantee absolute memory reclamation.
- **Runtime Panic Prevention:** Fixed a fatal bug in the Go Orchestrator (`main.go`) where calling `flag.Parse()` prematurely caused runtime panics when allocating the Fork Server engine path. The execution and CLI parsing sequence is now strictly ordered.

## [v1.0.2-beta] - 2026-03-29
### 🛡️ The Great Stabilizer & Red Zone Rescue Update

- **System V ABI Red Zone Protection:** Fixed a critical flaw in `radon-cc.c` where injected assembly trampolines clobbered the target's 128-byte Red Zone. Stack pointers are now safely shifted (`leaq -128(%rsp), %rsp`), preventing the tracer from corrupting local variables and causing false-positive `SIGSEGV` crashes.
- **Thread-Safe Corpus Orchestration:** Overhauled the Go Orchestrator's memory model. Implemented strict `sync.Mutex` locking mechanisms between the high-speed fuzzing loop and the TUI dashboard, eliminating fatal Goroutine data races and ensuring UI stability at maximum exec/s.
- **Catastrophic SHM Leak Resolved:** Patched a severe memory leak in `main.go` where fatal execution errors bypassed the `defer` cleanup routines. The 64KB Shared Memory coverage map is now strictly detached and destroyed upon failure, preventing OS-level RAM exhaustion.
- **Blind I/O Redirection Fix:** Corrected the initialization sequence in `fork-server.c`. Target payload descriptors are now verified *before* routing `stdout`/`stderr` to `/dev/null`, ensuring that payload loading failures are properly logged to the Orchestrator instead of failing silently.
- **Trampoline PRNG Collisions:** Fixed an issue in the compiler wrapper where parallel compilations occurring within the same second generated identical basic block IDs. The random seed is now strictly XOR'd with the Process ID (`time(NULL) ^ getpid()`) to guarantee globally unique edge mapping.
- **Dynamic Path Resolution:** Eradicated hardcoded execution paths across the codebase. The Orchestrator and Compiler now support dynamic routing via CLI arguments, allowing Radon to be executed seamlessly from any directory structure.

## [v1.0.1-beta] - 2026-03-27
### ☢️ Havoc Mutation Engine & I/O Optimization Update

- **Havoc Stacking Architecture (v1.0):** Completely overhauled the `mutator.go` engine. Replaced the legacy single-point mutation logic with a high-intensity "Havoc Stacking" pattern. The engine now applies 1-4 concurrent mutation strategies—including Magic Number injection (0xFFFFFFFF, INT_MAX/MIN), block overwriting, and cross-payload block swapping—to maximize code path discovery.
- **Zero-Disk I/O Pipeline (SSD Preservation):** Migrated the entire mutation-to-execution workflow from physical storage to a high-performance RAM-based architecture. By utilizing the Linux `tmpfs` (`/dev/shm`), Radon now eliminates SSD wear-and-tear and achieves near-zero latency during payload delivery to the target.
- **Atomic Assembly Instrumentation Fix:** Resolved a critical "Register Clobbering" vulnerability in `radon-cc.c`. Implemented manual stack preservation (`pushq/popq %rcx`) around the `__radon_trace` trampoline calls to prevent architectural side-effects and eliminate false-positive Segmentation Faults during instrumentation.
- **Target STDIN Routing (Stream-Aware):** Re-engineered the `fork-server.c` I/O redirection module. The execution engine now performs active `dup2` descriptor mapping to pipe mutated RAM-disk payloads directly into the target's `STDIN`, ensuring full compatibility with command-line utilities and interactive binaries.
- **Professional Metadata & Identity Mapping:** Standardized project identity and documentation structures. Integrated advanced `README.md` technical specifications and aligned the repository for enterprise-grade collaborative fuzzing workflows.

## [1.0.0-beta] - 2026-03-22

### 🚀 Added
- **Core Fuzzing Engine:** Initial release of the autonomous, coverage-guided fuzzing architecture.
- **Compiler Wrapper (`radon-cc`):** Intercepts standard `gcc` calls to seamlessly inject instrumentation into target binaries.
- **Pure Assembly Tracer (`radon-trace.S`):** Nanosecond-level execution edge tracking written in pure x86_64 Assembly without destroying CPU registers.
- **Go Orchestrator:** The main "brain" of the fuzzer, managing mutations, payload delivery, and coverage analysis.
- **High-Performance Fork Server:** OS-level process cloning mechanism in C to bypass `execve()` overhead and achieve massive executions per second.
- **Autonomous Feedback Loop:** Added genetic algorithm logic where Radon evaluates the 64KB Shared Memory bitmap to discover and save novel execution paths automatically.
- **Matrix-Style TUI:** Real-time terminal dashboard displaying execution speed, path discoveries, queue size, and crash metrics.
- **Auto-Seed Generation:** Smart initialization that automatically creates a default starting payload if the `input/` directory is empty.
- **Build System:** Added `build.sh` to compile C, Go, and Assembly components together seamlessly.

### 🛡️ Security
- Workspace isolation: Crashes and queues are strictly separated in the `fuzzer_workspace` directory to prevent data corruption during high-speed runs.
