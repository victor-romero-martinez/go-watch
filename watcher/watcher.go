package watcher

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gow/config"
	"gow/runner"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	// Mutex para proteger el acceso concurrente al mapa de hashes
	mu         sync.Mutex
	fileHashes map[string]string
	commander  runner.Commander
	config     *config.RunnerConfig
	targetFile string
	// Canal para controlar el debounce
	debounceCh chan struct{}
	verbose    bool
}

func NewWatcher(commander runner.Commander, cfg *config.RunnerConfig, targetFile string, verbose bool) *Watcher {
	return &Watcher{
		fileHashes: make(map[string]string),
		commander:  commander,
		config:     cfg,
		targetFile: targetFile,
		debounceCh: make(chan struct{}, 1),
		verbose:    verbose,
	}
}

func calculateFileHash(filepath string) (string, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()

	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func (w *Watcher) Run(ctx context.Context) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("Error al inicializar fsnotify: %w", err)
	}
	defer watcher.Close()

	// Añadimos el directorio padre para capturar eventos de Rename/Create/Remove
	dir := filepath.Dir(w.targetFile)
	if err := watcher.Add(dir); err != nil {
		return fmt.Errorf("Error al vigilar el directorio %s: %w", dir, err)
	}

	w.updateHash(w.targetFile)

	go w.handleExecution()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			if filepath.Clean(event.Name) == filepath.Clean(w.targetFile) {
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					w.debounceCh <- struct{}{}
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Printf("Error de fsnotify: %c\n", err)
		case <-w.debounceCh:
			go w.handleDebounce()
		}
	}
}

func (w *Watcher) handleDebounce() {
	w.log("Iniciando verificación de cambios para %s", w.targetFile)
	time.Sleep(100 * time.Millisecond)

	select {
	case <-w.debounceCh:
		// Se encontró otro evento, ignorarlo
	default:
		// No hay mas eventos pendientes, podemos ejecutar.
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	newHash, err := w.updateHash(w.targetFile)
	if err != nil {
		fmt.Printf("Error al calcular hash: %v\n", err)
		return
	}

	oldHash, exists := w.fileHashes[w.targetFile]

	if !exists || newHash != oldHash {
		fmt.Printf("[DONE] Archivo modificado. Ejecutando %s...\n", w.targetFile)
		w.fileHashes[w.targetFile] = newHash
		w.log("Hash cambiado: %s", newHash)
		w.handleExecution()
	} else {
		w.log("[INFO] Archivo guardado, pero el contenido no cambió. Ignorando.")
	}
}

func (w *Watcher) updateHash(filepath string) (string, error) {
	return calculateFileHash(filepath)
}

func (w *Watcher) handleExecution() {
	ext := filepath.Ext(w.targetFile)
	var rule *config.Rule

	for i, r := range w.config.Rules {
		if r.Extension == ext {
			rule = &w.config.Rules[i]
			break
		}
	}

	if rule == nil {
		fmt.Printf("[WARN] No se encontró regla para la extensión %s\n", ext)
		return
	}

	cmd := rule.ExecutionCommand
	args := make([]string, len(rule.ExecutionArgs))

	// Sustituir $FILE por el path real
	for i, arg := range rule.ExecutionArgs {
		if arg == "$FILE" {
			args[i] = w.targetFile
		} else {
			args[i] = arg
		}
	}

	// Crear un contexto con el timeout de la configuración
	ctx, cancel := context.WithTimeout(context.Background(), w.config.DefaultTimeout)
	defer cancel()

	err := w.commander.Run(ctx, cmd, args)

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Printf("[ERR ] EJECUCIÓN FALLIDA: Timeout de %v alcanzado.\n", w.config.DefaultTimeout)
		} else {
			fmt.Printf("[ERR ] EJECUCIÓN FALLIDA: %v\n", err)
		}
	} else {
		fmt.Println("[ OK ] Ejecución completada.")
	}
}

func (w *Watcher) log(format string, v ...interface{}) {
	if w.verbose {
		fmt.Printf("[DEBUG] "+format+"\n", v...)
	}
}
