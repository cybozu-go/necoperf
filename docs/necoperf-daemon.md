# necoperf-daemon command reference

```console
necoperf-daemon <subcommand> args...
```

- [`necoperf-daemon start`](#necoperf-daemon-start)

## `necoperf-daemon start`

Start necoperf-daemon on the server.

| Option | Default value |Description |
|:-------|:--------------|:-----------|
| `--port` | `6543` | Port number on which the grpc server runs |
| `--metrics-port` | `6541` | Port number on which the metrics server runs |
| `--runtime-endpoint` | `unix:///run/containerd/containerd.sock` | Container runtime endpoint to connect to |
| `--work-dir` | `/var/necoperf` | Directory for storing profiling results |
