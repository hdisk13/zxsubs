package picker

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/hdisk13/zxsubs/internal/azure"
)

func sampleSubs() []azure.Subscription {
	return []azure.Subscription{
		{Name: "Contoso Production", ID: "11111111-1111-1111-1111-111111111111", TenantID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", State: "Enabled"},
		{Name: "Contoso Development", ID: "22222222-2222-2222-2222-222222222222", TenantID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", State: "Enabled", IsDefault: true},
		{Name: "Fabrikam Playground", ID: "33333333-3333-3333-3333-333333333333", TenantID: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", State: "Enabled"},
		{Name: "Legacy Billing", ID: "44444444-4444-4444-4444-444444444444", TenantID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", State: "Disabled"},
	}
}

func mustModel(t *testing.T, next tea.Model) model {
	t.Helper()
	m, ok := next.(model)
	if !ok {
		t.Fatalf("got %T", next)
	}
	return m
}

func TestNewModelStartsOnCurrent(t *testing.T) {
	m := newModel(sampleSubs())
	if m.filtered[m.cursor].Name != "Contoso Development" {
		t.Fatalf("cursor on %q", m.filtered[m.cursor].Name)
	}
}

func TestArrowsAndEnter(t *testing.T) {
	m := newModel(sampleSubs())
	m = mustModel(t, updateKey(m, tea.KeyUp))
	if m.filtered[m.cursor].Name != "Contoso Production" {
		t.Fatalf("after up, cursor on %q", m.filtered[m.cursor].Name)
	}

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mustModel(t, next)
	if m.chosen == nil || m.chosen.Name != "Contoso Production" {
		t.Fatalf("chosen = %+v", m.chosen)
	}
	if cmd == nil {
		t.Fatal("expected tea.Quit")
	}
}

func TestCancelKeys(t *testing.T) {
	for _, key := range []tea.KeyType{tea.KeyEsc, tea.KeyCtrlC} {
		m := newModel(sampleSubs())
		next, cmd := m.Update(tea.KeyMsg{Type: key})
		m = mustModel(t, next)
		if !m.canceled {
			t.Fatalf("key %v should cancel", key)
		}
		if cmd == nil {
			t.Fatal("expected tea.Quit")
		}
	}

	m := newModel(sampleSubs())
	m = mustModel(t, updateRunes(m, "q"))
	if !m.canceled {
		t.Fatal("q should cancel")
	}
}

func TestTypeToFilter(t *testing.T) {
	m := newModel(sampleSubs())
	m = mustModel(t, updateRunes(m, "fab"))
	if len(m.filtered) != 1 || m.filtered[0].Name != "Fabrikam Playground" {
		t.Fatalf("filtered = %+v", m.filtered)
	}

	m = mustModel(t, updateKey(m, tea.KeyBackspace))
	if m.filter != "fa" {
		t.Fatalf("filter after backspace = %q", m.filter)
	}

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mustModel(t, next)
	if m.chosen == nil || m.chosen.Name != "Fabrikam Playground" {
		t.Fatalf("chosen = %+v", m.chosen)
	}
}

func TestFilterNoMatchEnterDoesNothing(t *testing.T) {
	m := newModel(sampleSubs())
	m = mustModel(t, updateRunes(m, "zzzz"))
	if len(m.filtered) != 0 {
		t.Fatal("expected no matches")
	}
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mustModel(t, next)
	if m.chosen != nil || cmd != nil {
		t.Fatal("enter on empty filter should not quit")
	}
}

func TestViewShowsContext(t *testing.T) {
	m := newModel(sampleSubs())
	v := m.View()
	for _, want := range []string{"zxsubs", "Contoso Production", "current", "…22222222", "Enabled", "Disabled"} {
		if !strings.Contains(v, want) {
			t.Errorf("view missing %q\n%s", want, v)
		}
	}
}

func updateKey(m model, k tea.KeyType) tea.Model {
	next, _ := m.Update(tea.KeyMsg{Type: k})
	return next
}

func updateRunes(m model, s string) tea.Model {
	var next tea.Model = m
	for _, r := range s {
		next, _ = next.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	return next
}
