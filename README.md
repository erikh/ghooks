# ghooks -- golang github hooks handler

## Notice

This is a heavily refactored version of https://github.com/Konboi/ghooks. It
includes several code refactors and simplifications.

## About ghooks

ghooks is github hooks receiver. inspired by
[GitHub::Hooks::Receiver](https://github.com/Songmu/Github-Hooks-Receiver),
[octoks](https://github.com/hisaichi5518/octoks)

# Install

```
go get github.com/erikh/ghooks
```

# Usage

```go
// sample.go
package main

import (
    "fmt"
    "log"

    "github.com/erikh/ghooks"
)


func main() {
    port := 8080
    hooks := ghooks.NewServer(port)

    hooks.On("push", pushHandler)
    hooks.On("pull_request", pullRequestHandler)
    hooks.Run()
}

func pushHandler(payload interface{}) {
    fmt.Println("puuuuush")
}

func pullRequestHandler(payload interface{}) {
    fmt.Println("pull_request")
}
```

```
go run sample.go
```

```
curl -H "X-GitHub-Event: push" -d '{"hoge":"fuga"}' http://localhost:8080
> puuuuush
```
