# Runs everything CI runs, locally: gofmt check, go vet, go test in both modules (the
# nvim oracle is its own module, DESIGN §8.2), a short fuzz pass.
#   ./scripts/test.ps1                # all; the oracle skips without nvim
#   ./scripts/test.ps1 -Fuzz 0        # skip fuzzing
#   ./scripts/test.ps1 -RequireNvim   # fail, not skip, without nvim (CI)
param(
    [int]$Fuzz = 10,  # seconds per fuzz target; 0 to skip
    [switch]$RequireNvim
)
$ErrorActionPreference = 'Stop'
$root = Split-Path $PSScriptRoot -Parent
Push-Location $root
try {
    $files = gofmt -l .
    if ($files) { $files; throw 'gofmt needed' }
    foreach ($mod in '.', 'oracle') {
        Push-Location $mod
        try {
            go vet ./...
            if ($LASTEXITCODE) { throw "go vet failed in $mod" }
            if ($mod -eq 'oracle' -and $RequireNvim) { $env:VID_REQUIRE_NVIM = '1' }
            go test ./...
            if ($LASTEXITCODE) { throw "go test failed in $mod" }
        } finally {
            Remove-Item Env:VID_REQUIRE_NVIM -ErrorAction Ignore
            Pop-Location
        }
    }

    if ($Fuzz -gt 0) {
        # @('./pkg', 'FuzzName') per target, as they land (Q-003, V-001, C-001)
        $targets = @()
        foreach ($t in $targets) {
            go test $t[0] -run '^$' -fuzz "^$($t[1])$" -fuzztime "$($Fuzz)s"
            if ($LASTEXITCODE) { throw "fuzz $($t[1]) failed" }
        }
    }
} finally {
    Pop-Location
}
