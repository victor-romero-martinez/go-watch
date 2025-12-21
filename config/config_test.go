package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_AutoGenerate(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "gow.config.json")

	// 1. Intentar cargar el archivo que NO existe
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("No debería dar error, debería crearlo: %v", err)
	}

	// 2. Verificar que el archivo exista físicamente
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("El archivo gow.config.json no fue creado automáticamente")
	}

	// 3. Verificar contenido básico
	if cfg.DefaultTimeout.Seconds() != 5 {
		t.Errorf("Se esperaba timeout de 5s, se obtuvo: %v", cfg.DefaultTimeout)
	}

	if len(cfg.Rules) < 1 {
		t.Error("La configuración generada debería tener reglas por defecto")
	}
}
