# Prompt estructurado para análisis asistido por IA

## Modelo seleccionado

**GitHub Copilot Chat en VS Code**, utilizando un modelo de razonamiento para revisión de código. El análisis debe ejecutarse sobre el commit y el árbol de trabajo que se indiquen, no sobre una copia supuesta del proyecto.

## Prompt

```text
Actúa como un revisor senior de Go, concurrencia, seguridad de software y sistemas de Machine Learning.

Repositorio objetivo:
https://github.com/KarimS21/PrograConcurrente2026-2.git

Contexto:
El proyecto implementa desde cero una red neuronal para regresión de fare_amount sobre datos de taxis. La inferencia secuencial se compara con un Worker Pool en Go. Existe un modelo Promela para abstraer la sincronización.

Alcance de archivos:
- go-code/redneuronal.go
- go-code/inference.go
- go-code/dataset.go
- go-code/*_test.go
- cmd/benchmark/main.go
- cmd/experiment/main.go
- scripts/*.ps1
- scripts/*.py
- promela/worker_pool.pml
- go.mod

Objetivos:
1. Detectar defectos de corrección, deadlocks, condiciones de carrera, pérdida o desorden de resultados y problemas de cierre de canales.
2. Revisar calidad Go: APIs, errores, validación de entradas, mantenibilidad, duplicación, complejidad y pruebas.
3. Revisar seguridad: lectura de archivos, uso de rutas recibidas por flags, datos sensibles, dependencias, ejecución de scripts y riesgo de denegación de servicio por consumo de memoria.
4. Revisar el diseño ML: normalización, variables categóricas, separación entrenamiento/evaluación, estabilidad numérica, reproducibilidad y fuga de información.
5. Revisar rendimiento: granularidad de tareas, overhead del Worker Pool, medición, calentamiento, media recortada, CPU, memoria y escalabilidad.
6. Comparar el modelo Promela con la implementación Go y señalar cualquier diferencia de abstracción.

Reglas:
- No inventes resultados de ejecución. Distingue evidencia observada, inferencia y recomendación.
- Para cada hallazgo indica severidad: Crítica, Alta, Media, Baja o Informativa.
- Incluye archivo, símbolo o línea aproximada, impacto, evidencia y remediación concreta.
- Reconoce también controles que sí funcionan.
- No propongas dependencias nuevas si la biblioteca estándar resuelve el problema razonablemente.
- No modifiques código durante el análisis.

Formato de salida:
1. Resumen ejecutivo.
2. Alcance, commit y limitaciones.
3. Hallazgos priorizados en una tabla: ID, severidad, ubicación, problema, impacto y recomendación.
4. Análisis específico de concurrencia y sincronización.
5. Análisis específico de seguridad.
6. Análisis específico del pipeline ML y calidad de datos.
7. Gaps de pruebas y observabilidad.
8. Plan de remediación ordenado por prioridad.
9. Evidencia de comandos ejecutados y resultados.
```

## Evidencia usada para este análisis

- Remoto: `https://github.com/KarimS21/PrograConcurrente2026-2.git`
- Rama: `main`
- Commit remoto base: `0cd16be`; incluir también el árbol de trabajo local para analizar los cambios actuales de Promela y documentación.
- Verificación Go: `go test -count=1 ./...` y `go vet ./...`
- Verificación Promela: Spin 6.4.9 ejecutado dentro de Docker.
