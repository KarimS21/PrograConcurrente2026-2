import csv
import statistics
from pathlib import Path

import matplotlib.pyplot as plt


INPUT = Path("results/inference_runs.csv")
OUTPUT = Path("results/inference_speedup.png")
TRIM_FRACTION = 0.2


def trimmed_mean(values):
    ordered = sorted(values)
    trim = int(len(ordered) * TRIM_FRACTION)
    kept = ordered[trim:len(ordered) - trim]
    return statistics.mean(kept), statistics.pstdev(kept)


def read_measurements():
    groups = {}
    with INPUT.open(newline="") as file:
        for row in csv.DictReader(file):
            key = (row["mode"], int(row["workers"]))
            groups.setdefault(key, []).append(float(row["duration_ns"]) / 1_000_000)
    return groups


groups = read_measurements()
sequential_mean, _ = trimmed_mean(groups[("sequential", 0)])
pool_groups = sorted((workers, values) for (mode, workers), values in groups.items() if mode == "worker_pool")
workers = [worker for worker, _ in pool_groups]
means = [trimmed_mean(values)[0] for _, values in pool_groups]
speedups = [sequential_mean / mean for mean in means]

figure, axes = plt.subplots(1, 2, figsize=(11, 4.5))
axes[0].axhline(sequential_mean, color="#444444", linestyle="--", label="Secuencial")
axes[0].plot(workers, means, marker="o", color="#1769aa", label="Worker Pool")
axes[0].set_title("Tiempo de inferencia")
axes[0].set_xlabel("Cantidad de workers")
axes[0].set_ylabel("Media recortada (ms)")
axes[0].set_xticks(workers)
axes[0].grid(alpha=0.25)
axes[0].legend()

axes[1].axhline(1, color="#444444", linestyle="--", label="Referencia")
axes[1].plot(workers, speedups, marker="o", color="#d95f02")
axes[1].set_title("Speedup respecto al secuencial")
axes[1].set_xlabel("Cantidad de workers")
axes[1].set_ylabel("$T_{secuencial} / T_{concurrente}$")
axes[1].set_xticks(workers)
axes[1].grid(alpha=0.25)

figure.tight_layout()
OUTPUT.parent.mkdir(parents=True, exist_ok=True)
figure.savefig(OUTPUT, dpi=180)
print(f"Figura escrita en {OUTPUT}")
