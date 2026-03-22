# cobra-cli-template

A template for building Go CLIs with Cobra using a test-first workflow.

## Stack

- CLI framework: `spf13/cobra`
- Test runner: `onsi/ginkgo`
- Assertions: `stretchr/testify/assert`
- Utility: `samber/lo`
- Validation: `kaptinlin/gozod`
- Mode Management: `louiss0/g-tools/mode` 

## Create a Project

```sh
git clone https://github.com/louiss0/cobra-cli-template .
```

or

```sh
gh repo create <project-name> --template louiss0/cobra-cli-template --public --clone
```

## Run Tests

```sh
ginkgo run ./...
```

```sh
ginkgo watch ./...
```

## Rename Template Project

```sh
go run ./scripts/rename-project
```

The script infers the new project name from the folder you run it in.

```sh
go run ./scripts/rename-project -project-name my-cli
```

`-module-path` is optional. If omitted, it infers your current module from
`go.mod` and swaps only the last path segment to `<project-name>`.

## Add Commands

1. Add a file in `cmd/`.
2. Create a `New<CommandName>() *cobra.Command` constructor.
3. Add flags in that constructor.
4. Register the command from `NewRootCmd()`.
5. Add tests in the same package.

## Notes

- Keep command behavior in `RunE` and return explicit errors.
- This template includes `custom_flags` for `cmd.Flags().Var(...)` use cases.
- `custom_errors` centralizes reusable error values.
