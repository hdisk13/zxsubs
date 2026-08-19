package azure

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testdata(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "account_list.json"))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestParseAccountList(t *testing.T) {
	subs, err := ParseAccountList(testdata(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(subs) != 4 {
		t.Fatalf("got %d subscriptions, want 4", len(subs))
	}

	if subs[0].Name != "Contoso Production" {
		t.Errorf("first name = %q", subs[0].Name)
	}
	if subs[1].IsDefault != true {
		t.Error("expected Contoso Development to be the default subscription")
	}
	if subs[3].State != "Disabled" {
		t.Errorf("legacy state = %q, want Disabled", subs[3].State)
	}
	if got := ShortID(subs[0].ID); got != "11111111" {
		t.Errorf("ShortID = %q, want 11111111", got)
	}
}

func TestParseAccountList_skipsEmptyID(t *testing.T) {
	subs, err := ParseAccountList([]byte(`[
		{"name":"No ID","id":"","isDefault":false,"state":"Enabled","tenantId":"t"},
		{"name":"Keep","id":"abc","isDefault":true,"state":"Enabled","tenantId":"t"}
	]`))
	if err != nil {
		t.Fatal(err)
	}
	if len(subs) != 1 || subs[0].Name != "Keep" {
		t.Fatalf("got %+v", subs)
	}
}

func TestParseAccountList_errors(t *testing.T) {
	if _, err := ParseAccountList(nil); err == nil {
		t.Fatal("empty input should fail")
	}
	if _, err := ParseAccountList([]byte("not-json")); err == nil {
		t.Fatal("invalid json should fail")
	}
}

func TestList_success(t *testing.T) {
	c := &Client{
		lookPath: func(string) (string, error) { return "/usr/bin/az", nil },
		exec: func(name string, arg ...string) ([]byte, []byte, error) {
			if name != "az" {
				t.Fatalf("binary %q", name)
			}
			want := []string{"account", "list", "--output", "json"}
			if strings.Join(arg, " ") != strings.Join(want, " ") {
				t.Fatalf("args %v", arg)
			}
			return testdata(t), nil, nil
		},
	}
	subs, err := c.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(subs) != 4 {
		t.Fatalf("got %d", len(subs))
	}
}

func TestList_missingAz(t *testing.T) {
	c := &Client{
		lookPath: func(string) (string, error) { return "", errors.New("not found") },
	}
	_, err := c.List()
	if err == nil || !strings.Contains(err.Error(), "was not found on PATH") {
		t.Fatalf("error = %v", err)
	}
}

func TestList_notLoggedIn(t *testing.T) {
	c := &Client{
		lookPath: func(string) (string, error) { return "az", nil },
		exec: func(string, ...string) ([]byte, []byte, error) {
			return nil, []byte("ERROR: Please run 'az login' to setup account."), errors.New("exit status 1")
		},
	}
	_, err := c.List()
	if err == nil || !strings.Contains(err.Error(), "az login") {
		t.Fatalf("error = %v", err)
	}
}

func TestList_empty(t *testing.T) {
	c := &Client{
		lookPath: func(string) (string, error) { return "az", nil },
		exec: func(string, ...string) ([]byte, []byte, error) {
			return []byte("[]"), nil, nil
		},
	}
	_, err := c.List()
	if err == nil || !strings.Contains(err.Error(), "no Azure subscriptions") {
		t.Fatalf("error = %v", err)
	}
}

func TestSet(t *testing.T) {
	var gotArgs []string
	c := &Client{
		lookPath: func(string) (string, error) { return "az", nil },
		exec: func(name string, arg ...string) ([]byte, []byte, error) {
			gotArgs = append([]string{name}, arg...)
			return nil, nil, nil
		},
	}
	if err := c.Set("22222222-2222-2222-2222-222222222222"); err != nil {
		t.Fatal(err)
	}
	want := "az account set --subscription 22222222-2222-2222-2222-222222222222"
	if strings.Join(gotArgs, " ") != want {
		t.Fatalf("ran %q", strings.Join(gotArgs, " "))
	}
}

func TestWrapAzError_generic(t *testing.T) {
	err := wrapAzError(errors.New("exit 1"), []byte("Something exploded"))
	if !strings.Contains(err.Error(), "Something exploded") {
		t.Fatal(err)
	}
}
