// Package azure wraps the Azure CLI (az) to list and select subscriptions.
package azure

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// Subscription is one entry from `az account list`.
type Subscription struct {
	Name      string `json:"name"`
	ID        string `json:"id"`
	TenantID  string `json:"tenantId"`
	State     string `json:"state"`
	IsDefault bool   `json:"isDefault"`
}

// Client talks to Azure through the existing az CLI session.
type Client struct {
	lookPath func(file string) (string, error)
	exec     func(name string, arg ...string) (stdout, stderr []byte, err error)
}

// New returns a Client that runs the real az binary on PATH.
func New() *Client {
	return &Client{
		lookPath: exec.LookPath,
		exec:     runCmd,
	}
}

func runCmd(name string, arg ...string) (stdout, stderr []byte, err error) {
	cmd := exec.Command(name, arg...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	return outBuf.Bytes(), errBuf.Bytes(), err
}

// List returns subscriptions from `az account list --output json`.
func (c *Client) List() ([]Subscription, error) {
	if _, err := c.lookPath("az"); err != nil {
		return nil, fmt.Errorf("Azure CLI (az) was not found on PATH.\nInstall it: https://learn.microsoft.com/cli/azure/install-azure-cli\nThen run: az login")
	}

	stdout, stderr, err := c.exec("az", "account", "list", "--output", "json")
	if err != nil {
		return nil, wrapAzError(err, stderr)
	}

	subs, err := ParseAccountList(stdout)
	if err != nil {
		return nil, err
	}
	if len(subs) == 0 {
		return nil, fmt.Errorf("no Azure subscriptions found.\nConfirm you are logged in: az login\nThen check: az account list")
	}
	return subs, nil
}

// Set makes id the active Azure CLI subscription via `az account set`.
func (c *Client) Set(id string) error {
	if _, err := c.lookPath("az"); err != nil {
		return fmt.Errorf("Azure CLI (az) was not found on PATH.\nInstall it: https://learn.microsoft.com/cli/azure/install-azure-cli\nThen run: az login")
	}

	_, stderr, err := c.exec("az", "account", "set", "--subscription", id)
	if err != nil {
		return wrapAzError(err, stderr)
	}
	return nil
}

// ParseAccountList decodes the JSON array produced by `az account list`.
func ParseAccountList(data []byte) ([]Subscription, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, fmt.Errorf("az account list returned no data.\nIf you are not logged in, run: az login")
	}

	var raw []Subscription
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("could not parse az account list JSON: %w", err)
	}

	subs := make([]Subscription, 0, len(raw))
	for _, s := range raw {
		if strings.TrimSpace(s.ID) == "" {
			continue
		}
		if s.Name == "" {
			s.Name = s.ID
		}
		subs = append(subs, s)
	}
	return subs, nil
}

// ShortID returns the last 8 hex characters of a subscription GUID.
func ShortID(id string) string {
	compact := strings.ReplaceAll(id, "-", "")
	if len(compact) <= 8 {
		return compact
	}
	return compact[len(compact)-8:]
}

func wrapAzError(err error, stderr []byte) error {
	msg := strings.TrimSpace(string(stderr))
	combined := strings.ToLower(msg + " " + err.Error())
	if strings.Contains(combined, "az login") ||
		strings.Contains(combined, "please run") ||
		strings.Contains(combined, "not logged in") {
		return fmt.Errorf("you do not appear to be logged in to Azure CLI.\nRun: az login")
	}
	if msg != "" {
		return fmt.Errorf("az failed: %s", msg)
	}
	return fmt.Errorf("az failed: %w", err)
}
