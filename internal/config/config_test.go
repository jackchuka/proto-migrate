package config

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
	}{
		{
			name: "valid config",
			yaml: `
source: proto/old
target: proto/new
excludes:
  - "*.tmp"
rules:
  - kind: package
    from: old.v1
    to: new.v1
`,
		},
		{
			name: "missing source",
			yaml: `
target: proto/new
rules:
  - kind: package
    from: old.v1
    to: new.v1
`,
			wantErr: true,
		},
		{
			name: "invalid rule kind",
			yaml: `
source: proto/old
target: proto/new
rules:
  - kind: unknown
    from: old.v1
    to: new.v1
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.yaml)
			_, err := parse(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRuleValidate(t *testing.T) {
	tests := []struct {
		name    string
		rule    Rule
		wantErr bool
	}{
		{
			name: "valid package rule",
			rule: Rule{Kind: "package", From: "old", To: "new"},
		},
		{
			name:    "package rule missing from",
			rule:    Rule{Kind: "package", To: "new"},
			wantErr: true,
		},
		{
			name: "valid regexp rule",
			rule: Rule{Kind: "regexp", Pattern: "old", Replace: "new"},
		},
		{
			name:    "regexp rule missing pattern",
			rule:    Rule{Kind: "regexp", Replace: "new"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
