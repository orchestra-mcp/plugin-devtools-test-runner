# Orchestra Plugin: devtools-test-runner

A tools plugin for the [Orchestra MCP](https://github.com/orchestra-mcp/framework) framework.

## Install

```bash
go install github.com/orchestra-mcp/plugin-devtools-test-runner/cmd@latest
```

## Usage

Add to your `plugins.yaml`:

```yaml
- id: tools.devtools-test-runner
  binary: ./bin/devtools-test-runner
  enabled: true
```

## Tools

| Tool | Description |
|------|-------------|
| `hello` | Say hello to someone |

## Related Packages

- [sdk-go](https://github.com/orchestra-mcp/sdk-go) — Plugin SDK
- [gen-go](https://github.com/orchestra-mcp/gen-go) — Generated Protobuf types
