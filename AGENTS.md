# Agents 

This project is a Cobra CLI application.
It's created to allow developer to deploy an executable using GoReleaser.

This project contains a few important packages:
- `cmd` the place where all commands go 
- `custom_errors` the place where all errors go 
- `custom_flags` the place where custom flags go 
- `output` the place where all CLI output manipulation goes
- `scripts` the place where all scripts go 
- `templates` the place where all templates typically used by Ginkgo go
- `validation` the place where all validation go. 

When you need a local executable build.
Use `go build . -ldflags "-X github.com/louiss0/g-tools/mode.buildMode=production"`
This makes sure that production builds use features that aren't tested. 

## Testing 


The assert function is setup in every test `*_suite_test.go` file.
This means you only need to use it in your test files. It's name is `assert`.

When creating a new test suite file, prefer `ginkgo bootstrap <suite_name> --template templates/testify-suite.txt`.
When creating a new test file, prefer `ginkgo generate <test_name> --template templates/testify-test.txt`.
If you create the files manually, match the structure provided by those templates.

You are supposed to always write Ginkgo code when writing tests! Don't write Go testing code!
Please prefer to run `ginkgo run <package>` over the normal go test command.

When doing coverage use this command 

```sh
ginkgo -r --cover --output-dir coverage --keep-separate-coverprofiles
```

## How to develop commands! 

In order for this template to work, you need to make Cobra commands that are functions that return `cobra.Command`.
They must look like the structure below. `<Command_Name>` is the name of the command you want to create.
Derive it from the file name without the `.go` extension.
If the file name contains an underscore, use the segment before the first underscore.

```go
func New<Command_Name>() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "",
		Short: "",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			
			return nil
		},
	}
	
	return cmd
}
```

When making flags make sure to place them in the function that returns the `cobra.Command`.
Declare them on the local `cmd` before returning it.
Do not place non-global flags on the `rootCmd` object or attach them from outside the `New<Command_Name>` function.

In the `RunE` function, return `nil` if the command succeeds and an error if it fails.
This WriteModeAwareOutput is in `cmd/pkg.go` use it to return messages! 
It's job is to allow output during production, but capture it in development! 


When validating command arguments always place the logic in the `Args:` function.
If an argument needs to look a specific way then just pass a function and use the built in validation
along side something new!
Then you should store the arguments in an annonymus struct using the `Run` or `RunE` function.

```go

var commandArgs struct {
	Package string
}

cmd := &cobra.Command{
		Args: func(cmd *cobra.Command, args []string) error {
		
		err := cobra.MaximumNArgs(1)(cmd, args)
		
		if err != nil {
			return err
		}
		
		if len(args) > 0 {
			
			if args[0] != strings.ToLower(args[0]) {
				return fmt.Errorf("argument '%s' must be all lowercase", args[0])
			}
		}
		
		return nil

		},
		RunE: func(cmd *cobra.Command, args []string) error {
			
			commandArgs.Package = args[0]
			
			return nil
			
		}
	}
```

When making sub commands always place them in the command creation function! 
Use `cmd.AddCommand(command)`. 

When validating flags please look at the custom flags package first! 
If the code can't be used then do your own flag validation.

When it comes to sub commands use the `AddGroup` function.
It's a function that makes sure that commands are strictly associated with a specific purpose.

You set it up like this in the root command. 

```go

cmd := &cobra.Command{}

cmd.AddGroup(&cobra.Group{ID: "manage", Title: "Management Commands"})
```

Then you assign the id's like this

```go
var backupCmd = &cobra.Command{
	Use: "backup", 
	Short: "Create a backup",
	GroupID: "manage"
}

```

If I ask you to generate markdown please consult this page <https://cobra.dev/docs/how-to-guides/clis-for-llms/>.

When it comes to flags remember these things below

To make sure two flags must be sent together use `cmd.MarkFlagsRequiredTogether()`
To make sure make sure that a specific flag must be used alone use `cmd.MarkFlagsMutuallyExclusive()`
For other flag related constraints find a function prefixed with Mark! It should contain what you need.

When it comes to passing values from one command to sub-commands I like to use the `PersistentPreRunE` function.
The point of this function is to make sure that values are valid. 
It's used to set flag values that will typically be used in subcommands.
The one in the root command is used to setup and set the context! 
The `GenerateContextFromMap` is supposed to be used to setup the dependencies that will be fetched from the context.
make sure this function is only used once after dependencys are generated with values. 
This function will return the new context! Make sure to use `cmd.SetContext()` at the least immediate return. 
Always make sure `SetContext` is used at the outer most final return!

```go
&cobra.Command{
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		ctx := GenerateContextFromMap(cmd, map[string]any{})
		cmd.SetContext(ctx)
		return nil
		
				
	},
}
```

If you decide to set things in the context you need to make helper functions to get things from the context!
These kinds of functions are Command Context Fetchers! They take the `cobra.Command` as the only parameter. 
Then they get a value from the context using the key that was used to get the context.
They are named `get<dependency>FromCommandContext` `<dependency>` is the name of the dependency that is getting fetched.
They should use a checked type assertion and return a descriptive error when the dependency is missing or invalid.

Example Command Context Fetcher
```go
func getGoEnvFromCommandContext(cmd *cobra.Command) (env.GoEnv, error) {
	goEnv, ok := cmd.Context().Value(_GO_ENV).(env.GoEnv)
	if !ok {
		return env.GoEnv{}, fmt.Errorf("missing %s in command context", _GO_ENV)
	}
	return goEnv, nil
}
```

The Root command is written in a way where it needs a dependencies struct.
Each prop needs a function to be passed that will instanciate a struct or return a function!
Make sure that the the dependencies are functions that 
- Return a struct 
- Return functions that only need values not other structs!

## Validation

Use goZod for runtime validation where Cobra dependency wiring or command input
can drift from compile-time intent. If the validation is specific to a series of code.
**inline it**. If validation is generic please place it in the `validation` package.

- Use `gozod.FromStruct[T]()` when validating typed dependency/config structs.
- Use the `validation.NewFunctionValuesStructSchema[T]()` helper when a struct
  is expected to only contain function fields and all functions must be set.
- Prefer schema-level `.Check(...)` for cross-field rules and for building
  messages that list missing dependencies.
- Keep per-command argument validation in Cobra `Args` functions and reserve
  goZod for internal runtime contracts and dependency shape checks.

## Using Utilities 

Make sure use consider using `samber/lo` when working with data structures.
When it comes to creating heavy strings do the best practice!
Please use utilites for making new values rather than using loops!
