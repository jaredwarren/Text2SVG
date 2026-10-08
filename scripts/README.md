# Scripts

Ad-hoc geometry / offset debugging programs used during development.

Each file is tagged with `//go:build ignore` so it is excluded from `go build ./...` and the test suite. Run an individual script with:

```bash
go run ./scripts/<name>.go
```

Outputs (SVG, PNG, DXF) should stay out of git; they are covered by the root `.gitignore`.
