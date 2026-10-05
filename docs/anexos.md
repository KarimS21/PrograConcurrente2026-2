# Anexos y guía de sustentación

Estos anexos acompañan el [informe final](informe.md). Se distinguen los archivos disponibles de las evidencias que todavía deben aportar los integrantes.

## Anexo A. Verificación formal

| Archivo | Qué demuestra |
|---|---|
| [worker_pool.pml](../promela/worker_pool.pml) | Abstracción del Worker Pool con cinco trabajos y tres workers. |
| [verify_spin.sh](../scripts/verify_spin.sh) | Comandos para generar y ejecutar los verificadores. |
| [spin-environment.txt](../results/spin-environment.txt) | Versiones de Spin y GCC y hash SHA-256 del modelo verificado. |
| [spin-safety.txt](../results/spin-safety.txt) | Búsqueda de deadlocks y violaciones de aserciones con `errors: 0`. |
| [spin-ltl.txt](../results/spin-ltl.txt) | Comprobación de finalización con `errors: 0`. |

En la demostración se debe mostrar la capacidad de los canales, el cierre coordinado, `assert(writers == 1)` y la propiedad LTL. Explicar que la exclusión de escritura se obtiene mediante un consumidor único y que el cálculo numérico está abstraído.

Para repetir la comprobación desde PowerShell, usando una distribución WSL con GCC:

```powershell
wsl -d Ubuntu -- sh /mnt/d/Concurrentes/PrograConcurrente2026-2/scripts/verify_spin.sh
```

La ruta corresponde a esta copia del proyecto; debe ajustarse si se clona en otra ubicación. El script conserva las salidas y comprueba que ambas búsquedas terminen sin errores ni límites de búsqueda excedidos.

## Anexo B. Trabajo con IA

**Herramienta seleccionada: GitHub Copilot Chat en VS Code.**

Se incluye el [prompt estructurado](ai-code-review-prompt.md) y el [análisis de GAPs](gap-analysis.md). Para acreditar el uso de la herramienta, falta guardar la captura o exportación de una interacción real.

La evidencia debe permitir reconocer:

1. La herramienta utilizada y el prompt enviado.
2. Los archivos del proyecto usados como contexto.
3. La respuesta recibida y sus recomendaciones.
4. La revisión del alumno: qué hallazgos aceptó, cuáles descartó y con qué evidencia.

Como ejemplos para contrastar con la respuesta, se pueden usar la falta de validación de valores no finitos, la dependencia de las pruebas respecto de un CSV no versionado y el coste de coordinación del Worker Pool. La captura debe corresponder a la conversación realizada; el informe no sustituye esa evidencia.

## Anexo C. Mediciones y pruebas

| Evidencia | Ubicación y alcance |
|---|---|
| Tiempos de inferencia | [inference_runs.csv](../results/inference_runs.csv): 11 mediciones por modalidad. |
| Recursos del proceso | [resource_runs.csv](../results/resource_runs.csv): máximos observados por configuración. |
| Gráfico | [inference_speedup.png](../results/inference_speedup.png): comparación de tiempos y speedup. |
| Prueba de igualdad | `go-code/inference_test.go`, `TestWorkerPoolMatchesSequentialInference`. Existe en el código; ejecución pendiente. |
| Pruebas de CSV | `go-code/dataset_test.go`. Requieren generar el CSV procesado. |

Las estadísticas se recalcularon a partir de los archivos existentes. No se repitieron las mediciones ni se ejecutaron las pruebas Go, porque Windows bloqueó el ejecutable. Para completar la evidencia, guardar las salidas de `go test`, `go vet` y `go test -race` cuando se ejecuten en un entorno compatible.

## Anexo D. Historial y participación

El [registro de Git](../results/git-history.txt) conserva el historial, autores, ramas y remoto consultados. Para mostrarlo durante la sustentación:

```powershell
git log --all --graph --decorate --oneline
git shortlog -sne --all
git branch -a
```

La evidencia actual identifica aportes de Karim y Juan Diego. No aparece un commit atribuible a Tomás. El remoto solo publica `main`; los merges existentes no acreditan por sí solos un flujo Gitflow completo. Una rama de funcionalidad puede borrarse después de integrarse, por lo que también conviene conservar los enlaces a sus PR.

### Propuesta para las siguientes contribuciones

Esta tabla es una propuesta de distribución de mejoras, no un registro de tareas ya realizadas:

| Integrante | Aporte propuesto | Rama sugerida |
|---|---|---|
| Karim | Mejorar validación del lector CSV. | `feature/validacion-csv` |
| Juan Diego | Añadir pruebas de límites del Worker Pool. | `feature/pruebas-concurrencia` |
| Tomás | Separar mediciones de tiempo y recursos. | `feature/mediciones` |

Para evidenciar Gitflow con aportes reales:

1. Crear y publicar `develop` desde la versión de `main` acordada por el equipo, conservando primero los cambios locales pendientes.
2. Cada integrante crea su rama `feature/...` desde `develop`, realiza su aporte y registra commits con su identidad habitual.
3. Abrir PR de cada funcionalidad hacia `develop`, revisar e integrar los cambios.
4. Preparar una rama `release/...` desde `develop` para la entrega; integrar la versión aprobada en `main` y devolver los ajustes a `develop`.
5. Conservar los enlaces a PR, hashes y autores, y mostrar el gráfico final del historial.

Después de realizar esas contribuciones, completar esta tabla con datos verificables:

| Integrante | Commit y aporte | PR hacia `develop` | Evidencia |
|---|---|---|---|
| Karim | Pendiente de la siguiente contribución. | Pendiente. | Añadir enlace o captura. |
| Juan Diego | Pendiente de la siguiente contribución. | Pendiente. | Añadir enlace o captura. |
| Tomás | Pendiente de acreditar su aporte. | Pendiente. | Añadir enlace o captura. |

Crear nombres de ramas o cambiar autores de commits anteriores no demuestra participación. El criterio requiere trabajo real de todos los integrantes y su integración visible.

## Anexo E. Guía de sustentación

### Orden de exposición sugerido

| Tiempo aproximado | Contenido | Evidencia que se muestra |
|---|---|---|
| 1 minuto | Problema, dataset y red neuronal. | Introducción del informe y arquitectura 9–16–8–1. |
| 2 minutos | Inferencia secuencial y Worker Pool. | `inference.go`: productor, workers, consumidor y cierre. |
| 2 minutos | Deadlocks, exclusión de escritura y finalización. | Modelo y salidas de Spin. |
| 1 minuto | Resultados de rendimiento. | Tabla y gráfico de speedup. |
| 1–2 minutos | Análisis con IA y tres GAPs prioritarios. | Prompt, interacción real e informe de GAPs. |
| 1–2 minutos | Conclusiones, recomendaciones y participación. | Conclusiones, historial y PR de los integrantes. |

El equipo puede repartirse los bloques según su conocimiento del trabajo. Todos deben poder explicar el diseño y los resultados, además de su aporte individual.

### Preguntas que conviene preparar

| Pregunta | Idea central de la respuesta |
|---|---|
| ¿Por qué no usan un mutex en las predicciones? | Un consumidor escribe el arreglo final y los workers leen pesos que ya no se modifican. |
| ¿Cómo mantienen el orden? | Cada resultado conserva el índice de su muestra. |
| ¿Quién cierra los canales? | El productor cierra trabajos y el coordinador cierra resultados después de esperar a los workers. |
| ¿Qué significa `errors: 0`? | No se encontró una violación de las propiedades buscadas en el espacio de estados del modelo. |
| ¿Spin verifica toda la aplicación Go? | Verifica la abstracción Promela; faltan comprobaciones de la implementación y del cálculo numérico. |
| ¿Por qué cuatro workers no superaron al secuencial? | La tarea es pequeña y la coordinación puede costar más que el ahorro; hace falta un perfil para confirmar la causa. |
| ¿Qué mejorarían primero? | Validación de datos y pruebas reproducibles; después medición y tareas por lotes. |
| ¿Se acredita la participación de todos? | Aún falta evidencia de Tomás y del flujo Gitflow completo. Deben mostrarse aportes y PR reales. |

### Evidencia de la exposición

**Enlace al video o grabación:** pendiente de incorporar.

Antes de entregar, comprobar que el enlace sea accesible para el docente y que la exposición permita reconocer a los integrantes y sus aportes. Los resultados mostrados deben coincidir con los anexos entregados.
