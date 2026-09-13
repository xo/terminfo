package terminfo

import "testing"

func TestDynamicVariables(t *testing.T) {
	for _, name := range []byte{'a', 'z'} {
		t.Run(string(name), func(t *testing.T) {
			program := "%{7}%P" + string(name) + "%g" + string(name) + "%d"
			p := newParametizer([]byte(program))
			defer p.reset()
			if got := p.exec(); got != "7" {
				t.Fatalf("got %q, want 7", got)
			}
		})
	}
}

func TestDynamicAndStaticVariablesAreSeparate(t *testing.T) {
	staticVars.Lock()
	saved := staticVars.vars[0]
	staticVars.vars[0] = 41
	staticVars.Unlock()
	defer func() { staticVars.Lock(); staticVars.vars[0] = saved; staticVars.Unlock() }()
	p := newParametizer([]byte("%{7}%Pa%ga%d/%gA%d"))
	defer p.reset()
	if got := p.exec(); got != "7/41" {
		t.Fatalf("got %q, want 7/41", got)
	}
}
