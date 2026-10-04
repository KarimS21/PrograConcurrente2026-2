# Informe de GAPs técnicos

## 1. Alcance y evidencia

Repositorio analizado: [PrograConcurrente2026-2](https://github.com/KarimS21/PrograConcurrente2026-2).

La revisión cubre la red neuronal, el lector CSV, el Worker Pool, los experimentadores, los scripts de medición, el modelo Promela y las pruebas Go. El commit remoto de referencia es `0cd16be`, rama `main`; el análisis incluye además cambios locales no publicados en el modelo Promela y en esta documentación.

Comandos ejecutados:

```text
go test -count=1 ./...     PASS
go vet ./...               PASS
Spin 6.4.9, modo safety   errors: 0
Spin 6.4.9, modo LTL      errors: 0
```

El análisis es estático y experimental sobre el árbol disponible. No sustituye una auditoría de producción, un análisis de dependencias CVE ni una prueba con todos los escenarios posibles del dataset.

## 2. Resumen ejecutivo

La implementación cumple el objetivo académico principal: separa entrenamiento e inferencia, implementa una red neuronal desde cero y compara una ruta secuencial con un Worker Pool. La verificación formal del modelo Promela no encontró deadlocks ni violaciones de las aserciones de exclusión mutua en el espacio de estados modelado.

Los principales gaps están en la robustez del pipeline de datos, la representación de variables categóricas, la cobertura de pruebas negativas, la medición de recursos y la distancia entre el modelo Promela y la implementación Go real. No se observó una condición de carrera en la inferencia implementada bajo el supuesto de que los pesos no se modifican durante `Predict`.

## 3. Hallazgos priorizados

| ID | Severidad | Ubicación | GAP | Impacto | Recomendación |
|---|---|---|---|---|---|
| GAP-01 | Media | `go-code/redneuronal.go`, `TaxiTrip.Features` | `RatecodeID`, `PULocationID`, `DOLocationID` y `payment_type` se tratan como números continuos. | La red puede interpretar que la zona 200 es aproximadamente el doble de la zona 100, aunque son categorías sin orden numérico. | Usar one-hot, hashing controlado o embeddings; documentar la decisión si se conserva la línea base numérica. |
| GAP-02 | Media | `go-code/dataset.go` | Se exige `trip_time_min`, pero el valor leído no se valida ni se utiliza; se recalcula desde las fechas. | Un CSV inconsistente puede pasar silenciosamente y producir resultados distintos de la columna preparada. | Validar la diferencia entre ambos valores con tolerancia o eliminar la columna del contrato de entrada. |
| GAP-03 | Media | `go-code/dataset.go` | La carga materializa todas las muestras en memoria y no limita tamaño de archivo o valores numéricos. | Un archivo local muy grande o malformado puede provocar consumo excesivo de memoria o `NaN`/Inf en la red. | Añadir límites, validar `math.IsNaN`/`math.IsInf`, rangos y ofrecer procesamiento por lotes. |
| GAP-04 | Media | `go-code/inference.go` | La función concurrente valida red nula, pero `PredictSequential` no lo hace. | API inconsistente y posible panic ante uso incorrecto. | Devolver error o definir explícitamente el contrato de ambas funciones. |
| GAP-05 | Media | `cmd/experiment/main.go` | El monitor de `runtime.MemStats` muestrea cada milisegundo y puede perder picos en ejecuciones cortas. | Los máximos de heap y goroutines pueden estar subestimados. | Medir antes/después, usar un monitor de mayor duración o separar una prueba específica de recursos. |
| GAP-06 | Media | `scripts/measure_resources.ps1` | CPU se muestrea cada 100 ms y el proceso ejecuta varias modalidades dentro de una misma corrida. | La cifra es un máximo del proceso completo, no una atribución precisa por modo. | Ejecutar un proceso por modalidad, reducir el intervalo y etiquetar claramente la métrica. |
| GAP-07 | Baja | `go-code/inference.go` | Cada llamada al Worker Pool crea goroutines de productor y cierre, aunque el forward pass sea pequeño. | El overhead domina el cálculo y produce speedup menor que uno en la configuración medida. | Usar lotes, aumentar granularidad por trabajo o comparar también un pool persistente. |
| GAP-08 | Baja | `go-code/*_test.go` | Faltan pruebas de CSV malformado, columnas ausentes, valores vacíos, límites y cancelación. | Errores de entrada y regresiones en la ruta de producción no quedan cubiertos. | Añadir tablas de casos inválidos y pruebas de comportamiento de canales. |
| GAP-09 | Baja | `promela/worker_pool.pml` | El modelo abstrae el arreglo de predicciones, la red neuronal y el consumidor de resultados; usa mensajes `STOP` en lugar del cierre real de un canal Go. | La evidencia formal aplica al protocolo abstracto, no a toda la implementación. | Documentar el alcance y añadir propiedades sobre correspondencia índice-resultado en un modelo más detallado. |
| GAP-10 | Informativa | Repositorio | No se observa un pipeline CI que ejecute `go test`, `go vet`, Spin y el benchmark. | Cambios futuros pueden romper verificaciones sin detectarse antes de entregar. | Agregar un workflow de GitHub Actions con pruebas y validaciones reproducibles. |

## 4. Concurrencia y seguridad

La inferencia concurrente comparte la red únicamente para lectura. Los workers reciben muestras por canal y cada resultado conserva su índice original. `sync.WaitGroup` coordina el cierre del canal de resultados. Bajo este contrato, no se observó una escritura concurrente conflictiva.

El modelo Promela utiliza una sección `atomic` y la variable `inCritical` para comprobar que no haya dos workers en la sección crítica. Los mensajes `STOP` permiten terminar los workers, evitando el deadlock artificial que existía cuando el modelo dejaba a los workers esperando indefinidamente por nuevos trabajos.

El lector CSV devuelve errores contextualizados con número de fila, lo cual es positivo. Sin embargo, recibe una ruta desde un flag y no valida tamaño, contenido numérico finito ni límites de negocio. Esto es aceptable para un experimento local, pero debe endurecerse si se expone como servicio o se procesan archivos no confiables.

## 5. Pipeline de Machine Learning

La separación entre entrenamiento y evaluación es correcta: el normalizador se ajusta con las filas de entrenamiento y luego se aplica al conjunto completo. La semilla fija permite reproducir la inicialización.

La principal limitación metodológica es la codificación numérica de variables categóricas. Además, el objetivo no se normaliza, por lo que la estabilidad del descenso de gradiente depende de la escala de `fare_amount` y del learning rate elegido. La evaluación actual mide MAE y MSE, pero no incluye comparación contra un baseline simple, intervalos de confianza ni validación cruzada.

## 6. Plan de remediación

1. Mantener las verificaciones Go y Promela en CI.
2. Añadir validación de finitud y rangos al lector CSV.
3. Cubrir errores de entrada y casos límite del Worker Pool.
4. Separar mediciones de recursos por modalidad y mejorar el muestreo.
5. Comparar la codificación numérica contra one-hot o embeddings.
6. Evaluar un baseline, por ejemplo la media del `fare_amount`, antes de interpretar la calidad de la red.
7. Optimizar la granularidad de trabajo antes de aumentar la cantidad de workers.
