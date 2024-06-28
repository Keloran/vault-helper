# Simple helper for Vault

## Remote Usage

```go
package main

import (
  "fmt"
  "log"

  vault "github.com/keloran/vault-helper"
)

func main() {
  v := vault.NewVault("vault.vault", "vault-token")
  err := v.GetRemoteSecrets("kv/secret")
  if err != nil {
    log.Fatal(err)
  }

  sec, err := v.GetSecret("tester")
  if err != nil {
    log.Fatal(err)
  }

  fmt.Println(v.KVSecrets)
  fmt.Println(sec)
}
```

## Local Usage

```go
package main

import (
  "fmt"
  "log"

  vault "github.com/keloran/vault-helper"
)

func main() {
  v := vault.NewVault("vault.vault", "vault-token")
  err := v.GetLocalSecrets("/secrets/secrets.json")
  if err != nil {
    log.Fatal(err)
  }

  sec, err := v.GetSecret("tester")
  if err != nil {
    log.Fatal(err)
  }

  fmt.Println(v.KVSecrets)
  fmt.Println(sec)
}
```
