# Simple helper for Vault

## Usage

```go
package main

import (
  "fmt"
  "log"

  vault "github.com/keloran/vault-helper"
)

func main() {
  v := vault.NewVault("vault.vault", "vault-token")
  err := v.GetSecrets("kv/secret")
  if err != nil {
    log.Fatal(err)
  }

  sec, err := v.GetSecret("tester")
  if err != nil {
    log.Fatal(err)
  }

  fmt.Println(v.Secrets)
  fmt.Println(sec)
}
```

