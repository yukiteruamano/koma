package config

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/yukiteruamano/koma/filesystem"
	"github.com/yukiteruamano/koma/key"
)

func TestDefaultsCoverAllKeys(t *testing.T) {
	if got, want := len(Default), key.DefinedFieldsCount; got != want {
		t.Errorf("Default has %d fields, want %d", got, want)
	}

	for fieldKey, field := range Default {
		if field.Key == "" {
			t.Errorf("field %q has an empty key", fieldKey)
		}
		if field.Value == nil {
			t.Errorf("field %q has a nil default value", fieldKey)
		}
		if field.Description == "" {
			t.Errorf("field %q is missing a description", fieldKey)
		}
	}
}

func TestEveryKeyHasDefault(t *testing.T) {
	for _, field := range Default {
		value := field.Value
		if value == nil {
			t.Errorf("key %q has nil default", field.Key)
		}
	}
}

func TestEnvExposedMatchesDefaults(t *testing.T) {
	if got, want := len(EnvExposed), key.DefinedFieldsCount; got != want {
		t.Errorf("EnvExposed has %d entries, want %d", got, want)
	}

	seen := make(map[string]bool)
	for _, env := range EnvExposed {
		if seen[env] {
			t.Errorf("duplicate env entry %q", env)
		}
		seen[env] = true

		if _, ok := Default[env]; !ok {
			t.Errorf("env entry %q has no default", env)
		}
	}
}

func TestFieldEnvName(t *testing.T) {
	field := Default[key.DownloaderPath]
	env := field.Env()

	if !strings.HasPrefix(env, "KOMA_") {
		t.Errorf("env name %q should start with KOMA_", env)
	}

	if !strings.Contains(env, "DOWNLOADER_PATH") {
		t.Errorf("env name %q should contain the uppercased key", env)
	}
}

func TestFieldPrettyContainsKeyAndDescription(t *testing.T) {
	field := Default[key.DownloaderConcurrency]
	pretty := field.Pretty()

	if !strings.Contains(pretty, field.Key) {
		t.Errorf("Pretty output should contain the key %q, got %q", field.Key, pretty)
	}
	if !strings.Contains(pretty, field.Description) {
		t.Errorf("Pretty output should contain the description, got %q", pretty)
	}
}

func TestResolveAliases(t *testing.T) {
	old := os.Getenv("HOME")
	defer func() { _ = os.Setenv("HOME", old) }()
	_ = os.Setenv("HOME", "/fakehome")

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "bare tilde", in: "~", want: "/fakehome"},
		{name: "tilde path", in: "~/Downloads", want: "/fakehome/Downloads"},
		{name: "env var expansion", in: "$HOME/books", want: "/fakehome/books"},
		{name: "plain absolute", in: "/abs/path", want: "/abs/path"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Set(key.DownloaderPath, tt.in)
			resolveAliases()

			if got := viper.GetString(key.DownloaderPath); got != tt.want {
				t.Errorf("resolved path = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSetupNoConfigFileUsesDefaults(t *testing.T) {
	filesystem.SetMemMapFs()

	if err := Setup(); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// defaults must be wired into viper even without a config file
	if got := viper.GetInt(key.DownloaderConcurrency); got != 4 {
		t.Errorf("default downloader.concurrency = %d, want 4", got)
	}
}
