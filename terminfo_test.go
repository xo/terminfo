package terminfo

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func TestLoad(t *testing.T) {
	for _, term := range slices.Sorted(maps.Keys(terms(t))) {
		t.Run(term, func(t *testing.T) {
			if err := os.Setenv("TERM", term); err != nil {
				t.Fatalf("could not set TERM environment variable, got: %v", err)
			}
			// open
			info, err := LoadFromEnv()
			if err != nil {
				t.Fatalf("term %s expected no error, got: %v", term, err)
			}
			// check we have at least one name
			if len(info.Names) < 1 {
				t.Errorf("term %s expected names to have at least one value", term)
			}
		})
	}
}

func TestOpen(t *testing.T) {
	for term, filename := range terms(t) {
		t.Run(strings.TrimPrefix(filename, "/"), func(t *testing.T) {
			// open
			ti, err := Open(filepath.Dir(filepath.Dir(filename)), term)
			if err != nil {
				t.Fatalf("term %s expected no error, got: %v", term, err)
			}
			if ti.File != filename {
				t.Errorf("term %s should have file %s, got: %s", term, filename, ti.File)
			}
			// check we have at least one name
			if len(ti.Names) < 1 {
				t.Errorf("term %s expected names to have at least one value", term)
			}
		})
	}
}

func TestValues(t *testing.T) {
	for term, filename := range terms(t) {
		t.Run(filepath.Base(filename), func(t *testing.T) {
			t.Parallel()
			ic, err := getInfocmpData(t, term)
			if err != nil {
				t.Fatalf("expected no error, got: %s", err)
			}
			// open the file directly: a different entry can claim this term
			// as an alias, and would then satisfy Load from the term cache
			ti, err := Open(filepath.Dir(filepath.Dir(filename)), term)
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			// check names
			if !reflect.DeepEqual(ic.names, ti.Names) {
				t.Errorf("names do not match")
			}
			// check bool caps
			for i, v := range ic.boolCaps {
				if v == nil {
					if _, ok := ti.BoolsM[i]; !ok {
						t.Errorf("expected bool cap %d (%s) to be missing", i, BoolCapName(i))
					}
				} else if v.(bool) != ti.Bools[i] {
					t.Errorf("bool cap %d (%s) should be %t", i, BoolCapName(i), v)
				}
			}
			// check extended bool caps
			if len(ic.extBoolCaps) != len(ti.ExtBools) {
				t.Errorf("should have same number of extended bools (%d, %d)", len(ic.extBoolCaps), len(ti.ExtBools))
			}
			for i, v := range ic.extBoolCaps {
				z, ok := ti.ExtBools[i]
				if !ok {
					t.Errorf("should have extended bool %d", i)
				}
				if v.(bool) != z {
					t.Errorf("extended bool cap %d (%s) should be %t", i, ic.extBoolNames[i], v)
				}
				n, ok := ti.ExtBoolNames[i]
				if !ok {
					t.Errorf("missing extended bool %d name", i)
				}
				if string(n) != ic.extBoolNames[i] {
					t.Errorf("extended bool %d name should be '%s', got: '%s'", i, ic.extBoolNames[i], string(n))
				}
			}
			// check num caps
			for i, v := range ic.numCaps {
				if v == nil {
					if _, ok := ti.NumsM[i]; !ok {
						// t.Errorf("term %s expected num cap %d (%s) to be missing", i, NumCapName(i))
					}
				} else if v.(int) != ti.Nums[i] {
					t.Errorf("num cap %d (%s) should be %d", i, NumCapName(i), v)
				}
			}
			// check extended num caps
			if len(ic.extNumCaps) != len(ti.ExtNums) {
				t.Errorf("should have same number of extended nums (%d, %d)", len(ic.extNumCaps), len(ti.ExtNums))
			}
			for i, v := range ic.extNumCaps {
				z, ok := ti.ExtNums[i]
				if !ok {
					t.Errorf("should have extended num %d", i)
				}
				if v.(int) != z {
					t.Errorf("extended num cap %d (%s) should be %t", i, ic.extNumNames[i], v)
				}
				n, ok := ti.ExtNumNames[i]
				if !ok {
					t.Errorf("missing extended num %d name", i)
				}
				if string(n) != ic.extNumNames[i] {
					t.Errorf("extended num %d name should be '%s', got: '%s'", i, ic.extNumNames[i], string(n))
				}
			}
			// check string caps
			for i, v := range ic.stringCaps {
				if v == nil {
					if _, ok := ti.StringsM[i]; !ok {
						// t.Errorf("term %s expected string cap %d (%s) to be missing", i, StringCapName(i))
					}
				} else if v.(string) != string(ti.Strings[i]) {
					t.Errorf("string cap %d (%s) is invalid:", i, StringCapName(i))
					t.Errorf("got:  %#v", string(ti.Strings[i]))
					t.Errorf("want: %#v", v)
				}
			}
			// check extended string caps
			if len(ic.extStringCaps) != len(ti.ExtStrings) {
				t.Errorf("should have same number of extended strings (%d, %d)", len(ic.extStringCaps), len(ti.ExtStrings))
			}
			for i, v := range ic.extStringCaps {
				z, ok := ti.ExtStrings[i]
				if !ok {
					t.Errorf("should have extended string %d", i)
				}
				if v.(string) != string(z) {
					t.Errorf("extended string cap %d (%s) should be %t", i, ic.extStringNames[i], v)
				}
				n, ok := ti.ExtStringNames[i]
				if !ok {
					t.Errorf("missing extended string %d name", i)
				}
				if string(n) != ic.extStringNames[i] {
					t.Errorf("extended string %d name should be '%s', got: '%s'", i, ic.extStringNames[i], string(n))
				}
			}
		})
	}
}

var shortCapNameMap map[string]int

type infocmp struct {
	names          []string
	boolCaps       map[int]any
	numCaps        map[int]any
	stringCaps     map[int]any
	extBoolCaps    map[int]any
	extNumCaps     map[int]any
	extStringCaps  map[int]any
	extBoolNames   map[int]string
	extNumNames    map[int]string
	extStringNames map[int]string
}

func init() {
	shortCapNameMap = make(map[string]int)
	for _, z := range [][]string{boolCapNames[:], numCapNames[:], stringCapNames[:]} {
		for i := 0; i < len(z); i += 2 {
			shortCapNameMap[z[i+1]] = i / 2
		}
	}
}

// unquote unquotes the C string literal s.
//
// Any bare " in the body is escaped first: infocmp did not escape the " in an
// entry's description prior to ncurses 6.5.
func unquote(s string) (string, error) {
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return "", fmt.Errorf("%s is not a quoted string", s)
	}
	var b strings.Builder
	b.WriteByte('"')
	for i, end := 1, len(s)-1; i < end; i++ {
		switch s[i] {
		case '\\':
			// copy the escape sequence as-is
			if i+1 < end {
				b.WriteByte(s[i])
				i++
			}
			b.WriteByte(s[i])
		case '"':
			b.WriteString(`\"`)
		default:
			b.WriteByte(s[i])
		}
	}
	b.WriteByte('"')
	return strconv.Unquote(b.String())
}

func TestUnquote(t *testing.T) {
	for _, test := range []struct {
		name, input, want string
	}{
		// ncurses 6.5 and later escapes the " in a description
		{"escaped quotes", `"linux+kbs|Linux fragment for \"backspace\" key"`, `linux+kbs|Linux fragment for "backspace" key`},
		// ncurses before 6.5 did not
		{"bare quotes", `"linux+kbs|Linux fragment for "backspace" key"`, `linux+kbs|Linux fragment for "backspace" key`},
		{"leading bare quote", `""standard" printer info for HP ttys"`, `"standard" printer info for HP ttys`},
		{"trailing bare quote", `"xterm focus-in/out event "keys""`, `xterm focus-in/out event "keys"`},
		{"escapes preserved", `"\033[%i%p1%d;%p2%dr"`, "\x1b[%i%p1%d;%p2%dr"},
		{"escaped backslash", `"a\\\"b"`, `a\"b`},
		{"empty", `""`, ``},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := unquote(test.input)
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if got != test.want {
				t.Errorf("unquote(%s) = %q, want %q", test.input, got, test.want)
			}
		})
	}
	for _, test := range []string{``, `"`, `abc`, `"abc`} {
		if _, err := unquote(test); err == nil {
			t.Errorf("unquote(%q) expected an error", test)
		}
	}
}

var staticCharRE = regexp.MustCompile(`(?m)^static\s+char\s+(.*)\s*\[\]\s*=\s*(".*");$`)

func getInfocmpData(t *testing.T, term string) (*infocmp, error) {
	t.Logf("loading infocmp data for term %s", term)
	c := exec.Command("/usr/bin/infocmp", "-E", "-x")
	c.Env = []string{"TERM=" + term}
	buf, err := c.CombinedOutput()
	if err != nil {
		t.Logf("shell error (TERM=%s):\n%s\n", term, string(buf))
		return nil, err
	}
	// read static strings
	m := staticCharRE.FindAllStringSubmatch(string(buf), -1)
	if !strings.HasSuffix(strings.TrimSpace(m[0][1]), "_alias_data") {
		return nil, errors.New("missing _alias_data")
	}
	names, err := unquote(m[0][2])
	if err != nil {
		return nil, fmt.Errorf("could not unquote _alias_data: %v", err)
	}
	ic := &infocmp{
		names:          strings.Split(names, "|"),
		boolCaps:       make(map[int]any),
		numCaps:        make(map[int]any),
		stringCaps:     make(map[int]any),
		extBoolCaps:    make(map[int]any),
		extNumCaps:     make(map[int]any),
		extStringCaps:  make(map[int]any),
		extBoolNames:   make(map[int]string),
		extNumNames:    make(map[int]string),
		extStringNames: make(map[int]string),
	}
	// load the declared strings, keyed by the identifier that the cap value
	// arrays reference. infocmp declares a <term>_s_<cap> for every string cap
	// value, and a name_of_<cap> for every extended cap name.
	caps := make(map[string]string, len(m))
	for i, s := range m[1:] {
		k := strings.TrimSpace(s[1])
		v, err := unquote(s[2])
		if err != nil {
			return nil, fmt.Errorf("could not unquote %d (%s): %v", i, k, err)
		}
		caps[k] = v
	}
	// extract the values
	for _, err := range []error{
		processSect(buf, caps, ic.boolCaps, ic.extBoolCaps, ic.extBoolNames, boolSectRE),
		processSect(buf, caps, ic.numCaps, ic.extNumCaps, ic.extNumNames, numSectRE),
		processSect(buf, caps, ic.stringCaps, ic.extStringCaps, ic.extStringNames, stringSectRE),
	} {
		if err != nil {
			return nil, err
		}
	}
	return ic, nil
}

// regexp's used by processSect.
var (
	boolSectRE   = regexp.MustCompile(`_bool_data\[\]\s*=\s*{`)
	numSectRE    = regexp.MustCompile(`_number_data\[\]\s*=\s*{`)
	stringSectRE = regexp.MustCompile(`_string_data\[\]\s*=\s*{`)
	endSectRE    = regexp.MustCompile(`(?m)^};$`)
	capValuesRE  = regexp.MustCompile(`(?m)^\s+/\*\s+[0-9]+:\s+([^\s]+)\s+\*/\s+(.*),$`)
	numRE        = regexp.MustCompile(`^[0-9]+$`)
	absCanRE     = regexp.MustCompile(`^(ABSENT|CANCELLED)_(BOOLEAN|NUMERIC|STRING)$`)
)

// processSect processes a text section in the infocmp C export.
func processSect(buf []byte, caps map[string]string, xx, yy map[int]any, extn map[int]string, sectRE *regexp.Regexp) error {
	var err error
	// extract section
	start := sectRE.FindIndex(buf)
	if start == nil || len(start) != 2 {
		return fmt.Errorf("could not find sect (%s)", sectRE)
	}
	end := endSectRE.FindIndex(buf[start[1]:])
	if end == nil || len(end) != 2 {
		return fmt.Errorf("could not find end of section (%s)", sectRE)
	}
	buf = buf[start[1] : start[1]+end[0]]
	// load caps
	m := capValuesRE.FindAllStringSubmatch(string(buf), -1)
	var extc int
	for i, s := range m {
		var skip bool
		// determine target
		target := xx
		k, ok := shortCapNameMap[s[1]]
		if !ok {
			target, k, extn[extc] = yy, extc, s[1]
			extc++
		}
		// get cap value
		var v any
		switch {
		case s[2] == "TRUE" || s[2] == "FALSE":
			v = s[2] == "TRUE"
		case numRE.MatchString(s[2]):
			var j int64
			j, err = strconv.ParseInt(s[2], 10, 32)
			if err != nil {
				return fmt.Errorf("line %d could not parse: %v", i, err)
			}
			v = int(j)
		case absCanRE.MatchString(s[2]):
			if !ok { // absent/canceled extended cap
				if strings.HasSuffix(s[2], "NUMERIC") {
					v = -1
				} else {
					skip = true
				}
			}
		default:
			v, ok = caps[s[2]]
			if !ok {
				return fmt.Errorf("cap '%s' not defined in cap table", s[2])
			}
		}
		if !skip {
			target[k] = v
		}
	}
	return nil
}

var termNameCache = struct {
	names map[string]string
	sync.Mutex
}{}

var fileRE = regexp.MustCompile("^([0-9]+|[a-zA-Z])/")

func terms(t *testing.T) map[string]string {
	termNameCache.Lock()
	defer termNameCache.Unlock()
	if termNameCache.names == nil {
		termNameCache.names = make(map[string]string)
		for _, dir := range termDirs() {
			err := filepath.Walk(dir, func(n string, fi os.FileInfo, err error) error {
				switch {
				case err != nil:
					return err
				case fi.IsDir(), dir == n, !fileRE.MatchString(n[len(dir)+1:]), fi.Mode()&os.ModeSymlink != 0:
					return nil
				}
				term := filepath.Base(n)
				if term == "xterm-old" {
					return nil
				}
				termNameCache.names[term] = n
				return nil
			})
			if err != nil {
				t.Fatalf("could not walk directory, got: %v", err)
			}
		}
	}
	return termNameCache.names
}

func termDirs() []string {
	var dirs []string
	for _, dir := range []string{"/lib/terminfo", "/usr/share/terminfo"} {
		fi, err := os.Stat(dir)
		if err != nil && os.IsNotExist(err) {
			continue
		} else if err != nil {
			panic(err)
		}
		if fi.IsDir() {
			dirs = append(dirs, dir)
		}
	}
	return dirs
}

func TestCapSizes(t *testing.T) {
	if CapCountBool*2 != len(boolCapNames) {
		t.Fatalf("boolCapNames should have same length as twice CapCountBool")
	}
	if CapCountNum*2 != len(numCapNames) {
		t.Fatalf("numCapNames should have same length as twice CapCountNum")
	}
	if CapCountString*2 != len(stringCapNames) {
		t.Fatalf("stringCapNames should have same length as twice CapCountString")
	}
}

func TestCapNames(t *testing.T) {
	for i := range CapCountBool {
		n, s := BoolCapName(i), BoolCapNameShort(i)
		if n == "" {
			t.Errorf("Bool cap %d should have name", i)
		}
		if s == "" {
			t.Errorf("Bool cap %d should have short name", i)
		}
		if n == s {
			t.Errorf("Bool cap %d name and short name should not equal (%s==%s)", i, n, s)
		}
	}
	for i := range CapCountNum {
		n, s := NumCapName(i), NumCapNameShort(i)
		if n == "" {
			t.Errorf("Num cap %d should have name", i)
		}
		if s == "" {
			t.Errorf("Num cap %d should have short name", i)
		}
		if n == s && n != "lines" {
			t.Errorf("Num cap %d name and short name should not equal (%s==%s)", i, n, s)
		}
	}
	for i := range CapCountString {
		n, s := StringCapName(i), StringCapNameShort(i)
		if n == "" {
			t.Errorf("String cap %d should have name", i)
		}
		if s == "" {
			t.Errorf("String cap %d should have short name", i)
		}
		if n == s && n != "tone" && n != "pulse" {
			t.Errorf("String cap %d name and short name should not equal (%s==%s)", i, n, s)
		}
	}
}
