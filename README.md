# Predicción de Tarifas de Taxi en NYC mediante Redes Neuronales y Concurrencia en Go

[![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)](#)
[![Python](https://img.shields.io/badge/Python-3776AB?style=for-the-badge&logo=python&logoColor=white)](#)
[![Jupyter](https://img.shields.io/badge/Jupyter-F37626?style=for-the-badge&logo=Jupyter&logoColor=white)](#)
[![Spin/Promela](https://img.shields.io/badge/Spin%2FPromela-Formal%20Verification-6c3483?style=for-the-badge)](#)
[![UPC](https://img.shields.io/badge/UPC-E3000F?style=for-the-badge&logo=upc&logoColor=white)](#)

Repositorio del trabajo final del curso **1ACC0065 — Programación Concurrente y Distribuida** (Ciencias de la Computación, UPC 2026-2).

El proyecto implementa desde cero una **Red Neuronal Feedforward** en Go (sin dependencias de terceros) para predecir el costo de viajes de taxi en NYC. El objetivo central es comparar el rendimiento de la inferencia **secuencial** frente a una versión **concurrente** basada en el patrón **Worker Pool** (channels + `sync.WaitGroup`), y validar formalmente la ausencia de condiciones de carrera mediante un modelo en **Promela/Spin**.

Se alinea con el **ODS 11 (Ciudades y Comunidades Sostenibles)** de la ONU al analizar y optimizar patrones de transporte urbano masivo.

---

## Equipo de Desarrollo

| Código | Nombres y Apellidos |
| :--- | :--- |
| U202221147 | Juan Diego Adrián Sánchez Sánchez |
| U201816862 | Karim Wagner Samanamud Mosquera |
| U201916314 | Tomas Alonso Pastor Salazar |

---

## Arquitectura del Proyecto

```
PrograConcurrente2026-2/
├── TrabajoParcial.ipynb          # Notebook de limpieza: parquet → CSV procesado
├── go.mod                        # Módulo Go "tp-concurrente" (Go 1.21, sin deps externas)
│
├── raw/                          # Dataset fuente (no modificar)
│   └── yellow_tripdata_2026-01.parquet  # 1.48 M filas, formato Parquet
│
├── processed/                    # ⚠️ EXCLUIDO del repo (.gitignore) — debe generarse localmente
│   └── yellow_tripdata_2026-01.csv      # 1,002,542 filas limpias (se genera con el Notebook)
│
├── go-code/                      # Paquete "gocode": librería core del modelo
│   ├── redneuronal.go            # Red neuronal (9→16→8→1, ReLU, backprop full-batch, Standardizer)
│   ├── inference.go              # PredictSequential / PredictWorkerPool / Evaluate (MAE, MSE)
│   ├── dataset.go                # Carga y parseo del CSV a []Sample
│   └── *_test.go                 # Tests unitarios y de integración
│
├── cmd/
│   ├── benchmark/main.go         # Benchmark rápido: imprime speedup en la terminal
│   └── experiment/main.go        # Experimento completo: N runs × M configs de workers → CSV
│
├── promela/
│   └── worker_pool.pml           # Modelo formal Promela: 3 workers, 5 jobs, LTL de completitud
│
├── scripts/
│   ├── measure_resources.ps1     # PowerShell: lanza experimentos y mide CPU/RAM del SO
│   └── plot_experiment.py        # Python: genera gráfico de speedup desde inference_runs.csv
│
└── results/                      # Artefactos de ejecución (generados, no commitear binarios)
    ├── inference_runs.csv         # Métricas de todas las runs (entrada para plot_experiment.py)
    ├── resource_runs.csv          # CPU/RAM por configuración de workers
    └── inference_speedup.png      # Gráfico final de speedup
```

---

## Prerrequisitos

| Herramienta | Versión mínima | Uso |
|---|---|---|
| Go | 1.21 | Compilar y ejecutar el modelo |
| Python | 3.9 | Limpiar el dataset y graficar resultados |
| `pandas`, `pyarrow` | cualquiera reciente | Leer el `.parquet` en el Notebook |
| `matplotlib` | cualquiera reciente | Generar el gráfico de speedup |
| Docker *(opcional)* | cualquiera | Verificación con Spin sin instalar localmente |
| Spin *(opcional)* | 6.x | Verificación formal del modelo Promela |

```powershell
# Instalar dependencias de Python
pip install pandas pyarrow matplotlib
```

---

## Guía de Ejecución (Paso a Paso)

### Paso 1 — Generar el Dataset Procesado

> ⚠️ El archivo `processed/yellow_tripdata_2026-01.csv` **no está en el repositorio** (excluido por `.gitignore`
> porque supera los límites de Git). Debe generarse localmente antes de cualquier ejecución en Go.

Ejecutar el Jupyter Notebook que lee `raw/yellow_tripdata_2026-01.parquet`, aplica los filtros de
limpieza y escribe el CSV limpio:

```powershell
# Desde la raíz del repositorio
jupyter nbconvert --to notebook --execute TrabajoParcial.ipynb --output TrabajoParcial_executed.ipynb
```

O simplemente abrir `TrabajoParcial.ipynb` en Jupyter / VS Code y ejecutar todas las celdas.

**Resultado:** `processed/yellow_tripdata_2026-01.csv` (~22 MB, 1,002,542 registros)

---

### Paso 2 — Tests Unitarios de Go

```powershell
# Ejecutar todos los tests con el race detector habilitado
go test -v -race -timeout 120s ./go-code/...
```

> ℹ️ Los tests `TestLoadSamplesFromProcessedCSV` y `TestProcessedCSVInferenceFlow` requieren
> que el CSV del Paso 1 ya exista en `processed/`.

**Salida:** Imprime en consola (`PASS`/`FAIL`). No genera archivos.

---

### Paso 3 — Benchmark Rápido (speedup de una sola pasada)

```powershell
go run ./cmd/benchmark/ `
  -csv           processed/yellow_tripdata_2026-01.csv `
  -max-rows      100000 `
  -train-rows    10000 `
  -epochs        20 `
  -learning-rate 0.001 `
  -workers       8
```

**Salida (solo consola):**
```
rows loaded: 100000
training rows: 10000
evaluation rows: 90000
workers: 8
load: 1.23s
training: 456ms
sequential inference: 27.4ms (MAE 5.1234, MSE 48.5678)
worker pool inference:  8.2ms (MAE 5.1234, MSE 48.5678)
speedup: 3.34x
```

---

### Paso 4 — Experimento Completo (múltiples workers × múltiples runs)

Este comando ejecuta `N` runs por cada configuración de workers, calcula la media recortada (*trimmed mean*)
y escribe un CSV con todas las métricas detalladas.

```powershell
go run ./cmd/experiment/ `
  -csv           processed/yellow_tripdata_2026-01.csv `
  -max-rows      100000 `
  -train-rows    10000 `
  -epochs        1 `
  -learning-rate 0.001 `
  -workers       "1,2,4,8" `
  -runs          11 `
  -trim          0.2 `
  -output        results/inference_runs.csv
```

**Salida consola:** Tabla de trimmed mean por modo/workers.  
**Archivo generado:** `results/inference_runs.csv`

---

### Paso 5 — Generar Gráfico de Speedup

```powershell
# Lee results/inference_runs.csv y escribe results/inference_speedup.png
python scripts/plot_experiment.py
```

**Archivo generado:** `results/inference_speedup.png`

---

### Paso 6 — Medir Recursos del Sistema (CPU/RAM) con PowerShell

El script compila el binario, lanza una ejecución separada por cada configuración de workers y
mide el consumo de CPU y RAM del proceso en tiempo real (muestreo cada 100 ms).

```powershell
.\scripts\measure_resources.ps1 `
  -MaxRows   100000 `
  -TrainRows 10000 `
  -Runs      11 `
  -Workers   "1,2,4,8" `
  -Output    "results/resource_runs.csv"
```

**Archivos generados:**

| Archivo | Contenido |
|---|---|
| `results/resource_runs.csv` | CPU%, RAM (Working Set, Heap Go, Sys Go) por config |
| `results/inference-{1,2,4,8}.csv` | Métricas de inferencia detalladas por config de workers |
| `results/experiment-{1,2,4,8}.out` | Stdout (trimmed mean) de cada ejecución |
| `results/experiment-{1,2,4,8}.err` | Stderr de cada ejecución |

---

### Paso 7 — Verificación Formal con Spin/Promela

El modelo `promela/worker_pool.pml` verifica formalmente que en un sistema con `WORKERS=3` y
`JOBS=5`, todos los trabajos son procesados exactamente una vez (propiedad LTL `all_jobs_complete`).

**Con Docker (no requiere instalar Spin):**

```powershell
# 1. Generar el verificador (pan.c)
docker run --rm -v "${PWD}/promela:/model" spinroot/spin spin -a /model/worker_pool.pml

# 2. Compilar pan.c
docker run --rm -v "${PWD}/promela:/model" --entrypoint gcc spinroot/spin -o /model/pan /model/pan.c

# 3. Ejecutar la verificación (búsqueda exhaustiva del espacio de estados)
docker run --rm -v "${PWD}/promela:/model" --entrypoint /model/pan spinroot/spin -a
```

**Con Spin instalado localmente:**

```powershell
cd promela
spin -a worker_pool.pml
gcc -o pan pan.c
./pan -a
```

**Salida esperada:** `State-vector ... No errors found.`

---

## Diseño Concurrente — Worker Pool

La función [`PredictWorkerPool`](go-code/inference.go) implementa el patrón canónico de Go:

```
Dispatcher Goroutine
    └─ envía jobs al canal (chan job)
           ↓
    ┌──────────────────┐
    │  Worker 1        │
    │  Worker 2   ──→  │ (chan result, buffered=len(samples))
    │  Worker N        │
    └──────────────────┘
           ↓
    Main Goroutine: range results → predictions[index]
```

**Garantías de corrección:**
- La `NeuralNetwork` es tratada como **memoria de solo lectura** durante la inferencia (`Predict` no modifica el estado de la red). → **Cero `sync.Mutex`** necesario.
- Cada índice del slice `predictions` es escrito por **exactamente una goroutina** (el dispatcher emite cada `job.index` una sola vez). → No hay data races en la escritura.
- El `defer workers.Done()` en cada worker garantiza que el `WaitGroup` se decrementa incluso ante un panic.
- El canal `results` es **buffered** (`cap = len(samples)`), eliminando bloqueos en la escritura de los workers.

---

## Resumen de Salidas por Comando

| Comando | Salida Consola | Archivos Generados |
|---|---|---|
| `jupyter nbconvert --execute ...` | Progreso | `processed/yellow_tripdata_2026-01.csv` |
| `go test -v -race ./go-code/...` | PASS/FAIL | — |
| `go run ./cmd/benchmark/` | Speedup y métricas | — |
| `go run ./cmd/experiment/` | Trimmed mean por modo | `results/inference_runs.csv` |
| `python scripts/plot_experiment.py` | Ruta de la imagen | `results/inference_speedup.png` |
| `.\scripts\measure_resources.ps1` | Confirmación | `results/resource_runs.csv`, `results/inference-*.csv`, `results/experiment-*.{out,err}` |
| Verificación Spin | `No errors found` | `promela/pan.c`, `promela/pan` |

---

## Fases del Proyecto (Entregables)

- [x] **PC1:** Investigación de Literatura (Estado del Arte), selección y limpieza del dataset (>1 M filas).
- [x] **PC2 (entrega actual):**
  - Implementación de la Red Neuronal en Go (sin librerías externas).
  - Implementación del patrón Worker Pool para inferencia concurrente.
  - Experimento de Speedup con múltiples configuraciones de workers (1, 2, 4, 8).
  - Medición de recursos del SO (CPU%, RAM) via PowerShell.
  - Validación formal de sincronización en Promela/Spin.
  - Gráficos de análisis de rendimiento.