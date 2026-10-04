# Red neuronal y procesamiento concurrente

## Resumen del trabajo

El proyecto implementa una red neuronal pequeña para regresión de `fare_amount` sobre registros de taxis. La red se entrena de forma secuencial y luego se comparan dos estrategias de inferencia:

- **Secuencial:** un registro por vez en una única goroutine.
- **Concurrente:** patrón **Worker Pool**, donde varios workers ejecutan el forward pass de distintos registros de forma independiente.

La implementación de la red se realizó desde cero en Go, sin utilizar una biblioteca de Machine Learning. El dataset se carga desde un CSV limpio ubicado en `processed/yellow_tripdata_2026-01.csv`.

## Modelo de datos y red neuronal

Cada registro utiliza nueve variables de entrada:

1. `passenger_count`
2. `trip_distance`
3. `trip_time_min`
4. `RatecodeID`
5. `PULocationID`
6. `DOLocationID`
7. `payment_type`
8. `pickup_hour`
9. `pickup_day_of_week`

La variable objetivo es `fare_amount`.

La arquitectura utilizada es:

```text
Input(9)
   |
Dense(16) + ReLU
   |
Dense(8) + ReLU
   |
Dense(1) + Linear
   |
Prediccion de fare_amount
```

Las entradas se normalizan utilizando la media y la desviación estándar calculadas sobre el conjunto de entrenamiento. Las estadísticas de normalización no se calculan sobre el conjunto de evaluación, evitando contaminación entre entrenamiento y evaluación.

El entrenamiento utiliza descenso de gradiente por lote completo, función de activación ReLU y error cuadrático medio (MSE). La inferencia no modifica los pesos de la red.

## j. Implementación secuencial y concurrente

La implementación secuencial recorre el arreglo de muestras y ejecuta `Predict` para cada registro:

```text
para cada muestra:
    prediccion = red.forward(muestra.features)
    guardar prediccion
```

La implementación concurrente utiliza un Worker Pool:

```text
Productor
    |
    v
Canal de trabajos
    |
    +--> Worker 1 --> forward --> resultado
    +--> Worker 2 --> forward --> resultado
    +--> Worker 3 --> forward --> resultado
    |
Canal de resultados
    |
Predicciones ordenadas
```

Cada trabajo contiene el índice y la muestra. El resultado contiene el mismo índice y la predicción. De esta forma, los workers pueden terminar en distinto orden sin alterar la correspondencia entre registro y predicción.

La implementación se encuentra en:

- `go-code/redneuronal.go`
- `go-code/inference.go`
- `go-code/dataset.go`

## k. Algoritmo y mecanismos de sincronización

El productor envía trabajos al canal `jobs`. Los workers reciben trabajos hasta que el canal se cierra, realizan el forward pass y envían un resultado al canal `results`.

Los mecanismos utilizados son:

- **Channels:** comunicación entre el productor y los workers, y entre los workers y el consumidor de resultados.
- **Goroutines:** ejecución concurrente de los workers.
- **`sync.WaitGroup`:** espera a que todos los workers terminen antes de cerrar el canal de resultados.
- **Índices de resultados:** cada worker escribe conceptualmente en una posición distinta del arreglo final.

La red neuronal se comparte entre los workers solamente para lectura. Como los pesos no cambian durante la inferencia, no es necesario utilizar un mutex para proteger la red. Cada resultado tiene un índice único, por lo que no hay dos workers escribiendo sobre la misma posición.

La función de evaluación calcula:

$$
MAE = \frac{1}{n}\sum_{i=1}^{n}|y_i - \hat{y}_i|
$$

$$
MSE = \frac{1}{n}\sum_{i=1}^{n}(y_i - \hat{y}_i)^2
$$

## Modelo inicial en Promela

El modelo ubicado en `promela/worker_pool.pml` representa tres workers y cinco trabajos. Los trabajos se comunican mediante el canal `jobs` y las finalizaciones mediante `results`.

La sección crítica abstracta se expresa mediante un bloque `atomic`:

```promela
atomic {
    assert(job < JOBS);
    assert(!completed[job]);
    completed[job] = true;
    completedCount++;
}
```

Las aserciones verifican que:

- El identificador del trabajo sea válido.
- Un trabajo no sea marcado dos veces como completado.
- El número de trabajos completados no supere el total.

Esto modela la ausencia de una condición de carrera en la actualización del estado compartido.

La verificación se ejecutó con Spin 6.4.9 dentro de Docker en dos modos. En modo safety, que habilita la detección de estados finales inválidos, se exploraron 55.021 estados almacenados y 100.425 transiciones, con `errors: 0`. En modo LTL se exploraron 84.852 estados almacenados y 271.011 transiciones, también con `errors: 0`:

```text
Safety: errors: 0
55021 states, stored
100425 transitions

LTL: errors: 0
84852 states, stored
271011 transitions
```

El modo safety comprueba ausencia de deadlocks/estados finales inválidos y violaciones de aserciones. El modo LTL comprueba la propiedad de progreso `all_jobs_complete`: mientras existan trabajos pendientes, eventualmente todos deben completarse. La aserción `assert(!inCritical)` representa la exclusión mutua de la sección crítica. El resultado `errors: 0` confirma estas propiedades para el espacio de estados finito definido por tres workers y cinco trabajos.

## l. Speedup y media recortada

El speedup se calculó mediante:

$$
Speedup = \frac{T_{secuencial}}{T_{concurrente}}
$$

La medición se realizó con la siguiente configuración:

- 100.000 filas cargadas.
- 10.000 filas utilizadas para entrenamiento.
- 90.000 filas utilizadas para evaluación.
- Una época de entrenamiento.
- Semilla de inicialización igual a `42`.
- 11 ejecuciones por configuración.
- Se eliminó el 20% de los valores más bajos y el 20% de los más altos.
- La media se calculó con las siete ejecuciones restantes.
- El tiempo de carga y entrenamiento se excluyó del tiempo de inferencia.

| Modelo | Workers | Media recortada (ms) | Desv. estándar (ms) | Speedup |
|---|---:|---:|---:|---:|
| Secuencial | 0 | 26.083 | 0.191 | 1.000 |
| Worker Pool | 1 | 69.143 | 0.833 | 0.377 |
| Worker Pool | 2 | 60.986 | 0.791 | 0.428 |
| Worker Pool | 4 | 59.143 | 1.641 | 0.441 |
| Worker Pool | 8 | 63.999 | 1.308 | 0.408 |

Los resultados completos se encuentran en `results/inference_runs.csv` y el gráfico correspondiente en `results/inference_speedup.png`.

![Tiempo de inferencia y speedup](../results/inference_speedup.png)

## m. Análisis de speedup, escalabilidad y trade-offs

La solución secuencial obtuvo el mejor tiempo para esta configuración. El mejor resultado del Worker Pool se obtuvo con cuatro workers, pero su speedup fue aproximadamente `0.441`, menor que uno. Esto significa que la versión concurrente tardó más que la secuencial.

La razón principal es que el forward pass de la red es pequeño: solamente contiene dos capas ocultas de 16 y 8 neuronas. En consecuencia, el coste de enviar trabajos por canales, planificar goroutines y coordinar la finalización de los workers es significativo frente al coste del cálculo matemático.

El aumento de workers tampoco produjo una mejora monotónica:

- Con un worker, el overhead es máximo porque se agrega concurrencia sin paralelismo real.
- Con dos workers, el tiempo mejora respecto a un worker.
- Con cuatro workers se obtiene el mejor resultado concurrente.
- Con ocho workers el tiempo aumenta nuevamente, debido al overhead de coordinación y planificación.

La solución concurrente sería más conveniente si cada trabajo tuviera un forward pass más costoso, si la red fuera más grande o si se procesaran lotes suficientemente grandes como para amortizar el coste de comunicación.

## n. Uso y rendimiento de recursos

La medición se realizó con 100.000 filas, 10.000 filas de entrenamiento y 11 ejecuciones por configuración. Para cada configuración se ejecutó el proceso de experimentación por separado. La aplicación registró memoria y goroutines mediante `runtime.ReadMemStats`; un script de PowerShell registró CPU y working set del proceso.

| Workers | CPU máxima del proceso (%) | Working set máximo (MB) | Heap máximo de Go (MB) | Memoria `Sys` máxima (MB) | Goroutines máximas |
|---:|---:|---:|---:|---:|---:|
| 1 | 14.61 | 46.05 | 36.48 | 52.39 | 5 |
| 2 | 19.32 | 46.14 | 36.48 | 52.16 | 6 |
| 4 | 22.72 | 46.48 | 36.50 | 52.15 | 8 |
| 8 | 22.61 | 46.52 | 36.52 | 52.21 | 12 |

Los resultados completos se encuentran en `results/resource_runs.csv`. La tabla de recursos se generó con `scripts/measure_resources.ps1`.

La CPU máxima aumenta desde 14.61% con un worker hasta 22.72% con cuatro workers. Con ocho workers no se observa una mejora adicional y el valor disminuye ligeramente a 22.61%, lo cual es consistente con el hecho de que el equipo tiene otros costes de planificación y sincronización.

El working set se mantiene aproximadamente entre 46.05 MB y 46.52 MB. El heap de Go también permanece prácticamente constante, entre 36.48 MB y 36.52 MB. Esto indica que agregar workers no produce un crecimiento significativo de memoria para este diseño, porque los workers comparten la red y procesan las muestras existentes.

El número máximo de goroutines sí aumenta con la cantidad de workers: pasa de 5 con un worker a 12 con ocho workers. Este aumento confirma que el paralelismo se está creando, aunque no se traduzca directamente en una reducción del tiempo de inferencia.

La tabla anterior debería interpretarse como uso máximo observado durante cada proceso experimental. La CPU fue muestreada externamente con PowerShell y la memoria/goroutines fueron observadas desde Go.

Un mayor uso de CPU no implica necesariamente un menor tiempo de ejecución. En este caso, cuatro workers utilizan más CPU que la versión con un worker, pero el coste de coordinación sigue siendo superior a la ganancia obtenida por el paralelismo.

## Reproducibilidad

Para ejecutar las pruebas:

```powershell
go test -count=1 ./...
go vet ./...
```

Para repetir la verificación Promela utilizando Docker Desktop:

```powershell
docker run --rm --entrypoint sh `
    -v "${PWD}\promela:/model" `
    kuniwak/spin `
    -lc "apk add --no-cache gcc musl-dev >/dev/null && spin -run /model/worker_pool.pml"
```

Para ejecutar una medición repetida:

```powershell
go run ./cmd/experiment `
  -max-rows 100000 `
  -train-rows 10000 `
  -epochs 1 `
  -workers 1,2,4,8 `
  -runs 11 `
  -trim 0.2 `
  -output results/inference_runs.csv
```

Para generar el gráfico:

```powershell
py -3 scripts/plot_experiment.py
```

Para ejecutar con el dataset completo, se puede utilizar `-max-rows 0`, aunque el tiempo y el consumo de memoria serán mayores:

```powershell
go run ./cmd/experiment -max-rows 0 -workers 1,2,4,8
```
