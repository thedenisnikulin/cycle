# cycle

simple go script to toggle / increment / decrement whatever comes in on stdin

```sh
echo true | cycle                # false
echo 41 | cycle                  # 42
echo 2026-09-11 | cycle --prev   # 2026-09-10
echo 0x0f | cycle                # 0x10
echo v1.2.9 | cycle              # v1.2.10
echo Monday | cycle --prev       # Sunday
echo '[ ]' | cycle               # [x] (markdown)
```

## install

```sh
go install github.com/thedenisnikulin/cycle@latest
```

## use with helix

```toml
[keys.normal]
"C-a" = ":pipe cycle"
"C-x" = ":pipe cycle --prev"

[keys.select]
"C-a" = ":pipe cycle"
"C-x" = ":pipe cycle --prev"
```

overrides default incr/decr keybinds. Select text and hit C-a/C-x

## inspiration

- [dial.nvim](https://github.com/monaqa/dial.nvim)
- [boole.nvim](https://github.com/nat-418/boole.nvim)
