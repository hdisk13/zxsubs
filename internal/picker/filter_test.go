package picker

import "testing"

func TestMatch(t *testing.T) {
	tests := []struct {
		query string
		parts []string
		want  bool
	}{
		{"", []string{"Contoso Production"}, true},
		{"  ", []string{"Contoso Production"}, true},
		{"prod", []string{"Contoso Production", "1111"}, true},
		{"PROD", []string{"Contoso Production"}, true},
		{"cprd", []string{"Contoso Production"}, true},
		{"1111", []string{"Contoso Production", "11111111-1111"}, true},
		{"disabled", []string{"Legacy Billing", "Disabled"}, true},
		{"zzzz", []string{"Contoso Production"}, false},
		{"fabrikam", []string{"Contoso Production"}, false},
	}
	for _, tt := range tests {
		if got := Match(tt.query, tt.parts...); got != tt.want {
			t.Errorf("Match(%q, %v) = %v, want %v", tt.query, tt.parts, got, tt.want)
		}
	}
}
