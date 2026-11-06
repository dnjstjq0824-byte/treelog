# Generates local-only development secrets for TreeBlog on Windows.
# Writes .\infra\dev\jwt_rsa (RS256 private key) and prints .env entries.
$ErrorActionPreference = 'Stop'
$devDir = $PSScriptRoot
$keyPath = Join-Path $devDir 'jwt_rsa'

if (-not (Test-Path $keyPath)) {
  $goSource = @'
package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func main() {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	block := &pem.Block{Type: "PRIVATE KEY", Bytes: der}
	if err := os.WriteFile(os.Args[1], pem.EncodeToMemory(block), 0600); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
'@
  $tmpFile = [IO.Path]::Combine([IO.Path]::GetTempPath(), 'treeblog-genkey-' + [Guid]::NewGuid().ToString('N') + '.go')
  try {
    [IO.File]::WriteAllText($tmpFile, $goSource)
    & go run $tmpFile $keyPath
    if ($LASTEXITCODE -ne 0) { throw 'go run key generator failed' }
    Write-Host "wrote $keyPath"
  } finally {
    Remove-Item -Force -ErrorAction SilentlyContinue $tmpFile
  }
} else {
  Write-Host "kept existing $keyPath"
}

$bytes = New-Object byte[] 32
$rng = New-Object System.Security.Cryptography.RNGCryptoServiceProvider
try {
  $rng.GetBytes($bytes)
} finally {
  $rng.Dispose()
}
$pepper = [Convert]::ToBase64String($bytes)

Write-Host ''
Write-Host 'Add the following to your .env (values are LOCAL ONLY, never commit .env):'
Write-Host ''
Write-Host 'TREEBLOG_JWT_PRIVATE_KEY_PATH=/run/secrets/treeblog/jwt_rsa'
Write-Host "TREEBLOG_CODE_PEPPER=$pepper"
Write-Host ''
Write-Host 'For host-side runs (go run ./cmd/api), use instead:'
Write-Host "TREEBLOG_JWT_PRIVATE_KEY_PATH=$keyPath"
