package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
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
)

func init() {
	flag.StringVar(&targetFile, "f", "", "El archivo a vigilar y ejecutar (ej. main.go).")
	flag.StringVar(&configPath, "c", "", "Ruta a un archivo de configuración personalizada.")
	flag.IntVar(&timeout, "t", 5, "Timeout de ejecución en segundos para evitar bucles infinitos.")
	flag.BoolVar(&verbose, "v", false, "Habilita la salida detallada (verbose) de logs internos.")
	flag.IntVar(&delay, "delay", 100, "Retraso de debounce en milisegundos después de la detección de un cambio.")
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

	if targetFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	cfg, err := LoadConfiguration()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERR ] Error al cargar la configuración: %v\n", err)
		os.Exit(1)
	}

	finalDelay := 100 * time.Millisecond

	flag.Visit(func(f *flag.Flag) {
		if f.Name == "t" {
			cfg.DefaultTimeout = time.Duration(timeout) * time.Second
		}
		if f.Name == "delay" {
			finalDelay = time.Duration(delay) * time.Millisecond
		}
	})

	commander := runner.NewOSCommander()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := watcher.NewWatcher(commander, cfg, targetFile, verbose, finalDelay)

	fmt.Printf("[INIT] gow iniciado. Vigilando: %s (Timeout: %v, Delay: %v).\n", targetFile, cfg.DefaultTimeout, finalDelay)

	if err := w.Run(ctx); err != nil {
		if err != context.Canceled {
			fmt.Fprintf(os.Stderr, "[ERR ] Error fatal del Watcher: %v\n", err)
		}
	}
}

func LoadConfiguration() (*config.RunnerConfig, error) {
	if configPath != "" {
		return config.LoadConfig(configPath)
	}

	localConfig := "gow.config.json"
	exePath, _ := os.Executable()
	binDir := filepath.Dir(exePath)
	defaultPath := filepath.Join(binDir, localConfig)

	return config.LoadConfig(defaultPath)
}
