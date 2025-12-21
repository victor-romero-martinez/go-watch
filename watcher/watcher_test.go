package watcher

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gow/config"
	"gow/internal/testutil"
)

func TestCalculateFileHash(t *testing.T) {
	// 1. Crear archivo temporal
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")

	content1 := []byte("hola mundo")
	if err := os.WriteFile(tmpFile, content1, 0644); err != nil {
		t.Fatal(err)
	}

	// 2. Calcular primer hash
	hash1, err := calculateFileHash(tmpFile)
	if err != nil {
		t.Fatalf("Error al calcular hash 1: %v", err)
	}

	// 3. Modificar contenido y calcular segundo hash
	content2 := []byte("hola mundo modificado")
	if err := os.WriteFile(tmpFile, content2, 0644); err != nil {
		t.Fatal(err)
	}

	hash2, err := calculateFileHash(tmpFile)
	if err != nil {
		t.Fatalf("Error al calcular el hash 2: %v", err)
	}

	// 4. Verificar que son diferentes
	if hash1 == hash2 {
		t.Errorf("Los hashes deberían ser diferentes. Hash: %s", hash1)
	}

	// 5. Restaurar al contenido original y asegurar que el hash se repite
	os.WriteFile(tmpFile, content1, 0644)
	hash3, _ := calculateFileHash(tmpFile)
	if hash1 != hash3 {
		t.Errorf("El hash debería volver a ser el mismo que el original")
	}
}

func TestWatcher_Debounce(t *testing.T) {
	// 1. Configuración de Mocks
	mock := &testutil.MockCommander{}
	cfg := &config.RunnerConfig{
		DefaultTimeout: 2 * time.Second,
	}
	tmpFile := filepath.Join(t.TempDir(), "main.go")
	os.WriteFile(tmpFile, []byte("fmt.PrintLn(1)"), 0644)

	w := NewWatcher(mock, cfg, tmpFile, false)

	// 2. Simular multiples eventos de guardado rápidos
	// Enviamos 5 señales de ejecución
	for i := 0; i < 5; i++ {
		// Simula un pequeño cambio para que el hash varié
		os.WriteFile(tmpFile, []byte(fmt.Sprintf("fmt.Println(%d)", i)), 0644)

		// Disparamos el canal de debounce manualmente o vía Run
		go w.handleDebounce()
		time.Sleep(10 * time.Microsecond) // Muy rápido, menos que los 100ms de debounce
	}

	if mock.GetCallCount() > 1 {
		t.Errorf("Se esperaba 1 ejecución, hubo %d", mock.GetCallCount())
	}
}
