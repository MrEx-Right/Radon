# Changelog

All notable changes to the **Radon Fuzzer** project will be documented in this file.

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
