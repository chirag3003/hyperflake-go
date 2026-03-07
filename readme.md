<h1 align="center" style="border-bottom: none;">❄️ Hyperflake </h1>

<h3 align="center">A simple and lightweight Go module to generate unique snowflake-like IDs starting from a custom epoch. </h3>

<br />

<p align="center">
<a href="https://github.com/chirag3003/hyperflake-go/actions/workflows/test.yml">
<img alt="Build states" src="https://github.com/chirag3003/hyperflake-go/actions/workflows/test.yml/badge.svg?branch=main">
</a>
</p>

<p align="center">
<a href="https://github.com/chirag3003/hyperflake-go/issues/new">Bug report</a>
</p>

<hr />

`hyperflake-go` is a Go library for generating unique and distributed IDs that are suitable for use as primary keys in distributed systems.

It generates 64-bit IDs that are composed of a timestamp, a datacenter ID, a machine ID, and a sequence number. These IDs are based on [Twitter's Snowflake ID](https://github.com/twitter-archive/snowflake/tree/snowflake-2010) generation algorithm.

## ID Structure 🧱

Each generated ID is a 64-bit integer with the following bit layout:

```
 63      62                    22 21        17 16       12 11          0
 ┌──────┬──────────────────────┬────────────┬───────────┬─────────────┐
 │ Sign │      Timestamp       │ Datacenter │  Machine  │  Sequence   │
 │  1b  │       41 bits        │   5 bits   │   5 bits  │   12 bits   │
 └──────┴──────────────────────┴────────────┴───────────┴─────────────┘
```

| Field | Bits | Range | Description |
|-------|------|-------|-------------|
| Sign | 1 | always `0` | Reserved; keeps IDs positive |
| Timestamp | 41 | ~69 year range | Milliseconds since custom epoch |
| Datacenter ID | 5 | 0–31 | Identifies the datacenter |
| Machine ID | 5 | 0–31 | Identifies the node/machine |
| Sequence | 12 | 0–4095 | Per-millisecond counter |

The default epoch is **January 1, 2000 UTC**.

## Installation 🚀

```bash
go get github.com/chirag3003/hyperflake-go
```

## Methods 🧮

The `Config` instance has the following methods:

- `GenerateHyperflakeID()`: Generates a unique ID. **Safe for concurrent use.**
- `DecodeID(id int64)`: Decodes a Hyperflake ID into its components.
- `SetMachineID(machineID int)`: Sets the machine ID.
- `SetDatacenterID(datacenterID int)`: Sets the datacenter ID.
- `GetMachineID()`: Returns the machine ID.
- `GetDatacenterID()`: Returns the datacenter ID.

## Usage 💻

```go
package main

import (
    "fmt"
    "github.com/chirag3003/hyperflake-go"
)

func main() {
    // datacenterID=3, machineID=7
    config := hyperflake.NewHyperflakeConfig(3, 7)

    id, err := config.GenerateHyperflakeID()
    if err != nil {
        fmt.Println("Error generating ID:", err)
        return
    }
    fmt.Println("ID:", id)
}
```

## Custom Epoch 🕰️

Use `NewHyperflakeConfigWithEpoch` to set a custom epoch (in milliseconds):

```go
package main

import (
    "fmt"
    "time"
    "github.com/chirag3003/hyperflake-go"
)

func main() {
    // Use January 1, 2024 as the epoch
    epoch := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
    config := hyperflake.NewHyperflakeConfigWithEpoch(3, 7, epoch)

    id, err := config.GenerateHyperflakeID()
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Println("ID:", id)
}
```

## Decoding IDs ✅

```go
package main

import (
    "fmt"
    "github.com/chirag3003/hyperflake-go"
)

func main() {
    config := hyperflake.NewHyperflakeConfig(3, 7)

    id, err := config.GenerateHyperflakeID()
    if err != nil {
        fmt.Println("Error generating ID:", err)
        return
    }

    decoded, err := config.DecodeID(id)
    if err != nil {
        fmt.Println("Error decoding ID:", err)
        return
    }
    fmt.Printf("Decoded: %+v\n", decoded)
    // Decoded: &{ID:... Signbit:0 DatacenterID:3 MachineID:7 SequenceNumber:0 TimestampSinceEpoch:... Timestamp:...}
}
```

## Concurrency 🔀

`GenerateHyperflakeID` is safe for concurrent use. A single `*Config` can be shared across goroutines:

```go
package main

import (
    "fmt"
    "sync"
    "github.com/chirag3003/hyperflake-go"
)

func main() {
    config := hyperflake.NewHyperflakeConfig(3, 7)

    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            id, err := config.GenerateHyperflakeID()
            if err != nil {
                fmt.Println("Error:", err)
                return
            }
            fmt.Println("ID:", id)
        }()
    }
    wg.Wait()
}
```

## Error Handling 😱

`GenerateHyperflakeID` returns an error if the system clock moves backwards (e.g. due to NTP adjustments or manual clock changes):

```go
id, err := config.GenerateHyperflakeID()
if err != nil {
    // err.Error() == "clock is moving backwards"
    log.Fatal(err)
}
```

## Bugs or Requests 🐛

If you encounter any problems feel free to open an [issue](https://github.com/chirag3003/hyperflake-go/issues/new) on GitHub.