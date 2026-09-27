package terminfo_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"unicode"
)

// These tests read the documents of the repository and not the package.
// terminfo is a small library, so its decisions stay in docs/PLAN.md with an
// index at the top (dbmeta D111). See D22.

// planPath is the file that holds the plan and the decisions.
var planPath = filepath.Join("docs", "PLAN.md")

// TestTheRootHoldsFourDocuments holds D22. A document in the root that is not
// one of the four is a document that nobody filed in docs/.
func TestTheRootHoldsFourDocuments(t *testing.T) {
	t.Parallel()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	allowed := map[string]bool{
		"README.md": true, "AGENTS.md": true, "CLAUDE.md": true, "CONTRIBUTING.md": true,
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".md") {
			continue
		}
		if !allowed[e.Name()] {
			t.Errorf("%s is in the repository root. Only README, AGENTS, CLAUDE and CONTRIBUTING"+
				" belong there, and every other document goes in docs/. See D22", e.Name())
		}
	}
	for name := range allowed {
		if _, err := os.Stat(name); err != nil {
			t.Errorf("expected %s in the repository root", name)
		}
	}
}

// decision is one decision heading in docs/PLAN.md, such as:
//
//	### D7. The decoder refuses an entry of 4096 bytes or more. Decided. Amended by D15.
type decision struct {
	num    int
	title  string
	status string
	anchor string
}

// decisionHeading matches a decision heading. The status starts with Decided
// or Proposed, and can name an amendment after it.
var decisionHeading = regexp.MustCompile(`^### D(\d+)\. (.+?)\. ((?:Decided|Proposed)(?:\. .+?)?)\.$`)

// decisions reads each decision heading in docs/PLAN.md, in order. It fails on
// a heading that starts with "### D" and does not have the form, and on a
// number that is out of order.
func decisions(t *testing.T) []decision {
	t.Helper()
	var out []decision
	for _, line := range strings.Split(string(read(t, planPath)), "\n") {
		if !strings.HasPrefix(line, "### D") {
			continue
		}
		m := decisionHeading.FindStringSubmatch(line)
		if m == nil {
			t.Errorf("%s: %q does not have the form \"### D<n>. <title>. <status>.\","+
				" where the status starts with Decided or Proposed", planPath, line)
			continue
		}
		num, _ := strconv.Atoi(m[1])
		if want := len(out) + 1; num != want {
			t.Errorf("%s: D%d follows D%d. The decisions are numbered in order, from D1", planPath, num, want-1)
		}
		out = append(out, decision{
			num:    num,
			title:  m[2],
			status: m[3],
			anchor: anchor(strings.TrimPrefix(line, "### ")),
		})
	}
	if len(out) == 0 {
		t.Fatalf("%s holds no decision", planPath)
	}
	return out
}

// anchor returns the anchor that GitHub gives to a heading: lower case, with
// each space made a hyphen, and without the punctuation.
func anchor(heading string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(heading) {
		switch {
		case r == ' ':
			b.WriteByte('-')
		case r == '-' || r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
		}
	}
	return b.String()
}

// TestTheDecisionIndexIsComplete checks the table at the top of docs/PLAN.md
// against the decision headings. A reader finds a decision by its number in
// that table, so a missing row or a stale title or status hides it.
func TestTheDecisionIndexIsComplete(t *testing.T) {
	t.Parallel()
	rows := make(map[int]string)
	for _, m := range regexp.MustCompile(`(?m)^\| \[D(\d+)\]\(.*$`).FindAllStringSubmatch(string(read(t, planPath)), -1) {
		num, _ := strconv.Atoi(m[1])
		if _, ok := rows[num]; ok {
			t.Errorf("%s has two rows for D%d", planPath, num)
		}
		rows[num] = m[0]
	}
	written := make(map[int]bool)
	for _, d := range decisions(t) {
		written[d.num] = true
		want := fmt.Sprintf("| [D%d](#%s) | %s | %s |", d.num, d.anchor, d.title, d.status)
		switch got, ok := rows[d.num]; {
		case !ok:
			t.Errorf("D%d has no row in the index in %s. Add:\n%s", d.num, planPath, want)
		case got != want:
			t.Errorf("D%d: the row in the index in %s is\n%s\nand the heading says\n%s", d.num, planPath, got, want)
		}
	}
	for num := range rows {
		if !written[num] {
			t.Errorf("the index in %s has a row for D%d, and no heading holds it", planPath, num)
		}
	}
}

// TestAnAmendmentPointsBothWays checks that each amendment is named in the
// headings of both decisions. If D15 amends D7, then D15 says "Amends D7" and
// D7 says "Amended by D15". A reader who finds only the older decision then
// sees at once that it no longer holds as written.
func TestAnAmendmentPointsBothWays(t *testing.T) {
	t.Parallel()
	status := make(map[int]string)
	for _, d := range decisions(t) {
		status[d.num] = d.status
	}
	naming := regexp.MustCompile(`\b(Amends|Amended by) D(\d+)\b`)
	for num, s := range status {
		for _, m := range naming.FindAllStringSubmatch(s, -1) {
			other, _ := strconv.Atoi(m[2])
			back := "Amended by"
			if m[1] == "Amended by" {
				back = "Amends"
			}
			switch otherStatus, ok := status[other]; {
			case !ok:
				t.Errorf("D%d says %q, and D%d is not a decision", num, m[0], other)
			case !regexp.MustCompile(`\b` + back + ` D` + strconv.Itoa(num) + `\b`).MatchString(otherStatus):
				t.Errorf("D%d says %q, and the heading of D%d does not say %q. An amendment "+
					"is named on both sides. See D22", num, m[0], other, fmt.Sprintf("%s D%d", back, num))
			}
		}
	}
}

// TestEveryDecisionReferenceExists checks that a bare decision number in a
// document or a Go file names a decision in docs/PLAN.md. A number that
// follows the name of another repository, such as "dbmeta D110", names a
// decision of that repository, and the test skips it.
func TestEveryDecisionReferenceExists(t *testing.T) {
	t.Parallel()
	written := make(map[string]bool)
	for _, d := range decisions(t) {
		written[strconv.Itoa(d.num)] = true
	}
	ref := regexp.MustCompile(`(\b(?:dbimp|dbmeta|dbtpl|dburl|tblfmt|usql) )?\bD([1-9][0-9]*)\b`)
	for _, path := range repoFiles(t, ".md", ".go") {
		for _, m := range ref.FindAllStringSubmatch(string(read(t, path)), -1) {
			if m[1] == "" && !written[m[2]] {
				t.Errorf("%s: names D%s, which is not a decision in %s. If it is a decision of"+
					" another repository, name that repository first, such as \"dbmeta D%s\"",
					path, m[2], planPath, m[2])
			}
		}
	}
}

// repoFiles returns each file with one of the given extensions. It skips the
// folders whose names start with a dot, which hold the agent skills, git and
// the download cache of gen.go.
func repoFiles(t *testing.T, exts ...string) []string {
	t.Helper()
	var out []string
	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case d.IsDir() && path != "." && strings.HasPrefix(d.Name(), "."):
			return filepath.SkipDir
		case d.IsDir():
			return nil
		}
		for _, ext := range exts {
			if strings.HasSuffix(path, ext) {
				out = append(out, path)
				break
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}
