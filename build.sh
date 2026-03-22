#!/bin/bash

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}[*] Building Radon Fuzzer Suite...${NC}"

# 1. Build the Radon Core Tracer (ASM)
gcc -c compiler/radon-trace.S -o compiler/radon-trace.o
if [ $? -ne 0 ]; then
    echo -e "${RED}[-] FATAL: Failed to compile Radon ASM Tracer!${NC}"
    exit 1
fi

# 2. Build the Radon Runtime Agent (C)
gcc -c compiler/radon-rt.c -o compiler/radon-rt.o
if [ $? -ne 0 ]; then
    echo -e "${RED}[-] FATAL: Failed to compile Radon Runtime Agent!${NC}"
    exit 1
fi

# 3. Build the Radon Compiler Wrapper (radon-cc)
gcc compiler/radon-cc.c -o radon-cc
if [ $? -ne 0 ]; then
    echo -e "${RED}[-] FATAL: Failed to compile radon-cc!${NC}"
    exit 1
fi

# 4. Compile the Execution Engine (Fork Server)
gcc execution-engine/fork-server.c -o execution-engine/fork-server.out
if [ $? -ne 0 ]; then
    echo -e "${RED}[-] FATAL: Execution Engine compilation failed!${NC}"
    exit 1
fi

# 5. Build the Go Orchestrator (radon CLI)
go build -o radon orchestrator/main.go
if [ $? -ne 0 ]; then
    echo -e "${RED}[-] FATAL: Go Orchestrator compilation failed!${NC}"
    exit 1
fi

# YENİ: Adam mermilerini dizebilsin diye şarjörü (input) önceden hazırla!
mkdir -p input

echo -e "${GREEN}[+] Radon successfully built!${NC}"
echo -e "${YELLOW}[*] USAGE INSTRUCTIONS:${NC}"
echo -e " 1. Put your custom seeds into the 'input/' directory (Optional)."
echo -e " 2. Instrument your target : ./radon-cc target.c -o target.out"
echo -e " 3. Start the fuzzing loop : ./radon --target ./target.out\n"