# snipe

Find and kill processes by port. That's it.

```
snipe 3000
```

```
──────────────────────────────────────────────────
  kill  :3000  pid 48291
  cmd   node server.js
  path  /home/you/project/server.js
──────────────────────────────────────────────────
```

## Install

**Homebrew:**

```
brew install onurogut/tap/snipe
```

**Go:**

```
go install github.com/onurogut/snipe@latest
```

**Binary:** grab one from [releases](https://github.com/onurogut/snipe/releases).

**From source:**

```
git clone https://github.com/onurogut/snipe.git
cd snipe
make install
```

## Usage

```
snipe 3000                  # kill whatever's on :3000
snipe 3000 8080 4200        # multiple ports
snipe 3000-3005             # port range
```

### Flags

```
-d    dry run, just show what would die
-l    list processes without killing
-i    ask before each kill
-q    quiet, exit code only
-g    try SIGTERM first, SIGKILL after 2s
-v    version
```

### Examples

See what's running without killing anything:

```
snipe -l 3000
```

Check before you wreck:

```
snipe -d 3000 8080
```

Be polite about it (SIGTERM, then SIGKILL if needed):

```
snipe -g 3000
```

Use in scripts:

```
snipe -q 3000 && echo "cleared" || echo "nothing there"
```

## How it works

`lsof` to find the pid, `kill` to end it, `ps` to show you what died. On Linux it falls back to `ss` or `/proc/net/tcp` if `lsof` isn't around.

## License

MIT
