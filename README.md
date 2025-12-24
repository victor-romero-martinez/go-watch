# go-watch 👀⚡

**go-watch** es un CLI escrito en Go que permite ejecutar y **recargar automáticamente un archivo cuando detecta cambios**, sin estar limitado a un lenguaje específico.

A diferencia de otros *watchers*, **go-watch no está atado a un runtime**: puede ejecutar **cualquier lenguaje** siempre que exista una **regla que indique cómo correrlo**.

---

## 🚀 Características

* 🔥 Hot reload de archivos individuales
* 🌍 Soporte para **cualquier lenguaje** mediante reglas
* 🧩 Configuración extensible vía JSON
* ⏱ Timeout para evitar loops infinitos
* 🧠 Debounce configurable para cambios de archivos
* 🧪 Modo verbose para debugging
* ➕ Agregar lenguajes en runtime desde CLI

---

## 📦 Instalación

```bash
go install https://github.com/victor-romero-martinez/go-watch@latest
```

O clonando el repositorio:

```bash
git clone https://github.com/victor-romero-martinez/go-watch.git
cd go-watch
go build -o gow
```

---

## 🛠 Uso básico

```bash
gow -f watchme.js
```

Cada vez que el archivo cambie, **go-watch lo volverá a ejecutar automáticamente**.

---

## 🚩 Flags disponibles

| Flag     | Descripción                                                   |
| -------- | ------------------------------------------------------------- |
| `-f`     | Archivo a vigilar y ejecutar                                  |
| `-c`     | Ruta a un archivo de configuración personalizada              |
| `-t`     | Timeout de ejecución en segundos (default: 5)                 |
| `-delay` | Retraso de debounce en ms después de un cambio (default: 100) |
| `-v`     | Habilita logs detallados                                      |
| `-a`     | Agrega un lenguaje dinámicamente                              |
| `-V`     | Muestra la versión del CLI                                    |

### Ejemplo

```bash
gow -f main.go -v -t 10
```

---

## ⚙️ Configuración

go-watch usa un archivo de configuración en formato JSON que define **cómo ejecutar cada lenguaje**.

### Ejemplo de configuración

```json
{
  "default_timeout_ms": 5000,
  "rules": [
    {
      "extension": ".go",
      "name": "Golang Runner",
      "execution_command": "go",
      "execution_args": ["run", "$FILE"],
      "needs_build": false
    },
    {
      "extension": ".js",
      "name": "Node.js Runner",
      "execution_command": "node",
      "execution_args": ["$FILE"],
      "needs_build": false
    },
    {
      "extension": ".rs",
      "name": "Rust Runner (Requires Build)",
      "execution_command": "/bin/sh",
      "execution_args": [
        "-c",
        "rustc $FILE -o /tmp/gow_bin && /tmp/gow_bin"
      ],
      "needs_build": true
    }
  ]
}
```

---

## 🧠 ¿Cómo funcionan las reglas?

Cada regla define:

* `extension`: extensión del archivo
* `execution_command`: comando principal
* `execution_args`: argumentos (puede usar `$FILE`)
* `needs_build`: indica si requiere compilación previa

Esto permite ejecutar **lenguajes compilados o interpretados** sin modificar el código del CLI.

---

## ➕ Agregar un lenguaje desde CLI

También podés agregar reglas dinámicamente:

```bash
gow -a ".py;Python Runner;python;$FILE"
```

Formato:

```
.ext;Nombre;Comando;Arg1;Arg2;...
```

---

## 🧪 Casos de uso

* Desarrollo rápido de scripts
* Probar archivos individuales
* Hot reload multi-lenguaje
* Prototipado y debugging

---

## 🗂️ Estructura del proyecto

```txt
./
├── LICENSE.txt
├── README.md
├── build.sh
├── go.mod
├── go.sum
├── main.go
├── gow.config.json
├── config/
│   ├── config.go
│   └── config_test.go
├── internal/
│   └── testutil/
│       └── mocks.go
├── runner/
│   ├── runner.go
│   └── runner_test.go
└── watcher/
    ├── watcher.go
    └── watcher_test.go
```