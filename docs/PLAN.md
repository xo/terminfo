# Plan

This file holds the plan for `terminfo`, then its decisions, then the open
questions for Ken. `terminfo` is a small library, so its decisions stay in
this file and do not get one file each (dbmeta D111).

## Purpose

`terminfo` reads the terminfo database in pure Go. A Go program uses it to
find what a terminal can do and to get the byte sequences that make the
terminal do it. A program that uses it does not need cgo or the `ncurses`
library.

## Current state

The package decodes the legacy file format and the 32-bit number format, with
the extended capabilities of `ncurses`. It evaluates parameterized strings. It
finds a color level from the environment. `cmd/infocmp` prints an entry as
`infocmp -1 -L` does.

The tests compare each decoded value with the output of the system `infocmp`
for every entry on the machine (D3). They pass on the `ncurses` of Ubuntu in
GitHub Actions (D17) and on `ncurses` 6.6.

The work in September 2026 fixed the parameter scanner, the `acs_chars`
order, local variables, and the reading of large entries and cancelled
extended strings. The faults that are still open are in
[BACKLOG.md](BACKLOG.md).

## Direction

The file format has not changed in decades, so the package does not grow new
features (D14). The work is to fix faults, to keep up with the `ncurses`
releases that the tests run against, and to keep the code idiomatic Go. A
higher level terminal library belongs in another package.

## Decisions

Each decision that shapes `terminfo`, in the order that it was made. The
entries are append only. A decision is never edited to change its conclusion.
When a later decision changes one, the later heading says "Amends Dn", and
the earlier heading says "Amended by Dn".

Each heading carries a status after the title. "Decided" means that the
history shows that Ken chose it. "Proposed" means that the history does not
show it, and Ken must answer. D1 to D20 were written on 2026-09-27 from the
git history, the pull requests and the code. Each names the commits that it
comes from. If no commit message, comment or pull request gives the reason,
the entry says so.

| Decision | Title | Status |
| --- | --- | --- |
| [D1](#d1-terminfo-is-a-rewrite-of-nhooyrterminfo-in-idiomatic-go-decided) | terminfo is a rewrite of nhooyr/terminfo in idiomatic Go | Decided |
| [D2](#d2-the-package-reads-the-database-in-pure-go-and-does-not-link-ncurses-decided) | The package reads the database in pure Go and does not link ncurses | Decided |
| [D3](#d3-the-system-infocmp-is-the-reference-for-every-decoded-value-decided) | The system infocmp is the reference for every decoded value | Decided |
| [D4](#d4-the-capability-constants-are-generated-from-the-ncurses-caps-file-decided) | The capability constants are generated from the ncurses Caps file | Decided |
| [D5](#d5-load-searches-the-directories-that-terminfo5-names-and-caches-each-entry-decided) | Load searches the directories that terminfo(5) names, and caches each entry | Decided |
| [D6](#d6-tests-run-in-travis-ci-decided-amended-by-d17) | Tests run in Travis CI | Decided. Amended by D17 |
| [D7](#d7-the-decoder-refuses-an-entry-of-4096-bytes-or-more-decided-amended-by-d15) | The decoder refuses an entry of 4096 bytes or more | Decided. Amended by D15 |
| [D8](#d8-the-acs_chars-value-is-ordered-by-source-character-as-infocmp-prints-it-decided) | The acs_chars value is ordered by source character, as infocmp prints it | Decided |
| [D9](#d9-printf-takes-the-capability-as-bytes-and-returns-a-string-decided) | Printf takes the capability as bytes and returns a string | Decided |
| [D10](#d10-colorlevelfromenv-reads-the-environment-before-the-database-decided) | ColorLevelFromEnv reads the environment before the database | Decided |
| [D11](#d11-a-sentinel-error-is-a-constant-of-type-error-decided) | A sentinel error is a constant of type Error | Decided |
| [D12](#d12-examples-are-in-_examples-which-the-go-command-skips-decided) | Examples are in _examples, which the go command skips | Decided |
| [D13](#d13-the-decoder-reads-the-32-bit-number-format-decided) | The decoder reads the 32-bit number format | Decided |
| [D14](#d14-terminfo-does-not-grow-new-features-decided) | terminfo does not grow new features | Decided |
| [D15](#d15-the-decoder-accepts-an-entry-of-up-to-32768-bytes-as-ncurses-61-does-decided-amends-d7) | The decoder accepts an entry of up to 32768 bytes, as ncurses 6.1 does | Decided. Amends D7 |
| [D16](#d16-a-lowercase-variable-is-local-to-one-evaluation-and-an-uppercase-variable-is-shared-decided) | A lowercase variable is local to one evaluation, and an uppercase variable is shared | Decided |
| [D17](#d17-tests-run-in-github-actions-on-the-ncurses-that-ubuntu-ships-decided-amends-d6) | Tests run in GitHub Actions, on the ncurses that Ubuntu ships | Decided. Amends D6 |
| [D18](#d18-the-module-declares-a-recent-go-version-decided) | The module declares a recent Go version | Decided |
| [D19](#d19-the-tests-accept-the-infocmp-output-of-ncurses-before-65-decided) | The tests accept the infocmp output of ncurses before 6.5 | Decided |
| [D20](#d20-the-module-requires-no-other-module-proposed) | The module requires no other module | Proposed |
| [D21](#d21-terminfo-makes-no-promise-of-backward-compatibility-decided) | terminfo makes no promise of backward compatibility | Decided |
| [D22](#d22-terminfo-is-set-up-for-coding-agents-as-every-xo-repository-is-decided) | terminfo is set up for coding agents as every xo repository is | Decided |

### D1. terminfo is a rewrite of nhooyr/terminfo in idiomatic Go. Decided.

Date: 2017-09-23. Commit: `24dccac`.

Anmol Sethi wrote `github.com/nhooyr/terminfo` in 2016. Ken rewrote it as
`github.com/xo/terminfo`. The commit message gives four aims:

1. Rewrite the code to be idiomatic Go.
2. Fix the capability names.
3. Add `cmd/infocmp`, which aims to be compatible with the `infocmp`
   utility.
4. Fix the decoding errors.

Ken said on pull request 2, in 2021, that he forked the package from a
project that nobody maintained, and that he cleaned it up.

### D2. The package reads the database in pure Go and does not link ncurses. Decided.

Date: 2017-09-23. Commits: `24dccac`, `b44407a`.

The package doc comment says that the package reads terminfo files in pure
Go. Anmol Sethi wrote it in 2016, and Ken kept it in the rewrite. In
`b44407a`, on 2018-01-06, Ken wrote in the README that the package is a pure
Go implementation, and that it is meant as a replacement for `ncurses` in
simple Go programs. The package has no cgo.

### D3. The system infocmp is the reference for every decoded value. Decided.

Date: 2017-09-23. Commits: `24dccac`, `52096cc`, `7ff4e87`.

`cmd/infocmp` aims to print what `infocmp -1 -L` prints (D1).
`cmd/infocmp/test.sh` compares the two for every entry in
`/usr/share/terminfo`. `52096cc` added tests "to validate actual terminfo
data". `TestValues` runs `/usr/bin/infocmp -E -x` on each entry on the
machine, and compares the names and every bool, number and string value with
the decoded ones. CI installs `ncurses-bin` and `ncurses-term` for this test.

If the package and `infocmp` disagree, the package is wrong.

### D4. The capability constants are generated from the ncurses Caps file. Decided.

Date: 2017-09-23. Commits: `24dccac`, `abceb7e`.

`gen.go` downloads an `ncurses` release from `ftp.gnu.org`, reads
`include/Caps`, and writes `capvals.go`. The file holds one constant for each
capability, with a comment from the `Caps` file, and the tables of long and
short names. `terminfo.go` runs it with `go:generate`. Nobody edits
`capvals.go` by hand. Without the `-ver` flag, `gen.go` takes the newest
release. It keeps each download in `.cache`, which git ignores.

### D5. Load searches the directories that terminfo(5) names, and caches each entry. Decided.

Date: 2017-09-23. Commits: `24dccac`, `198e0b2`.

`Load` searches `$TERMINFO`, then `$HOME/.terminfo`, then each directory in
`$TERMINFO_DIRS`, then `/etc/terminfo`, `/lib/terminfo` and
`/usr/share/terminfo`. Its doc comment names terminfo(5) as the source of
this order. Charith Ellawala added `/etc/terminfo` and `/lib/terminfo` in
2016, from `/etc/terminfo/README`, and Ken kept them in the rewrite.

`Open` puts each entry that it reads into a cache, under each of its names.
`Load` answers from that cache before it searches.

### D6. Tests run in Travis CI. Decided. Amended by D17.

Date: 2017-09-23. Commits: `4c701b4`, `c09ca58`.

Ken added `.travis.yml` on the day of the rewrite. It ran the tests and sent
the coverage to Coveralls. Ken removed it on 2020-09-28, and the repository
had no CI until D17.

### D7. The decoder refuses an entry of 4096 bytes or more. Decided. Amended by D15.

Date: 2017-09-24. Commit: `9aa4502`.

Ken set `maxFileLength` to 4096 when he fixed the reading of extended
capabilities. `Decode` refuses a larger buffer with `ErrInvalidFileSize`.
4096 was the limit of the legacy format. The commit does not give the
reason.

### D8. The acs_chars value is ordered by source character, as infocmp prints it. Decided.

Date: 2017-10-02. Commits: `ce0306d`, `7fb97dc`.

`acs_chars` holds pairs of bytes. The first byte of a pair is the character
that a program asks for, and the second byte is what the terminal draws.
The decoder removes a pair whose first byte repeats, and sorts the pairs by
the first byte. The comment in `dec.go` points at `repair_acsc` in
`progs/dump_entry.c` of `ncurses` 6.3 as the model, and spells it
`repair_ascc`. This makes the value match what `infocmp` prints
(D3).

The first version sorted the second bytes and used them as keys, which broke
a pair whose two bytes differ. Vitaliy fixed it in pull request 5, and Ken
applied the commit on 2026-09-14.

### D9. Printf takes the capability as bytes and returns a string. Decided.

Date: 2017-10-05. Commit: `a749c6d`.

Ken renamed `Sprintf` to `Printf`. `Printf` takes the capability value as
`[]byte`, which is how `Terminfo.Strings` holds it, and the parameters as
`...any`. It returns a string. `Fprintf` writes the same string to an
`io.Writer`. The commit does not give the reason for the name.

### D10. ColorLevelFromEnv reads the environment before the database. Decided.

Date: 2017-10-09. Commits: `a42875d`, `ab3d866`.

`ColorLevelFromEnv` returns one of four levels: none, basic, hundreds and
millions. It reads `COLORTERM`, `FORCE_COLOR`, `TERM_PROGRAM` and
`TERM_PROGRAM_VERSION` first. Only if none of them decides does it load the
entry for `$TERM` and read `max_colors`. In 2018, Ken changed the
`COLORTERM` test to look for `truecolor` or `24bit` anywhere in the value.
`ChromaFormatterName` maps each level to a formatter name of
`github.com/alecthomas/chroma`.

### D11. A sentinel error is a constant of type Error. Decided.

Date: 2018-04-01. Commit: `20ea4ff`.

Ken converted the errors to "sentinel form". Each error is a constant of
`type Error string` with an `Err` prefix, declared in `terminfo.go`. A
caller compares with `==` or `errors.Is`.

### D12. Examples are in _examples, which the go command skips. Decided.

Date: 2019-01-25. Commits: `1a4775e`, `abceb7e`.

Ken renamed `examples` to `_examples`. The go command skips a folder whose
name starts with `_`, so `go build ./...` and `go test ./...` do not build
the example. The commit does not give the reason. In 2022, `abceb7e` moved
the example from `_examples/simple/main.go` to `_examples/main.go`.

### D13. The decoder reads the 32-bit number format. Decided.

Date: 2020-02-18. Commit: `454e5b6`, pull request 1.

Matt Peterson added the format that pull request 1 calls the "new ncurses
6.2 format". It has the magic number `01036`, and it stores each number in
32 bits and not in 16 bits. Ken merged the pull request. `Decode` reads
either magic number and sets the width of a number to match.

### D14. terminfo does not grow new features. Decided.

Date: 2021-02-16. Pull request 2.

Ken said on pull request 2 that the package looks unmaintained because there
is nothing to maintain. The terminfo file format stayed more or less the
same for decades. He can clean up the code, but there is nothing to add in
features or capabilities. He once planned a higher level library on top of
this one, in the manner of `ncurses`, and said that such a library is a
separate package.

### D15. The decoder accepts an entry of up to 32768 bytes, as ncurses 6.1 does. Decided. Amends D7.

Date: 2026-08-10. Commit: `a92d969`, pull request 3.

Sven Joachim raised `maxFileLength` from 4096 to 32768. The commit says that
32768 is the limit from `ncurses` 6.1 on, and that recent `xterm` and
`xterm-256color` entries are larger than 4096 bytes, so `infocmp` failed on
them. Ken merged the pull request. It cites Debian bug 1143527.

### D16. A lowercase variable is local to one evaluation, and an uppercase variable is shared. Decided.

Date: 2026-09-14. Commits: `79cb10a`, `3b29800`, pull request 6.

A parameterized string can store a value with `%P` and read it back with
`%g`. `%Pa` to `%Pz` store a dynamic variable, which belongs to one call of
`Printf`. `%PA` to `%PZ` store a static variable, which every call shares.
A mutex guards the static variables, from `3b29800` in 2016.

`%ga` read the static store until Vitaliy fixed it in pull request 6. Ken
applied the commit.

### D17. Tests run in GitHub Actions, on the ncurses that Ubuntu ships. Decided. Amends D6.

Date: 2026-09-14. Commit: `7ff4e87`.

Ken added `.github/workflows/test.yml`. It runs on `ubuntu-latest`, installs
`ncurses-bin` and `ncurses-term`, and prints the `infocmp` version. Then it
checks `gofmt -l .`, and runs `go build ./...`, `go vet ./...` and
`go test -count=1 ./...`. It runs on each push to `main`, on each pull
request, and by hand.

This replaces the Travis CI of D6, which Ken removed in 2020.

### D18. The module declares a recent Go version. Decided.

Date: 2026-09-14. Commits: `9ba9fdd`, `abceb7e`, `7ff4e87`.

Ken added `go.mod` in 2018. `abceb7e` set `go 1.19` in 2022, and `7ff4e87`
set `go 1.25` in 2026. CI tests Go 1.25 and the latest stable Go. The code
uses what Go 1.25 gives, such as `range` over an integer and the `slices`
and `maps` packages. No commit gives a rule for when the version goes up.

### D19. The tests accept the infocmp output of ncurses before 6.5. Decided.

Date: 2026-09-14. Commit: `ce65c9f`.

Before `ncurses` 6.5, `infocmp -E` did not escape a `"` in the description
of an entry. `strconv.Unquote` then refused the string. `ce65c9f` added
`unquote`, which escapes each bare `"` first, and `TestUnquote`. The tests
now pass on an older `ncurses` too.

### D20. The module requires no other module. Proposed.

Date: 2026-09-14. Commit: `c203b35`.

`c203b35` removed the last `require` line, `golang.org/x/exp`, and `go.sum`.
The package now imports only the standard library. No commit or comment says
that it must stay that way.

`gen.go` imports `github.com/kenshaw/snaker`, which `go.mod` does not
require, so `go generate` fails now. See open question 1 and B1.

### D21. terminfo makes no promise of backward compatibility. Decided.

Date: 2026-09-27. Source: tblfmt D37.

Ken said that none of the `xo` projects make a promise about their APIs. The
code is idiomatic Go, but the `xo` projects do not follow the compatibility
promise of the Go project. A change can rename or remove an exported
identifier when that makes the code better. tblfmt D37 records this for
every `xo` project.

### D22. terminfo is set up for coding agents as every xo repository is. Decided.

Date: 2026-09-27. Source: dbmeta D110 and dbmeta D111.

Ken set one agent setup for every repository in the `xo` namespace on
2026-09-27. dbmeta D110 records it, and dbmeta D111 says which projects keep
one file for each decision. `terminfo` follows it:

1. The root holds four documents: `README.md`, `AGENTS.md`, `CLAUDE.md` and
   `CONTRIBUTING.md`. Every other document goes in `docs/`.
2. `AGENTS.md` holds the rules for a coding agent, because Codex reads that
   file. It opens with the three standing rules of dbmeta D110.
3. `CLAUDE.md` is an ordinary file that holds one line, `@AGENTS.md`. It is
   not a symbolic link.
4. `simple-english` and `go-pedantry` are ordinary folders in
   `.agents/skills` and in `.claude/skills`, installed with `--copy`.
   `skills-lock.json` names the source of each.
5. `.gitignore` ignores `.claude/settings.local.json`, and `.gitattributes`
   holds `* text=auto eol=lf`.
6. `docs/PLAN.md` holds the plan, the decisions and the open questions.
   `terminfo` is a small library, so the decisions stay in this file, with
   the index at the top (dbmeta D111). `docs/BACKLOG.md` holds the work that
   is known and not done.

`skills_test.go` and `docs_test.go` hold the checks.
`TestSkillsAreCopies` fails on a symbolic link, on a missing copy, on two
copies that differ, and on a skill folder that `skills-lock.json` does not
name. `TestClaudeImportsAgents` fails unless `CLAUDE.md` holds only
`@AGENTS.md`. `TestTheRootHoldsFourDocuments` fails on any other document in
the root. `TestTheDecisionIndexIsComplete` checks each row of the table above
against its heading, and `TestAnAmendmentPointsBothWays` checks that each
amendment is named on both sides.

## Open questions for Ken

An open question is not a decision. When Ken answers one, it becomes a
decision above, and the question goes.

1. Must the module keep no `require` line (D20)? `gen.go` imports
   `github.com/kenshaw/snaker`, which `go.mod` does not require, so
   `go generate` fails. Does `snaker` go into `go.mod`, does `gen.go` get
   its own module, or does `gen.go` stop using `snaker`? See B1.
2. Which `ncurses` release wrote the current `capvals.go`? Its comments match
   neither 6.2 nor 6.3, which are in `.cache`. Must `gen.go` pin a release,
   and must `capvals.go` name the release that wrote it? See B2.
3. `ColorLevelFromEnv` returns `ColorLevelNone` for a terminal whose
   `max_colors` is 8 or 16, and `ColorLevelBasic` for one with 17 to 255.
   Is that deliberate? See B7.
4. Must CI run `golangci-lint`, with a `.golangci.yml` as in dburl and
   tblfmt? The repository has neither now.
