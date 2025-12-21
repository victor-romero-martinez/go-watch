package testutil

import (
	"context"
	"sync"
)

type MockCommander struct {
	mu          sync.Mutex
	CallCount   int
	LastCommand string
	LastArgs    []string
	ReturnErr   error
}

func (m *MockCommander) Run(ctx context.Context, command string, args []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CallCount++
	m.LastCommand = command
	m.LastArgs = args

	return m.ReturnErr
}

// Permite leer el contador de forma segura desde los test.
func (m *MockCommander) GetCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.CallCount
}
