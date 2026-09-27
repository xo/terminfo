# Backlog

This file holds the work that is known and not done. Each item names where it
came from. The repository has issues turned off on GitHub, no open pull
request, and no `TODO` comment in the code, so every item came from a pull
request, the code or the work of 2026-09-27.

When an item is done, delete it. If the work needed a decision, write the
decision in [PLAN.md](PLAN.md). An item that needs Ken to choose is also an
open question at the end of [PLAN.md](PLAN.md).

## Faults

### B1. go generate fails, because go.mod does not require snaker

Source: found on 2026-09-27, by running `go build gen.go`.

`gen.go` imports `github.com/kenshaw/snaker`. `c203b35` removed the last
`require` line, and `go.mod` never required `snaker`, so the go command
cannot build `gen.go`. `capvals.go` cannot be written again until this is
fixed. The fix depends on open question 1.

### B2. The ncurses release that wrote capvals.go is not recorded

Source: found on 2026-09-27, by comparing `capvals.go` with `.cache`.

Without `-ver`, `gen.go` takes the newest release on `ftp.gnu.org`, so each
run can write a different file. The comments that `7ff4e87` wrote match
neither `ncurses` 6.2 nor 6.3, which are the two releases in `.cache`.
`capvals.go` does not say which release wrote it. See open question 2.

### B3. %c prints a NUL byte for an int parameter

Source: found on 2026-09-27, with a test that was then removed.

`Printf([]byte("%p1%c"), 65)` returns `"\x00"`, and not `"A"`. `popByte`
returns 0 for any value that is not a `byte`, and a caller usually passes an
`int`. `%s` has the same problem: `popString` returns `""` for any value that
is not a `string`. `ncurses` converts the value. The formatted forms, such as
`%3c`, use the same functions.

### B4. Open panics on an empty name

Source: found on 2026-09-27, with a test that was then removed.

`Open(dir, "")` slices `name[0:1]` and panics. `Load` refuses an empty name
with `ErrEmptyTermName`, but `Open` does not.

### B5. Load returns ErrDatabaseDirectoryNotFound when the entry is missing

Source: found on 2026-09-27, in `load.go`.

`Load("no-such-term")` returns `ErrDatabaseDirectoryNotFound`, while the
directories exist. The error names the wrong fault. `ErrFileNotFound`
exists and is the one that `Open` returns.

### B6. Load stops when user.Current fails

Source: found on 2026-09-27, in `load.go`.

`Load` calls `user.Current` to find `$HOME/.terminfo`. If that call fails,
`Load` returns the error at once. It does not try `$TERMINFO`, which it
already added to the list, or the system directories.

### B7. ColorLevelFromEnv returns ColorLevelNone for 8 or 16 colors

Source: found on 2026-09-27, in `color.go`.

When no environment variable decides, `ColorLevelFromEnv` reads
`max_colors`. A value of 16 or less returns `ColorLevelNone`, a value of 256
or more returns `ColorLevelHundreds`, and a value from 17 to 255 returns
`ColorLevelBasic`. A terminal with 8 colors gets no color. See open
question 3.

### B8. Terminfo.Strings holds a nil value for each absent capability

Source: found on 2026-09-27, with a test that was then removed.

For `xterm-256color`, `Strings` has 413 entries and 230 of them are nil.
`readStrings` stores every index, present or not. `StringCaps` and
`StringCapsShort` then return a key for each absent capability.
`cmd/infocmp` skips the nil values, which hides the problem there.

### B9. Decode refuses an entry of exactly 32768 bytes

Source: found on 2026-09-27, in `dec.go`.

`Decode` refuses a buffer when `len(buf) >= maxFileLength`. D15 says that
32768 bytes is the limit, so the test can be off by one. `ncurses` also keeps
4096 as the limit for the legacy magic number. Compare with
`_nc_read_termtype` in `ncurses/tinfo/read_entry.c` before you change it.

### B10. A caller can change a cached entry for every other caller

Source: found on 2026-09-27, in `load.go` and `terminfo.go`.

`Load` and `Open` return the same `*Terminfo` from the cache to every caller.
Its maps are exported, so a caller that writes to one changes it for all
the others.

### B11. The example does not pass go vet

Source: found on 2026-09-27, by running `go vet ./_examples/`.

`termputs` takes a format string, and `main` passes it the output of
`Colorf`, which is not a constant. `go vet ./...` skips `_examples` (D12),
so CI does not see it. In `termtitle`, `var once sync.Once` is declared in
the function, so the `Once` runs on every call.

### B12. cmd/infocmp/test.sh and comp.sh name a path on one machine

Source: found on 2026-09-27, in `cmd/infocmp`.

Both scripts run `$HOME/src/misc/icdiff/icdiff`. The scripts fail on any
other machine at the last step.

## Tests

### B13. TestValues does not check that a cancelled number is missing

Source: found on 2026-09-27, in `terminfo_test.go`.

The check for a cancelled number is a comment. The same test checks
cancelled bools and strings. The test also runs `/usr/bin/infocmp` by its
full path, so it fails on a machine that has `infocmp` elsewhere.

### B14. TestLoad sets TERM with os.Setenv

Source: found on 2026-09-27, in `terminfo_test.go`.

`TestLoad` calls `os.Setenv("TERM", ...)` and does not restore the value.
`t.Setenv` restores it when the test ends.

## Cleanup

### B15. The code uses io/ioutil and path

Source: found on 2026-09-27, in `terminfo.go` and `load.go`.

`Open` calls `ioutil.ReadFile`, which is deprecated, and not `os.ReadFile`.
`Open` and `Load` join paths on disk with `path.Join`, and not with
`path/filepath`.

### B16. terminfo.go holds a Puts func in a comment

Source: found on 2026-09-27, in `terminfo.go`.

`Puts` expands padding such as `$<5>`. It came from the 2016 code and has
been a comment since the rewrite in `24dccac`. Either finish it or delete
it.

### B17. The generated-code line in capvals.go comes after the package clause

Source: found on 2026-09-27, in `gen.go` and `capvals.go`.

The Go convention puts `// Code generated ... DO NOT EDIT.` before the first
text that is not a comment. `gen.go` writes it after `package terminfo`, so a
tool that follows the convention does not treat `capvals.go` as generated.

### B18. Some comments are wrong

Source: found on 2026-09-27, while reading the code for `AGENTS.md`.

- The doc comment of `Colorf` starts with "Color takes".
- `pushParamFn` says "skip the '}'" where it skips the parameter digit.
- `dec.go` names `repair_ascc`, and the `ncurses` function is `repair_acsc`.
- The comment on `parametizer` says "represents the a scan state".
- The comments on `ExtBoolNames`, `ExtNumNames` and `ExtStringNames` name
  `ExtBoolsNames`, `ExtNumsNames` and `ExtStringsNames`.
- `canonicalizeAscChars` and the end of `Decode` hold lines of code in
  comments.

### B19. CI runs no linter

Source: open question 4 in [PLAN.md](PLAN.md).

The repository has no `.golangci.yml`, and CI does not run `golangci-lint`.
dburl and tblfmt have both.

### B20. The README badge points at godoc.org

Source: found on 2026-09-27, in `README.md`.

godoc.org sends each request to pkg.go.dev now. The README example also
names `_examples/simple/main.go` in its first line, and `abceb7e` moved the
file to `_examples/main.go`. Pull request 2 said in 2021 that the example is
broken. The example builds, but see B11.
