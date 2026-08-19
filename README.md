# zxsubs

A small Go CLI that opens a full-terminal picker for the Azure subscriptions already available in your Azure CLI session.

`zxsubs` does not log you in. Sign in with `az login` first; this tool only lists and selects (`az account list` / `az account set`).

## Prerequisites

- [Go](https://go.dev/dl/) 1.22 or later (to install or build)
- [Azure CLI](https://learn.microsoft.com/cli/azure/install-azure-cli) (`az`) on your `PATH`
- An interactive terminal
- An existing Azure CLI login:

```bash
az login
```

## Install

```bash
go install github.com/hdisk13/zxsubs@latest
```

Or from a clone:

```bash
git clone https://github.com/hdisk13/zxsubs.git
cd zxsubs
go build -o zxsubs .
```

## Usage

```bash
zxsubs
```

The picker fills the terminal:

| Key | Action |
| --- | --- |
| `↑` / `↓` | Move the highlight |
| type | Fuzzy-filter the list (name, id, state, tenant) |
| `Backspace` | Delete the last filter character |
| `Enter` | Set that subscription as the active Azure CLI subscription |
| `q` / `Esc` / `Ctrl+C` | Cancel — nothing is changed |

The currently active subscription is labeled **current**. Each row also shows a short id suffix, state, and tenant suffix so similarly named subscriptions stay distinguishable.

On a successful pick, `zxsubs` runs `az account set --subscription <id>` and prints:

```
Active subscription: Contoso Development (22222222-2222-2222-2222-222222222222)
```

### Exit codes

| Code | Meaning |
| --- | --- |
| `0` | A subscription was selected and set |
| `1` | Canceled, `az` missing, not logged in, or another error |

Canceling never calls `az account set`. Scripts can use `zxsubs && …` so later commands only run after a real selection.

### Errors

If `az` is not installed or you are not logged in, `zxsubs` exits `1` and tells you what to run (`az login`, or install Azure CLI).

## Development

```bash
go test ./...
go build -o zxsubs .
```

Azure CLI is not required for tests. Listing and JSON parsing use a fixture of `az account list` output; the picker model is driven with key events instead of a live terminal.
