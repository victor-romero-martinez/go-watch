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
	mu            sync.Mutex
	fileHashes    map[string]string
	commander     runner.Commander
	config        *config.RunnerConfig
	targetFile    string
	debounceCh    chan struct{} // Canal para controlar el debounce
	verbose       bool
	delay         time.Duration
	cancelCurrent context.CancelFunc
}

func NewWatcher(commander runner.Commander, cfg *config.RunnerConfig, targetFile string, verbose bool, delay time.Duration) *Watcher {
	return &Watcher{
		fileHashes: make(map[string]string),
		commander:  commander,
		config:     cfg,
		targetFile: targetFile,
		debounceCh: make(chan struct{}, 1),
		verbose:    verbose,
		delay:      delay,
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

	h, _ := w.updateHash(w.targetFile)
	w.mu.Lock()
	w.fileHashes[w.targetFile] = h
	w.mu.Unlock()

	w.handleExecution()

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
	time.Sleep(w.delay)

	select {
	case <-w.debounceCh:
		// Se encontró otro evento, ignorarlo
	default:
		// No hay mas eventos pendientes, podemos ejecutar.
	}

	newHash, err := w.updateHash(w.targetFile)
	if err != nil {
		fmt.Printf("Error al calcular hash: %v\n", err)
		return
	}

	w.mu.Lock()
	oldHash := w.fileHashes[w.targetFile]
	isDifferent := newHash != oldHash

	if isDifferent {
		w.fileHashes[w.targetFile] = newHash
	}
	w.mu.Unlock()

	// 3. Ejecutar fuera del Lock para evitar Deadlock
	if isDifferent {
		fmt.Printf("[DONE] Archivo modificado. Ejecutando %s...\n", w.targetFile)
		w.handleExecution()
	} else {
		w.log("Archivo guardado, pero el contenido no cambió. Ignorando.")
	}
}

func (w *Watcher) updateHash(filepath string) (string, error) {
	return calculateFileHash(filepath)
}

func (w *Watcher) handleExecution() {
	// 1. CANCELAR PROCESO ANTERIOR (Si existe)
	w.mu.Lock()
	if w.cancelCurrent != nil {
		w.log("Cancelando ejecución anterior...")
		w.cancelCurrent()
	}

	// Crear un contexto con el timeout de la configuración
	ctx, cancel := context.WithTimeout(context.Background(), w.config.DefaultTimeout)
	w.cancelCurrent = cancel
	w.mu.Unlock()

	// 2. BUSCAR REGLA
	ext := filepath.Ext(w.targetFile)
	var rule config.Rule
	found := false
	for _, r := range w.config.Rules {
		if r.Extension == ext {
			rule = r
			found = true
			break
		}
	}

	if !found {
		fmt.Printf("[WARN] No se encontró regla para la extensión %s\n", ext)
		cancel()
		return
	}

	// 3. PREPARAR ARGUMENTOS
	args := make([]string, len(rule.ExecutionArgs))
	// Sustituir $FILE por el path real
	for i, arg := range rule.ExecutionArgs {
		if arg == "$FILE" {
			args[i] = w.targetFile
		} else {
			args[i] = arg
		}
	}

	// 4. EJECUTAR EN SEGUNDO PLANO (Goroutine)
	go func(c context.Context, fCancel context.CancelFunc, r config.Rule, a []string) {
		defer fCancel()

		// El Commander.Run debería imprimir directamente a os.Stdout
		err := w.commander.Run(c, r.ExecutionCommand, a)

		if err != nil {
			if c.Err() == context.DeadlineExceeded {
				fmt.Printf("[ERR ] Timeout alcanzado.\n")
			} else if c.Err() == context.Canceled {
				w.log("Proceso cancelado por nuevo cambio.")
			} else {
				fmt.Printf("[ERR ] Ejecución fallida: %v\n", err)
			}
		} else {
			if c.Err() == nil {
				fmt.Println("[ OK ] Ejecución completada.")
			}
		}
	}(ctx, cancel, rule, args)
}

func (w *Watcher) log(format string, v ...interface{}) {
	if w.verbose {
		fmt.Printf("[DEBUG] "+format+"\n", v...)
	}
}
