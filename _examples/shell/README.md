# Shell example

This is an example adapted from
<https://github.com/alecthomas/kong/blob/master/_examples/shell/main.go>. It
adds completions to `rm` and `ls` commands.

Try completions using the `COMP_LINE` environment variable. For example:

```shell
 $ COMP_LINE="shell " go run .
rm
ls
install-completions

$ COMP_LINE="shell -" go run .
--help
--debug
```
