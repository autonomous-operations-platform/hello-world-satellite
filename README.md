> ! TODO !
> Before merging, I would like to have a pipeline that generates and pushes a new OCM bundle to this repo whenever a tag is created on main or something like that
> People can then pull from the repo
[![CodeQL Advanced](https://github.com/autonomous-operations-platform/hello-world-satellite/actions/workflows/codeql.yml/badge.svg)](https://github.com/autonomous-operations-platform/hello-world-satellite/actions/workflows/codeql.yml) [![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/autonomous-operations-platform/hello-world-satellite/badge)](https://scorecard.dev/viewer/?uri=github.com/autonomous-operations-platform/hello-world-satellite) [![REUSE status](https://api.reuse.software/badge/github.com/autonomous-operations-platform/hello-world-satellite)](https://api.reuse.software/info/github.com/autonomous-operations-platform/hello-world-satellite)

# Hello World Satellite

A minimal AOS satellite demonstrating the collector → observer → operator pattern using a CPU temperature monitor and fan speed controller.

| Component | Name | Role |
|---|---|---|
| Collector | `temperature-collector` | Mocks a CPU temperature sensor (oscillates between 30–95°C) and emits telemetry every 5s with `source`, `value`, and `unit` attributes. |
| Observer | `temperature-threshold` | Watches `cpu-temp` telemetry. Emits a `temperature-high` insight above 75°C and a `temperature-low` insight below 60°C. The band between the two thresholds prevents oscillation. |
| Operator | `fan-speed-controller` | Reacts to temperature insights: sets fan speed to 80% on `temperature-high`, and back to 20% on `temperature-low`. Fails with a simulated transient error 30% of the time. |

New to AOS? The [Getting Started guide](./GETTING-STARTED.md) builds this satellite by hand, step by step.

## Local Setup

### 1. Install Go

Install Go 1.27.1 from [go.dev/dl](https://go.dev/dl/).

### 2. Set up the Go workspace

Copy the example workspace file and adjust the `replace` path to point to your local `aos-framework` clone:

```bash
cp go.work.example go.work
```

The `go.work` file is gitignored. It redirects the framework import to your local checkout for all build and test commands.

### 4. Fetch and vendor dependencies

In workspace mode, use `go get` to add or update dependencies (it respects the `go.work` replace), then vendor:

```bash
go get github.com/autonomous-operations/aos-framework@latest
go work vendor
```

> Note: `go mod tidy` does not respect `go.work` replace directives and will fail when pointing to a local framework checkout. Use `go get` instead.

### 5. Build

```bash
go build .
```

### 6. Configure and run

Then run:

```bash
./hello-world-satellite
```


## Support, Feedback, Contributing

This project is open to feature requests/suggestions, bug reports etc. via [GitHub issues](https://github.com/autonomous-operations-platform/hello-world-satellite/issues). Contribution and feedback are encouraged and always welcome. For more information about how to contribute, the project structure, as well as additional contribution information, see our [Contribution Guidelines](https://github.com/autonomous-operations-platform/.github/blob/main/CONTRIBUTING.md).

## Security / Disclosure

If you find any bug that may be a security problem, please follow our instructions at [in our security policy](https://github.com/autonomous-operations-platform/.github/blob/main/SECURITY.md) on how to report it. Please do not create GitHub issues for security-related doubts or problems.

## Code of Conduct

We as members, contributors, and leaders pledge to make participation in our community a harassment-free experience for everyone. By participating in this project, you agree to abide by its [Code of Conduct](https://github.com/autonomous-operations-platform/.github/blob/main/CODE_OF_CONDUCT.md) at all times.

## Licensing

Copyright 2026 SAP SE or an SAP affiliate company and Autonomous Operations Platform contributors. Please see our [LICENSE](./LICENSE) for copyright and license information. Detailed information including third-party components and their licensing/copyright information is available [via the REUSE tool](https://api.reuse.software/info/github.com/autonomous-operations-platform/hello-world-satellite).
