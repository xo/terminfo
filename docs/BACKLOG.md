# Backlog

This file holds the work that is known and not done. Each item names where it
came from. The repository has issues turned off on GitHub, no open pull
request, and no `TODO` comment in the code, so every item came from a pull
request, the code or the work of 2026-09-27.

When an item is done, delete it. If the work needed a decision, write the
decision in [PLAN.md](PLAN.md). An item that needs Ken to choose is also an
open question at the end of [PLAN.md](PLAN.md).

## Faults

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

### B9. Decode refuses an entry of exactly 32768 bytes

Source: found on 2026-09-27, in `dec.go`.

`Decode` refuses a buffer when `len(buf) >= maxFileLength`. D15 says that
32768 bytes is the limit, so the test can be off by one. `ncurses` also keeps
4096 as the limit for the legacy magic number. Compare with
`_nc_read_termtype` in `ncurses/tinfo/read_entry.c` before you change it.

### B12. cmd/infocmp/test.sh and comp.sh name a path on one machine

Source: found on 2026-09-27, in `cmd/infocmp`.

Both scripts run `$HOME/src/misc/icdiff/icdiff`. The scripts fail on any
other machine at the last step.

## Tests

### B13. TestValues does not check cancelled numbers and strings

Source: found on 2026-09-27, in `terminfo_test.go`.

The checks for a cancelled number and a cancelled string were comments. The
lint fixes for D25 removed them. The test checks cancelled bools only. The
parser of the `infocmp -E` output gives a nil value both for an absent
capability and for a cancelled one. `NumsM` and `StringsM` hold only the
cancelled ones, so the old checks fail on every entry. A fix must tell the
two apart in the parser first. The test also runs `/usr/bin/infocmp` by its
full path, so it fails on a machine that has `infocmp` elsewhere.

