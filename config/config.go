package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const DefaultConfigJSON = `{
    "default_timeout_ms": 5000,
    "rules": [
        {
            "extension": ".go",
            "name": "Golang",
            "execution_command": "go",
            "execution_args": ["run", "$FILE"],
            "needs_build": false
        },
        {
            "extension": ".js",
            "name": "Node.js",
            "execution_command": "node",
            "execution_args": ["$FILE"],
            "needs_build": false
        },
        {
            "extension": ".rs",
            "name": "Rust",
            "execution_command": "/bin/sh",
            "execution_args": ["-c", "rustc $FILE -o /tmp/gow_bin && /tmp/gow_bin"],
            "needs_build": true
        }
    ]
}`

type RunnerConfig struct {
	DefaultTimeout time.Duration `json:"default_timeout_ms"`
	Rules          []Rule        `json:"rules"`
}

type Rule struct {
	Extension        string   `json:"extension"`
	Name             string   `json:"name"`
	ExecutionCommand string   `json:"execution_command"`
	ExecutionArgs    []string `json:"execution_args"`
	NeedsBuild       bool     `json:"needs_build"`
}

func LoadConfig(path string) (*RunnerConfig, error) {
	data, err := os.ReadFile(path)

	if os.IsNotExist(err) {
		fmt.Printf("[INFO] Configuración no encontrada. Creando archivo por defecto en: %s\n", path)

		if err := os.WriteFile(path, []byte(DefaultConfigJSON), 0644); err != nil {
			return nil, fmt.Errorf("[ERR ] No se pudo crear el archivo de configuración: %w", err)
		}

		data = []byte(DefaultConfigJSON)
	} else if err != nil {
		return nil, err
	}

	aux := struct {
		DefaultTimeoutMS int64  `json:"default_timeout_ms"`
		Rules            []Rule `json:"rules"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return nil, err
	}

	cfg := &RunnerConfig{
		DefaultTimeout: time.Duration(aux.DefaultTimeoutMS) * time.Millisecond,
		Rules:          aux.Rules,
	}

	// Basic validation and sane defaults
	if cfg.DefaultTimeout <= 0 {
		cfg.DefaultTimeout = 5 * time.Second
	}
	if len(cfg.Rules) == 0 {
		return nil, fmt.Errorf("[ERR ] Config inválido: no hay reglas en %s", path)
	}

	return cfg, nil
}

func SaveConfig(path string, cfg *RunnerConfig) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Esto evita que el JSON se llene de ceros (nanosegundos)
	aux := struct {
		DefaultTimeoutMS int64  `json:"default_timeout_ms"`
		Rules            []Rule `json:"rules"`
	}{
		// Convertimos el tiempo de Go de nuevo a milisegundos simples
		DefaultTimeoutMS: int64(cfg.DefaultTimeout / time.Millisecond),
		Rules:            cfg.Rules,
	}

	enc := json.NewEncoder(file)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "\t")

	return enc.Encode(aux)
}
