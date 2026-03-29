package corpus

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// FuzzInput represents a single test case in the corpus
type FuzzInput struct {
	Data     []byte
	Filename string
	IsSeed   bool
}

// Manager handles the input queue and crash storage
type Manager struct {
	mu           sync.Mutex 
	Queue        []FuzzInput
	CurrentIndex int
	OutputDir    string
}

// NewManager initializes the corpus and crash directories
func NewManager(outputDir string) *Manager {
	os.MkdirAll(filepath.Join(outputDir, "crashes"), 0755)
	os.MkdirAll(filepath.Join(outputDir, "queue"), 0755)

	return &Manager{
		Queue:        make([]FuzzInput, 0),
		CurrentIndex: 0,
		OutputDir:    outputDir,
	}
}

// LoadSeeds reads initial test cases from the input directory
func (cm *Manager) LoadSeeds(inputDir string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	os.MkdirAll(inputDir, 0755)

	files, err := os.ReadDir(inputDir)
	if err != nil {
		return fmt.Errorf("failed to read input directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		path := filepath.Join(inputDir, file.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		cm.Queue = append(cm.Queue, FuzzInput{
			Data:     data,
			Filename: file.Name(),
			IsSeed:   true, 
		})
	}

	if len(cm.Queue) == 0 {
		defaultSeed := []byte("RADON_DEFAULT_SEED_12345")
		defaultPath := filepath.Join(inputDir, "default_seed.txt")
		os.WriteFile(defaultPath, defaultSeed, 0644)
		
		cm.Queue = append(cm.Queue, FuzzInput{
			Data:     defaultSeed,
			Filename: "default_seed.txt",
			IsSeed:   true,
		})
	}

	return nil
}

// GetNext retrieves the next payload from the circular queue
func (cm *Manager) GetNext() ([]byte, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if len(cm.Queue) == 0 {
		return nil, fmt.Errorf("FATAL: Execution queue is empty. No seeds provided")
	}

	input := cm.Queue[cm.CurrentIndex]
	cm.CurrentIndex = (cm.CurrentIndex + 1) % len(cm.Queue)
	
	return input.Data, nil
}

// SaveCrash writes a crashing payload to the output directory
func (cm *Manager) SaveCrash(data []byte, crashID string) {
	
	filename := fmt.Sprintf("crash_%s.bin", crashID)
	path := filepath.Join(cm.OutputDir, "crashes", filename)
	os.WriteFile(path, data, 0644)
}

func (cm *Manager) SaveSeed(data []byte) {
	cm.mu.Lock()
	filename := fmt.Sprintf("id_%06d", len(cm.Queue))
	
	cm.Queue = append(cm.Queue, FuzzInput{
		Data:     data,
		Filename: filename,
		IsSeed:   false, 
	})
	cm.mu.Unlock() 
	
	
	path := filepath.Join(cm.OutputDir, "queue", filename)
	os.WriteFile(path, data, 0644)
}


func (cm *Manager) QueueSize() int {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	return len(cm.Queue)
}