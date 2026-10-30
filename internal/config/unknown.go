package config

import (
	"log/slog"
	"reflect"
	"sort"
	"strings"
)

// unknownKeyWarnCap bounds the WARN line: a config that drifted badly
// must not turn one log line into a novel. The full count is logged
// beside the shown keys.
const unknownKeyWarnCap = 20

// schemaPaths walks the Config type and returns every dotted koanf
// path the schema accepts, plus the subset of paths whose schema type
// is a map — those carry DYNAMIC keys (identities by extension,
// passkey users by email), so their children can never be statically
// known and must not be flagged.
func schemaPaths() (known map[string]bool, dynamic map[string]bool) {
	known = map[string]bool{}
	dynamic = map[string]bool{}
	var walk func(t reflect.Type, parent string)
	walk = func(t reflect.Type, parent string) {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			tag := field.Tag.Get("koanf")
			if tag == "" || tag == "-" {
				continue
			}
			if comma := strings.Index(tag, ","); comma >= 0 {
				tag = tag[:comma]
			}
			path := tag
			if parent != "" {
				path = parent + "." + tag
			}
			known[path] = true
			ft := field.Type
			if ft.Kind() == reflect.Pointer {
				ft = ft.Elem()
			}
			switch ft.Kind() {
			case reflect.Struct:
				walk(ft, path)
			case reflect.Map:
				dynamic[path] = true
			}
		}
	}
	walk(reflect.TypeFor[Config](), "")
	return known, dynamic
}

// unknownConfigPaths diffs the merged koanf raw map against the schema:
// any path the binary does not know is returned sorted. It is the
// detection half of the silent-pin-gap warning — an older binary fed a
// newer config silently drops every key it lacks (koanf Unmarshal
// ignores unknowns), which is exactly how a deployed capability
// declaration evaporates without a trace (the 2026-10-30 fax train).
// Dynamic-key map children (identities, passkey users) are exempt.
func unknownConfigPaths(raw map[string]any) []string {
	known, dynamic := schemaPaths()
	var unknown []string
	var walk func(m map[string]any, parent string)
	walk = func(m map[string]any, parent string) {
		for key, val := range m {
			path := key
			if parent != "" {
				path = parent + "." + key
			}
			if !known[path] {
				// The whole subtree under an unknown path is one
				// unknown: reporting each child would bury the root
				// cause (the map the binary predates) in noise.
				unknown = append(unknown, path)
				continue
			}
			if child, ok := val.(map[string]any); ok && !dynamic[path] {
				walk(child, path)
			}
		}
	}
	walk(raw, "")
	sort.Strings(unknown)
	return unknown
}

// warnUnknownConfigKeys emits the one WARN listing unknown keys (key
// PATHS only — never values, which include secrets) with the shown
// list capped at unknownKeyWarnCap. Zero unknown keys log nothing, so
// a clean boot stays byte-identical in the journal.
func warnUnknownConfigKeys(raw map[string]any) {
	unknown := unknownConfigPaths(raw)
	if len(unknown) == 0 {
		return
	}
	shown := unknown
	if len(shown) > unknownKeyWarnCap {
		shown = shown[:unknownKeyWarnCap]
	}
	slog.Warn(
		"unknown configuration keys are ignored — a newer config on an older binary shows up here instead of silently vanishing",
		"keys", strings.Join(shown, ", "),
		"total", len(unknown),
	)
}
