# Informe de GAPs técnicos

## 1. Objetivo y alcance

Este informe identifica aspectos que pueden mejorarse en la calidad, seguridad y concurrencia del proyecto [PrograConcurrente2026-2](https://github.com/KarimS21/PrograConcurrente2026-2). Un GAP es una diferencia entre el comportamiento actual y el comportamiento esperado o una práctica recomendable.

Se revisaron la red neuronal, el lector CSV, la inferencia, las pruebas, los comandos de experimentación y los scripts. La base es el commit `f9a7181` de `main`, confirmado en el remoto durante la revisión del 4 de octubre de 2026. El modelo Promela y la documentación incluyen cambios locales posteriores.

El [prompt](ai-code-review-prompt.md) define el análisis solicitado a **GitHub Copilot Chat en VS Code**. Los hallazgos siguientes se sustentan en la lectura del código y en las evidencias del repositorio. La captura o exportación de la interacción con la herramienta seleccionada sigue pendiente.

## 2. Resumen del análisis

El proyecto separa correctamente el entrenamiento y la inferencia. Los workers leen la misma red y devuelven resultados con su índice original. Un consumidor guarda las predicciones, por lo que el orden de finalización de los workers no cambia su correspondencia con las muestras.

Las mejoras principales son validar los datos de entrada, hacer las pruebas independientes del dataset completo y medir con mayor precisión el coste de la concurrencia. La seguridad se analiza para una aplicación local que procesa CSV; los problemas detectados no deben presentarse como vulnerabilidades de un servicio web.

## 3. GAPs identificados

Se usa severidad **media** cuando el problema puede afectar resultados, estabilidad o reproducibilidad; **baja** para mejoras de mantenimiento o eficiencia; e **informativa** para límites de la evidencia.

| ID | Área y severidad | Ubicación y evidencia | Consecuencia | Recomendación |
|---|---|---|---|---|
| GAP-01 | Datos / Media | `dataset.go`, `parseCSVFloat`: convierte con `strconv.ParseFloat`, sin comprobar finitud. Tampoco valida que la fecha final sea posterior a la inicial. | `NaN`, infinitos o duraciones negativas pueden contaminar predicciones y métricas. | Rechazar valores con `math.IsNaN` y `math.IsInf`; comprobar fechas y rangos coherentes con el dataset. |
| GAP-02 | Calidad / Baja | `dataset.go`, `requiredCSVColumns`: exige `trip_time_min`, pero la duración se recalcula en `TaxiTrip.Features`. | Una columna inconsistente pasa sin advertencia; el contrato de entrada resulta confuso. | Definir una fuente de duración y documentarla, o comparar ambos valores con una tolerancia. |
| GAP-03 | API / Media | `inference.go`: `PredictWorkerPool` rechaza una red nula, pero `PredictSequential` no la valida. | Con muestras no vacías, la ruta secuencial puede causar un panic. | Unificar el contrato de las funciones y probar el caso de red nula. |
| GAP-04 | Seguridad y recursos / Media | `LoadSamplesFromCSV` carga todas las filas si `maxRows <= 0`; `PredictWorkerPool` crea un canal de resultados con capacidad `len(samples)`. | El consumo de memoria crece con la cantidad de registros, especialmente con archivos grandes no confiables. | Establecer límites para entradas externas y evaluar procesamiento por lotes y un buffer acotado. |
| GAP-05 | Concurrencia / Media | `NeuralNetwork.Train` modifica los pesos; `Predict` los lee sin bloqueo. Los comandos actuales entrenan antes de inferir. | Si se usan simultáneamente sobre la misma red, aparecen accesos conflictivos a los pesos. | Documentar la precondición de terminar el entrenamiento antes de inferir; si se requiere simultaneidad, usar una copia estable de los pesos o sincronización. |
| GAP-06 | Pruebas / Media | `dataset_test.go` usa `../processed/yellow_tripdata_2026-01.csv`, excluido de Git y ausente en esta copia. | Las pruebas no se reproducen directamente después de clonar. | Usar un CSV pequeño y controlado en `testdata` o generado con `t.TempDir`; separar las pruebas de integración. |
| GAP-07 | Pruebas / Baja | `inference_test.go` prueba igualdad y cero workers; faltan casos de muestras vacías, red nula y más workers que muestras. | Los límites de la API tienen poca cobertura. | Añadir esos casos y pruebas de CSV malformado; ejecutar el detector de carreras. |
| GAP-08 | Medición / Media | `cmd/experiment/main.go`, `measure`: cronometra hasta después de detener el monitor y esperar `monitor.Wait`; muestrea memoria cada 1 ms. | El tiempo incluye instrumentación y puede perder picos entre muestras. | Separar la medición de tiempo de la de recursos y registrar el equipo y la versión de Go. |
| GAP-09 | Medición / Media | `measure_resources.ps1` observa cada proceso cada 100 ms. El proceso incluye carga, entrenamiento e inferencias secuenciales y concurrentes. | La CPU máxima no representa exclusivamente al Worker Pool. | Etiquetar las cifras como máximos del proceso completo y ejecutar modalidades aisladas para compararlas. |
| GAP-10 | Rendimiento / Baja | `PredictWorkerPool` envía un trabajo por muestra. Los tiempos guardados muestran speedup menor que uno. | El coste de coordinación puede superar el ahorro de cálculo de una red pequeña. | Comparar trabajos por lotes y distintas cantidades de workers; confirmar la causa con un perfil de CPU. |
| GAP-11 | Modelo predictivo / Media | `TaxiTrip.Features` usa identificadores de zonas, tarifa y pago como valores numéricos continuos. | La red interpreta distancias numéricas entre categorías que no tienen ese significado. | Comparar con una codificación categórica, como one-hot, y evaluar el efecto sobre MAE y MSE. |
| GAP-12 | Evaluación / Media | Los comandos calculan MAE y MSE, pero no comparan contra una predicción de referencia. | No hay suficiente evidencia para afirmar que la red predice bien. | Comparar con la media de tarifas del entrenamiento sobre el mismo conjunto de evaluación. |
| GAP-13 | Verificación / Informativa | `worker_pool.pml` representa cinco trabajos, tres workers y resultados simbólicos. | Spin comprueba ese protocolo finito; no verifica el cálculo numérico ni todo posible uso de la API Go. | Presentar el alcance y complementar con pruebas de igualdad y `go test -race`. |

## 4. Aspectos correctamente implementados

- **Comunicación por canales:** los workers reciben tareas y envían resultados sin modificar el arreglo final.
- **Cierre coordinado:** el productor cierra `jobs`; el coordinador cierra `results` después de `WaitGroup.Wait`.
- **Escritor único:** el consumidor guarda cada predicción usando su índice original.
- **Validaciones existentes:** se rechazan columnas ausentes, errores de conversión, red nula en la ruta concurrente y cantidades de workers no positivas.
- **Normalización:** los comandos ajustan media y desviación solo con las muestras de entrenamiento.
- **Reproducibilidad parcial:** se fija la semilla de la red y se conservan CSV de las mediciones.

## 5. Interpretación de seguridad y concurrencia

Leer una ruta proporcionada por el usuario es parte del funcionamiento de una aplicación local y no demuestra por sí solo una vulnerabilidad. El riesgo concreto es aceptar contenido inválido o una cantidad de datos que exceda los recursos disponibles. Los errores del lector ya incluyen contexto de la fila, lo que facilita el diagnóstico.

Compartir la red para lectura es apropiado cuando el entrenamiento ya terminó. En ese escenario, los workers no necesitan un mutex para los pesos y solo el consumidor escribe las predicciones. El caso de entrenamiento simultáneo es una limitación de uso de la API, no un fallo observado en los comandos actuales.

También debe precisarse el momento de la predicción: duración y distancia reales se conocen al finalizar el viaje. Si se desea estimar la tarifa antes de iniciar el recorrido, habría que sustituirlas por estimaciones disponibles en ese momento.

## 6. Evidencia y limitaciones

| Comprobación | Estado |
|---|---|
| Lectura del código y revisión del historial | Realizadas sobre la copia local y el commit base indicado. |
| Medias y speedup de los CSV | Recalculados; coinciden con la tabla del informe final. No se repitió el benchmark. |
| Verificación del modelo en Spin | Ejecutada con Spin 6.5.2 en WSL: safety y LTL con `errors: 0`. Registros en `results/spin-*.txt`. |
| Pruebas Go, análisis con `go vet` y detector de carreras | Pendientes: Windows bloquea `go.exe`; además falta generar el CSV usado por las pruebas de integración. |
| Interacción con la IA seleccionada | Pendiente de adjuntar captura o exportación. |

## 7. Recomendaciones prioritarias

1. Preparar las pruebas con un CSV controlado y ejecutarlas en un entorno que permita Go, incluyendo `-race`.
2. Rechazar valores no finitos, fechas incoherentes y entradas que excedan límites razonables.
3. Documentar que no se debe entrenar la red mientras se realizan predicciones.
4. Separar la medición de tiempos y recursos; después comparar tareas por lotes.
5. Evaluar una codificación categórica y una predicción de referencia antes de afirmar mejoras de precisión.

Estas recomendaciones quedan propuestas para una siguiente iteración; su presencia en el informe no significa que ya estén implementadas.
