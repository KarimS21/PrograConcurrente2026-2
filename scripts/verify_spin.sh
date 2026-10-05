#!/bin/sh
# Ejecutar desde Linux o WSL con Spin y GCC disponibles.
set -eu

project_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
output_dir="$project_dir/results"
check_dir=$(mktemp -d)
trap 'rm -rf -- "$check_dir"' EXIT HUP INT TERM

mkdir -p "$output_dir"
cd "$check_dir"
if ! command -v spin >/dev/null 2>&1; then
    # En Ubuntu/Debian, usar una copia temporal sin instalar paquetes del sistema.
    apt-get download spin
    dpkg-deb -x spin*.deb tool
    PATH="$check_dir/tool/usr/bin:$PATH"
    export PATH
fi
{
    spin -V
    gcc --version | head -n 1
    sha256sum "$project_dir/promela/worker_pool.pml"
} > "$output_dir/spin-environment.txt"

spin -a "$project_dir/promela/worker_pool.pml"
gcc -O2 -DNOCLAIM -DSAFETY -o pan-safety pan.c
./pan-safety -b -q > "$output_dir/spin-safety.txt" 2>&1
gcc -O2 -o pan-ltl pan.c
./pan-ltl -a -b > "$output_dir/spin-ltl.txt" 2>&1

# Pan puede devolver 0 aun si encuentra errores: comprobar tambien la salida.
for report in "$output_dir/spin-safety.txt" "$output_dir/spin-ltl.txt"; do
    if ! grep -Eq 'errors: 0([[:space:]]|$)' "$report" ||
        grep -Eiq 'search not completed|depth limit reached|out of memory' "$report"; then
        cat "$report"
        exit 1
    fi
done
printf '%s\n' 'Spin: safety y LTL completados con errors: 0. Ver results/spin-*.txt.'
