# kongplete

Package kongplete lets you generate shell completions for your command-line programs using
github.com/alecthomas/kong and github.com/posener/complete.

#### Examples

##### Complete

Complete runs completion for a kong parser

```golang

var cli struct {
    Foo struct {
    } `kong:"cmd"`
}

_ = cli

```
