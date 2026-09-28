param(
    [int]$MaxRows = 100000,
    [int]$TrainRows = 10000,
    [int]$Runs = 11,
    [string]$Workers = "1,2,4,8",
    [string]$Output = "results/resource_runs.csv"
)

$ErrorActionPreference = "Stop"
$workerValues = $Workers.Split(",") | ForEach-Object { [int]$_.Trim() }
$binary = "results/experiment.exe"

New-Item -ItemType Directory -Force -Path "results" | Out-Null
go build -o $binary ./cmd/experiment

$rows = @()
foreach ($workerCount in $workerValues) {
    $stdout = "results/experiment-$workerCount.out"
    $stderr = "results/experiment-$workerCount.err"
    $process = Start-Process -FilePath (Resolve-Path $binary) -PassThru -RedirectStandardOutput $stdout -RedirectStandardError $stderr -ArgumentList @(
        "-max-rows", $MaxRows,
        "-train-rows", $TrainRows,
        "-runs", $Runs,
        "-workers", $workerCount,
        "-output", "results/inference-$workerCount.csv"
    )

    $maxCpu = 0.0
    $maxWorkingSet = 0L
    $previousCpu = $process.TotalProcessorTime.TotalMilliseconds
    $previousTime = Get-Date
    while (-not $process.HasExited) {
        Start-Sleep -Milliseconds 100
        $process.Refresh()
        $now = Get-Date
        $cpuDelta = $process.TotalProcessorTime.TotalMilliseconds - $previousCpu
        $timeDelta = ($now - $previousTime).TotalMilliseconds
        if ($timeDelta -gt 0) {
            $cpuPercent = 100 * $cpuDelta / $timeDelta / [Environment]::ProcessorCount
            if ($cpuPercent -gt $maxCpu) { $maxCpu = $cpuPercent }
        }
        if ($process.WorkingSet64 -gt $maxWorkingSet) { $maxWorkingSet = $process.WorkingSet64 }
        $previousCpu = $process.TotalProcessorTime.TotalMilliseconds
        $previousTime = $now
    }
    $process.WaitForExit()

    $measurements = Import-Csv "results/inference-$workerCount.csv"
    $maxHeap = ($measurements | Measure-Object -Property peak_heap_bytes -Maximum).Maximum
    $maxSys = ($measurements | Measure-Object -Property peak_sys_bytes -Maximum).Maximum
    $maxGoroutines = ($measurements | Measure-Object -Property max_goroutines -Maximum).Maximum
    $rows += [pscustomobject]@{
        workers = $workerCount
        max_cpu_percent = [math]::Round($maxCpu, 2)
        max_working_set_mb = [math]::Round($maxWorkingSet / 1MB, 2)
        max_heap_mb = [math]::Round([double]$maxHeap / 1MB, 2)
        max_runtime_sys_mb = [math]::Round([double]$maxSys / 1MB, 2)
        max_goroutines = $maxGoroutines
    }
}

$rows | Export-Csv -NoTypeInformation -Path $Output
Write-Output "Resource results written to $Output"
