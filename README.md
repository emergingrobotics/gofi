<img src="images/gofi-logo.png" width="33%" alt="gofi logo">

# gofi - Go UniFi Controller Client

[![Go Reference](https://pkg.go.dev/badge/github.com/unifi-go/gofi.svg)](https://pkg.go.dev/github.com/unifi-go/gofi/src)
[![Go Report Card](https://goreportcard.com/badge/github.com/unifi-go/gofi)](https://goreportcard.com/report/github.com/unifi-go/gofi)

Programmatic control of Ubiquiti UniFi UDM Pro devices. This repository is two things:

- **Command-line program** — a unified tool (`gofi`) with subcommands for managing fixed IPs,
  listing clients, inspecting networks, and more.
- **A Go module (SDK)** — a type-safe, concurrent-safe client library you can import into
  your own programs.

If you just want to get work done from the shell, start with **[Part 1: Command-line
programs](#part-1-command-line-programs)**. If you're writing Go code against a UniFi
controller, jump to **[Part 2: Using gofi in your Go program](#part-2-using-gofi-in-your-go-program-sdk)**.

---

# Part 1: Command-line programs

## Step 1 — Get a UniFi API key (recommended)

gofi supports two ways to authenticate. A **cloud API key** used through Ubiquiti's Site
Manager connector is the recommended one: it works even when the controller isn't directly
reachable on your LAN, needs no session cookies, and is the preferred authentication method.

**Create the key:**

1. Sign in at [unifi.ui.com](https://unifi.ui.com) using a **Site Admin** or **Owner**
   account. A key inherits the permissions of the account that created it.
2. Go to your **profile icon → API → Create API Key**.
3. Grant it the **UniFi Applications → Network** scope.

> Cloud keys (from unifi.ui.com) are a different credential from console-issued keys (from
> the console's own UI). Connector access requires the **cloud** key.

**Find your console ID** (the connector needs it):

```bash
export UNIFI_API_KEY=...    # the key you just created
curl -s -H "X-API-KEY: $UNIFI_API_KEY" https://api.ui.com/v1/hosts
```

**Export both values** so the programs can find them:

```bash
export UNIFI_API_KEY=...
export UNIFI_CONSOLE_ID=...
```

That's it — every request now goes to `https://api.ui.com`, which forwards it to your
console. No direct network route to the UDM is required.

### Alternative — local username/password

The cloud API key above is the intended path — it's what most of gofi's own testing and
day-to-day use goes through. If the controller isn't reachable at `api.ui.com` at all
(an isolated network, no internet route) or you'd rather not create a cloud key, a local
username/password path exists as a secondary, **less-tested** alternative. See
[`docs/alternate-local-api.md`](docs/alternate-local-api.md) for the full setup,
including TLS/self-signed-certificate handling. If you're not sure which to use, use
the cloud API key above.

## Step 2 — Build and install the programs

```bash
make install        # build the utilities and install them to ~/bin
make utilities      # or just build them into ./bin
make examples       # build the example programs into ./bin/examples
```

`make install` puts `gofi` on your `PATH` (override the destination with `make install INSTALL_DIR=/somewhere/else`).

## Step 3 — Shell completion (optional)

`gofi completion <shell>` doesn't configure anything by itself — it prints a completion
script to stdout. What you do with that output determines whether completion is
temporary (this shell session only) or permanent (every new shell).

**Try it first, without installing anything:**

```bash
source <(gofi completion bash)   # bash
gofi completion zsh | source     # zsh (as a one-off; see below for the real setup)
```

Completion works for this shell session only; open a new terminal and it's gone. Use
this to confirm completion behaves the way you want before wiring it in permanently.

**Install permanently:**

<details>
<summary>bash</summary>

Requires the `bash-completion` package (`apt install bash-completion`, `brew install
bash-completion`, etc.) — without it, sourced completion scripts are silently ignored.

System-wide (needs root, affects every user):

```bash
gofi completion bash | sudo tee /etc/bash_completion.d/gofi > /dev/null
```

Per-user, no root required:

```bash
mkdir -p ~/.local/share/bash-completion/completions
gofi completion bash > ~/.local/share/bash-completion/completions/gofi
```

Restart your shell, or `source` the file directly, to pick it up.

</details>

<details>
<summary>zsh</summary>

Pick a directory already on your `fpath` (check with `echo $fpath`), or add one:

```bash
mkdir -p ~/.zsh/completions
gofi completion zsh > ~/.zsh/completions/_gofi
```

Then, in `~/.zshrc`, before the `compinit` line:

```zsh
fpath=(~/.zsh/completions $fpath)
autoload -Uz compinit && compinit
```

If you use a framework (oh-my-zsh, prezto), drop `_gofi` into its custom completions
directory instead (e.g. `~/.oh-my-zsh/custom/completions/`) and skip the `fpath` edit.

Restart your shell after making this change — `zsh` only rebuilds its completion cache
on `compinit`.

</details>

<details>
<summary>fish</summary>

```fish
gofi completion fish > ~/.config/fish/completions/gofi.fish
```

Picked up automatically in new fish sessions — no restart-triggering config edit needed.

</details>

<details>
<summary>powershell</summary>

```powershell
gofi completion powershell | Out-String | Invoke-Expression
```

To make this permanent, add that line to your PowerShell profile (`$PROFILE`).

</details>

Once installed, `gofi <TAB>` lists areas, `gofi ips <TAB>` lists actions, and
`gofi ips add --<TAB>` lists flags. Flag *values* (target names, MAC addresses, sites)
don't complete — only the command and flag names themselves.

## The gofi commands

One binary, `gofi <area> <action>`:

| Area | What it manages |
|---|---|
| `ips` | Fixed IP + DNS reservations, in ISC DHCP host-declaration format |
| `dns` | Local DNS records, independent of `ips` |
| `network` | Networks (VLANs): subnet, DHCP pool, DNS servers (read-only) |
| `clients` | Currently-connected stations, with offline OUI vendor lookup |
| `users` | Known-client identity records, connected or not |
| `profile` | Capture networks + WLANs + fixed IPs as JSON, apply one back |
| `config` | gofi's own configuration file (acts on your machine, not a controller) |

```bash
gofi network list
gofi ips add --name nas --mac aa:bb:cc:dd:ee:01 --ip 192.168.1.13
gofi profile export > bench.json
```

**[`docs/gofi-user-guide.md`](docs/gofi-user-guide.md) is the full reference** — every
area, every action, every flag, with the reasoning behind how the command tree is
shaped.

---

# Part 2: Using gofi in your Go program (SDK)

The gofi module gives you type-safe, concurrent-safe access to UniFi Network Application
endpoints (v1, v2, REST, and WebSocket) — the same library the `gofi` CLI is built on.

```bash
go get github.com/unifi-go/gofi
```

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/unifi-go/gofi/src"
)

func main() {
    client, err := gofi.New(&gofi.Config{},
        gofi.WithAPIKey(os.Getenv("UNIFI_API_KEY")),
        gofi.WithConnector(os.Getenv("UNIFI_CONSOLE_ID")))
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect(ctx)

    devices, err := client.Devices().List(ctx, "default")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d devices\n", len(devices))
    for _, device := range devices {
        fmt.Printf("- %s (%s)\n", device.Name, device.Model)
    }
}
```

**[`docs/api-guide.md`](docs/api-guide.md) is the full SDK reference** — every service,
common operations for devices/networks/WLANs/clients/firewall/events, error handling,
testing with the mock server, retry/timeout options, and local-auth setup.

---

## Development

```bash
make test          # Run all tests
make coverage      # Generate coverage report
make lint          # Run linter
make build         # Build the module and utilities
make examples      # Build all examples to bin/examples/
make utilities     # Build all utilities to bin/
make install       # Install utilities to ~/bin
make all           # Run lint, test, and build
```

## Requirements

- Go 1.22 or later
- UniFi UDM Pro with Network Application 10.x+
- Admin access to the controller

## Compatibility

Tested with UniFi OS 4.x and 5.x, Network Application 10.x, on UDM Pro, UDM SE, and UDR.

## Documentation

- [gofi user guide](./docs/gofi-user-guide.md) — every CLI area, action, and flag
- [SDK / API guide](./docs/api-guide.md) — the full Go library reference and example programs
- [Alternate local API](./docs/alternate-local-api.md) — local username/password auth (secondary path)
- [Design](./docs/DESIGN.md) — architecture details
- [GoDoc](https://pkg.go.dev/github.com/unifi-go/gofi/src) — API reference

## Contributing

Contributions are welcome. Please ensure all tests pass (`make test`), code passes linting
(`make lint`), new features include tests, and changes maintain backward compatibility.

## License

MIT License.

## Acknowledgments

- Inspired by [paultyng/go-unifi](https://github.com/paultyng/go-unifi) (Terraform provider patterns)
- Type patterns from [unpoller/unifi](https://github.com/unpoller/unifi) (FlexInt/FlexBool)
- API patterns from [thib3113/unifi-client](https://github.com/thib3113/unifi-client) (TypeScript)
