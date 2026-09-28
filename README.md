# Predicción de Tarifas de Taxi en NYC mediante Redes Neuronales y Concurrencia en Go

[![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)](#) [![Python](https://img.shields.io/badge/Python-3776AB?style=for-the-badge&logo=python&logoColor=white)](#) [![Jupyter](https://img.shields.io/badge/Jupyter-F37626?style=for-the-badge&logo=Jupyter&logoColor=white)](#) [![UPC](https://img.shields.io/badge/UPC-E3000F?style=for-the-badge&logo=upc&logoColor=white)](#)

Repositorio oficial del trabajo final del curso **1ACC0065 - Programación Concurrente y Distribuida** (Ciencias de la Computación, UPC 2026). 

Este proyecto implementa técnicas de paralelismo a nivel de tareas (*Task-level parallelism*), mediante patrones concurrentes en **Go** (*Worker Pools*, *Pipelines*), para el entrenamiento y predicción de un modelo de Red Neuronal sobre un conjunto de datos masivo. El proyecto se alinea con el **ODS 11 (Ciudades y Comunidades Sostenibles)** de la ONU al buscar la optimización y comprensión de patrones de transporte urbano.

## Equipo de Desarrollo

| Código | Nombres y Apellidos |
| :--- | :--- |
| U202221147 | Juan Diego Adrián Sánchez Sánchez |
| U201816862 | Karim Wagner Samanamud Mosquera |
| U201916314 | Tomas Alonso Pastor Salazar |

## El Dataset: NYC Yellow Taxi Trip Records
El objetivo del modelo es predecir el costo de un viaje en taxi (`fare_amount`) en Nueva York. A partir de una muestra inicial de 1.48 millones de registros extraída del formato `.parquet`, se realizó un proceso estricto de limpieza y perfilamiento para evitar *data leakage* (fuga de datos) y asegurar la consistencia temporal/matemática de las variables.

**Pasos de Limpieza:**
1. **Selección de Variables:** Se descartaron métricas calculadas post-viaje o que sesgan el modelo predictivo base (ej. propinas `tip_amount`, peajes `tolls_amount`, impuestos y recargos adicionales `extra`, `mta_tax`, etc.).
2. **Imputación y Formateo:** Tratamiento de nulos en `passenger_count` (reemplazados por 0 temporalmente) y conversión a tipos enteros.
3. **Filtros Lógicos:** Retención exclusiva de viajes con sentido operativo:
   * `trip_distance > 0`
   * `fare_amount > 0`
   * `passenger_count > 0`
   * `tpep_dropoff_datetime > tpep_pickup_datetime` (Coherencia temporal).
4. **Resultado Final:** Dataset 100% limpio y tabular con **1,002,542 registros** listos para ser consumidos por el algoritmo en Go.

## Fases del Proyecto (Entregables)

- [x] **Entregable 1 (PC1):** Investigación de Literatura (Estado del Arte de Concurrencia en Machine Learning), Selección y Limpieza del Dataset (>1 Millón de filas).
