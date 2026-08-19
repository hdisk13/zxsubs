package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/hdisk13/zxsubs/internal/azure"
)

type fakeCLI struct {
	subs    []azure.Subscription
	listErr error
	setErr  error
	setID   string
}

func (f *fakeCLI) List() ([]azure.Subscription, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.subs, nil
}

func (f *fakeCLI) Set(id string) error {
	f.setID = id
	return f.setErr
}

func sample() []azure.Subscription {
	return []azure.Subscription{
		{Name: "Contoso Development", ID: "2222", IsDefault: true, State: "Enabled"},
		{Name: "Contoso Production", ID: "1111", State: "Enabled"},
	}
}

func TestRunSelect(t *testing.T) {
	cli := &fakeCLI{subs: sample()}
	var out, err bytes.Buffer
	pick := func(subs []azure.Subscription) (azure.Subscription, bool, error) {
		return subs[1], true, nil
	}
	code := run(nil, cli, &out, &err, pick)
	if code != 0 {
		t.Fatalf("exit %d, stderr %s", code, err.String())
	}
	if cli.setID != "1111" {
		t.Fatalf("set %q", cli.setID)
	}
	if !strings.Contains(out.String(), "Contoso Production") || !strings.Contains(out.String(), "1111") {
		t.Fatalf("stdout %q", out.String())
	}
}

func TestRunCancelDoesNotSet(t *testing.T) {
	cli := &fakeCLI{subs: sample()}
	var out, errBuf bytes.Buffer
	pick := func([]azure.Subscription) (azure.Subscription, bool, error) {
		return azure.Subscription{}, false, nil
	}
	code := run(nil, cli, &out, &errBuf, pick)
	if code != 1 {
		t.Fatalf("exit %d", code)
	}
	if cli.setID != "" {
		t.Fatalf("set was called with %q", cli.setID)
	}
	if !strings.Contains(errBuf.String(), "Canceled.") {
		t.Fatalf("stderr %q", errBuf.String())
	}
}

func TestRunListError(t *testing.T) {
	cli := &fakeCLI{listErr: errors.New("you do not appear to be logged in to Azure CLI.\nRun: az login")}
	var out, errBuf bytes.Buffer
	code := run(nil, cli, &out, &errBuf, nil)
	if code != 1 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(errBuf.String(), "az login") {
		t.Fatalf("stderr %q", errBuf.String())
	}
}

func TestRunSetError(t *testing.T) {
	cli := &fakeCLI{subs: sample(), setErr: errors.New("az failed")}
	var out, errBuf bytes.Buffer
	pick := func(subs []azure.Subscription) (azure.Subscription, bool, error) {
		return subs[0], true, nil
	}
	code := run(nil, cli, &out, &errBuf, pick)
	if code != 1 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(errBuf.String(), "az failed") {
		t.Fatalf("stderr %q", errBuf.String())
	}
}

func TestRunHelp(t *testing.T) {
	var out, errBuf bytes.Buffer
	code := run([]string{"-h"}, &fakeCLI{}, &out, &errBuf, nil)
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(errBuf.String(), "az login") {
		t.Fatalf("usage %q", errBuf.String())
	}
}

func TestRunRejectsArgs(t *testing.T) {
	var out, errBuf bytes.Buffer
	code := run([]string{"surprise"}, &fakeCLI{}, &out, &errBuf, nil)
	if code != 1 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(errBuf.String(), "takes no arguments") {
		t.Fatalf("stderr %q", errBuf.String())
	}
}
