# Gargle

*This is a work in progress and offers no compatibility guarantees.*

## Motivation

Gargle aims to provide a flexible command line parser while avoiding the
complexity of deep integration.

What differentiates Gargle?

- No special casing required for common flags and commands, such as `--help`.
- Code-generated usage is more flexible, lighter weight, and often more readable
  than templates.
- Minimal side-effects help separate declaration from execution.
- Light-weight and isolated.

## Examples

### Adding Help

Any action can be written to write application usage. Following is an example
which illustrates adding help using the default usage writer.

```golang
cmd := &Command{/*...*/}
cmd.AddFlags(gargle.NewHelpFlag(nil))
```

The value `nil` could be replaced by any `Action` to customize usage.

### Negative Booleans

Some applications prefer to provide negated versions of boolean flags. This
example shows how to create a flag which can be explicitly enabled or disabled
via `--enable-foo` and `--disable-foo`.

```golang
var foo bool
cmd := &Command{/*...*/}
cmd.AddFlags(
    &gargle.Flag{Name: "enable-foo", Value: gargle.BoolVar(&foo)},
    &gargle.Flag{Name: "disable-foo", Hidden: true, Value: gargle.NegatedBoolVar(&foo)},
)
```

### Environment Defaults

It's often convenient to accept environment variables in place of flags. This
example shows how to wrap a value with an environment-backed default value.

```golang
func EnvDefault(v gargle.Value, key string) gargle.Value {
    if s, ok := os.LookupEnv(key); ok {
        return gargle.WithDefault(v, s)
    }
    return v
}
```

```golang
setting := "<none>"
Flag(Name: "setting", EnvDefault(gargle.StringVar(&setting), "APP_SETTING")
```

This sets `value` in order of precedence to:

1. The flag `--setting` if provided.
1. The environment variable `APP_SETTING` if provided.
1. The value `"<none>"` if none of the above.

## Why "Gargle"?

The Go ecosystem is rife with puns. In short, GoArgParse -> GArg -> Gargle.

## Alternatives

- [flag](https://golang.org/pkg/flag/):
  - Fast, simple, and built in
  - Doesn't directly support sub-commands
  - Difficult to customize
- [Cobra](https://github.com/spf13/cobra):
  - Extremely customizable
  - Complex interface, provides a code generator
  - Integrates with [Viper](https://github.com/spf13/cobra)
  - Template-based usage formatting
- [Kingpin](https://github.com/alecthomas/kingpin)
  - Fluent-style interface
  - Template-based usage formatting
