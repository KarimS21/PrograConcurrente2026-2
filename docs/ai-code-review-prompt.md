# Prompt estructurado para análisis asistido por IA

## Modelo seleccionado

**GitHub Copilot Chat en VS Code**

## Objetivo

Analizar el proyecto de red neuronal y Worker Pool para identificar GAPs: diferencias entre la implementación actual y las prácticas necesarias para mejorar su calidad, seguridad y concurrencia.

## Prompt

Copiar el siguiente texto en el chat con el repositorio abierto y adjuntar los archivos indicados como contexto:

```text
Actúa como revisor de un proyecto universitario de Programación Concurrente en Go.

Repositorio: https://github.com/KarimS21/PrograConcurrente2026-2
Commit base: f9a7181, rama main. Considera también el modelo Promela y la
documentación del árbol de trabajo local, aclarando cuáles no están publicados.

El proyecto implementa una red neuronal desde cero para predecir fare_amount
de viajes de taxi. El entrenamiento es secuencial. La inferencia compara un
recorrido secuencial con un Worker Pool de goroutines y canales.

Revisa estos archivos:
- go-code/redneuronal.go, inference.go, dataset.go y sus pruebas.
- cmd/benchmark/main.go y cmd/experiment/main.go.
- scripts/measure_resources.ps1 y scripts/plot_experiment.py.
- promela/worker_pool.pml y go.mod.

Analiza:
1. Calidad: claridad, validación de entradas, manejo de errores y pruebas.
2. Seguridad: datos inválidos, consumo de memoria y manejo de archivos.
3. Concurrencia: Worker Pool, cierre de canales, WaitGroup, condiciones de
   carrera, deadlocks y conservación del orden de resultados.
4. Modelo predictivo: normalización, variables categóricas y evaluación.
5. Rendimiento: coste de coordinación, speedup y límites de las mediciones.
6. Relación entre el modelo Promela y el código Go.

Entrega un informe en Markdown con:
- Un resumen breve del funcionamiento.
- Una tabla de GAPs con ID, severidad, archivo o función, evidencia,
  consecuencia y recomendación concreta.
- Los aspectos correctamente implementados.
- Las limitaciones de la revisión y las mejoras prioritarias.

Usa lenguaje claro, con profundidad adecuada para un trabajo universitario.
Sustenta cada hallazgo en el código. No inventes fallos ni resultados de
comandos. Diferencia lo observado de las hipótesis. No modifiques el código.
```

## Cómo evidenciar el trabajo con IA

1. Enviar el prompt con los archivos del proyecto como contexto.
2. Guardar una captura o exportación que muestre la herramienta, el prompt y la respuesta.
3. Contrastar los hallazgos con el código y explicar cuáles se aceptan y por qué.
4. Adjuntar esa evidencia al informe final. El análisis técnico preparado está en [gap-analysis.md](gap-analysis.md).

La herramienta seleccionada es la indicada arriba. La captura o exportación de su ejecución todavía debe incorporarse; este archivo documenta el prompt y no acredita por sí solo una conversación realizada.

## Documentos relacionados

- [Informe final](informe.md): resultados, explicación de Spin y conclusiones.
- [Informe de GAPs](gap-analysis.md): revisión técnica del proyecto.
- [Anexos y guía de sustentación](anexos.md): evidencias y puntos para la exposición.
