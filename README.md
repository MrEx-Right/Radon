<div align="center">
  <h1>☢️ RADON FUZZER</h1>
  <p><b>An autonomous, coverage-guided white-box fuzzer built from scratch in C, Go, and pure x86_64 Assembly.</b></p>
  
  <p align="center">
    <img src="https://img.shields.io/badge/version-Beta-FF8C00?style=flat-square" alt="Version">
    <img src="https://img.shields.io/badge/build-passing-4c1?style=flat-square" alt="Build">
    <img src="https://img.shields.io/badge/license-GPLv3-0059b3?style=flat-square" alt="License">
    <img src="https://img.shields.io/badge/language-C_%7C_Go_%7C_ASM-00ADD8?style=flat-square" alt="Languages">
    <img src="https://img.shields.io/badge/platform-Linux-lightgrey?style=flat-square" alt="Platform">
  </p>
</div>

---

## ⚡ Overview
Radon is a high-performance vulnerability discovery engine. It injects lightweight coverage-tracking payloads into target binaries during compilation and uses a highly optimized Genetic Algorithm via a Go-based Orchestrator to navigate through the target's execution paths, finding crashes at breakneck speeds.

Unlike standard wrappers, Radon implements its own instrumentation, IPC bridging, and mutation engines from the ground up.

## 🔥 Key Features
* **Custom Compiler Wrapper (`radon-cc`)**: Intercepts GCC to inject AFL-style XOR coverage trampolines directly into the target's assembly graph.
* **Pure Assembly Tracer (`radon-trace.S`)**: A nanosecond-level instrumentation engine that tracks execution edges in a 64KB Shared Memory map without disrupting CPU flags.
* **High-Speed Fork Server**: Clones the target process at the OS level to avoid `execve()` overhead, achieving massive Execs/Sec.
* **Autonomous Feedback Loop**: The Go Orchestrator acts as the "brain," analyzing the coverage map to learn new paths and automatically saving interesting mutations back into the execution queue.
* **Matrix-Style TUI**: Real-time terminal dashboard tracking crashes, queue size, path discoveries, and execution speed.
* **Auto-Seed Generation**: No inputs provided? No problem. Radon dynamically generates its own starting payload if the `input` directory is empty.

## 🧠 Architecture
1. **The Factory (`radon-cc`)**: Injects the Radon Runtime (`radon-rt`) and Assembly Tracer into the target's source code.
2. **The Matrix (`Shared Memory`)**: A 64KB bitmap connecting the Fuzzer and the Target for real-time edge coverage feedback.
3. **The Brain (`Orchestrator`)**: Written in Go. Mutates the payloads, evaluates coverage, stores crashes, and renders the UI.

## 🚀 Quick Start

### 1. Build the Fuzzer Suite
Clone the repository and compile the Radon toolkit. The build script automatically compiles the C components, the ASM tracer, and the Go orchestrator.
```bash
git clone [https://github.com/MrEx-Right/Radon.git](https://github.com/MrEx-Right/Radon.git)
cd Radon
./build.sh
```
### 2. Instrument Your Target
Use `radon-cc` instead of `gcc` to compile your vulnerable C program. This will embed Radon's tracking agents into the binary:
```bash
./radon-cc test-targets/kurban.c -o kurban.out
```

### 3. Unleash the Swarm
Start the fuzzing loop. Radon will automatically generate the required workspace and start hunting for crashes.
```bash
./radon --target ./kurban.out
```

*(Optional): Place your custom seed files in the `input/` directory before running Radon to give it a head start.*

## 📂 Crash Triage
When Radon successfully breaks the target, the crashing payloads (causing `SIGSEGV` or `SIGABRT`) will be automatically saved in:
`./fuzzer_workspace/crashes/`

---
**⚠️ Disclaimer:** *Developed for educational purposes and vulnerability research. Do not use on targets you do not own or have explicit permission to test.*
