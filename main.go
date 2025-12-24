package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gow/config"
	"gow/runner"
	"gow/watcher"
)

var (
	targetFile  string
	configPath  string
	timeout     int
	verbose     bool
	delay       int
	showVersion bool
	version     string
	addRuleRaw  string
)

const ConfigFileName = "gow.config.json"

func init() {
	flag.StringVar(&targetFile, "f", "", "El archivo a vigilar y ejecutar (ej. main.go).")
	flag.StringVar(&configPath, "c", "", "Ruta a un archivo de configuración personalizada.")
	flag.IntVar(&timeout, "t", 5, "Timeout de ejecución en segundos para evitar bucles infinitos.")
	flag.BoolVar(&verbose, "v", false, "Habilita la salida detallada (verbose) de logs internos.")
	flag.IntVar(&delay, "delay", 100, "Retraso de debounce en milisegundos después de la detección de un cambio.")
	flag.StringVar(&addRuleRaw, "a", "", "Agrega un lenguaje: '.ext;Nombre;Comando;Arg1;Arg2'.")
	flag.BoolVar(&showVersion, "V", false, "Version del cli.")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Uso: gow -f <archivo>\n")
		fmt.Fprintf(os.Stderr, "Ejemplo: gow -f ejercicio.go -t 10\n\n")
		fmt.Fprintln(os.Stderr, "Flags disponibles:")
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	if showVersion {
		fmt.Printf("Version %s\n", version)
		return
	}

	cfg, configPathResolved := loadOrDefaultConfiguration()

	if addRuleRaw != "" {
		handleAddRule(cfg, configPathResolved)
		return
	}

	if targetFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	cfg.DefaultTimeout = time.Duration(timeout) * time.Second
	finalDelay := time.Duration(delay) * time.Millisecond

	commander := runner.NewOSCommander()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := watcher.NewWatcher(commander, cfg, targetFile, verbose, finalDelay)

	fmt.Printf(
		"[INIT] gow iniciado. Vigilando: %s (Timeout: %v, Delay: %v).\n",
		targetFile, cfg.DefaultTimeout, finalDelay,
	)

	if err := w.Run(ctx); err != nil && err != context.Canceled {
		exitErr("Error fatal del Watcher: %v", err)
	}
}

func resolverConfigPath() (string, error) {
	if configPath != "" {
		return configPath, nil
	}

	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}

	binDir := filepath.Dir(exePath)
	return filepath.Join(binDir, ConfigFileName), nil
}

func LoadConfiguration() (*config.RunnerConfig, string, error) {
	path, err := resolverConfigPath()
	if err != nil {
		return nil, "", err
	}

	cfg, err := config.LoadConfig(path)
	return cfg, path, err
}

func loadOrDefaultConfiguration() (*config.RunnerConfig, string) {
	cfg, path, err := LoadConfiguration()
	if err != nil || cfg == nil {
		if !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "[WARN] Config inválida, usando valores por defecto: %v\n", err)
		}
		cfg = &config.RunnerConfig{
			DefaultTimeout: 5 * time.Microsecond,
			Rules:          []config.Rule{},
		}
	}

	return cfg, path
}

func exitErr(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "[ERR ] "+format+"\n", args...)
	os.Exit(1)
}

func handleAddRule(cfg *config.RunnerConfig, path string) {
	parts := strings.Split(addRuleRaw, ";")

	if len(parts) < 3 {
		exitErr("Formato de -a inválido. Use: '.ext;Nombre;Comando;Args...'")
	}
	if !strings.HasPrefix(parts[0], ".") {
		exitErr("La extensión debe empezar con '.'")
	}

	newRule := config.Rule{
		Extension:        parts[0],
		Name:             parts[1],
		ExecutionCommand: parts[2],
		ExecutionArgs:    parts[3:],
	}

	for i, r := range cfg.Rules {
		if r.Extension == newRule.Extension {
			cfg.Rules[i] = newRule
			goto save
		}
	}
	cfg.Rules = append(cfg.Rules, newRule)

save:
	if err := config.SaveConfig(path, cfg); err != nil {
		exitErr("No se pudo guardar la configuración: %v", err)
	}

	fmt.Printf("[ OK ] Lenguaje %s (%s) agregado/actualizado en %s\n",
		newRule.Name, newRule.Extension, path)
}
