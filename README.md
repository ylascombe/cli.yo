# cli.yo

Simple CLI made with golang and cobra that contains basic helpers commands for kubernetes (and more in future?)

For now, there is no CI/CD configured nor packaging model, it is pure POC

# Usage

```
make help
```

For example, kube commands:
```
go run ./main.go kube -h
These commands will use current context or help to set it.

Usage:
  cli.yo kube [flags]
  cli.yo kube [command]

Available Commands:
  debugHost   Run a debug pod on specific host in host mode
  debugPod    Run a debug-pod

Flags:
  -h, --help   help for kube

Use "cli.yo kube [command] --help" for more information about a command.
```
