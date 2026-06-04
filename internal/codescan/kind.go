package codescan

import (
	"strings"

	"github.com/specue/specue/internal/model"
)

// A code module is multilingual (MANIFESTO P20 / code-as-module): one module may
// hold Go, proto, k8s manifests, TS, all at once. So the scanner picks how to read
// a file by the FILE, not by a per-module kind — fileKind matches each path to the
// first language whose extension it carries, and isTestPath decides verification
// context by NAME, independent of language.

func hasSuffix(p model.FilePath, suffix string) bool { return strings.HasSuffix(string(p), suffix) }
func contains(p model.FilePath, sub string) bool     { return strings.Contains(string(p), sub) }

// fileLang names the comment family a file uses; the matcher (annRe) is already
// language-agnostic, so this exists only to decide which files are scannable at all
// (skip binaries, assets) and to anchor future per-language nuance. A file with no
// known language is not scanned.
type fileLang int

const (
	langNone fileLang = iota
	langGo
	langProto
	langYAML // k8s / helm manifests
	langTS   // TS/JS — React SPA or k6 scripts (disambiguated by name, not here)
)

// fileLang returns the language of a path by extension, langNone if unscannable.
func langOf(p model.FilePath) fileLang {
	switch {
	case hasSuffix(p, ".go"):
		return langGo
	case hasSuffix(p, ".proto"):
		return langProto
	case hasSuffix(p, ".yaml"), hasSuffix(p, ".yml"), hasSuffix(p, ".tpl"):
		return langYAML
	case hasSuffix(p, ".ts"), hasSuffix(p, ".tsx"), hasSuffix(p, ".js"):
		return langTS
	}
	return langNone
}

// IsScannable reports whether a path is a source file the scanner reads at all
// (a known language). Exported so the engine's content-key tracks the same set the
// scanner reads.
func IsScannable(p model.FilePath) bool { return langOf(p) != langNone }

// isTestPath reports whether a file is a verification (test) context — a covering
// test proves a contract, ordinary code only implements it. Decided by NAME across
// languages (not by a per-module kind, which broke down once a module mixes a SPA
// and k6 load scripts in the same .ts extension): the conventional test suffixes
// plus a /test/ or /tests/ path segment.
func isTestPath(p model.FilePath) bool {
	switch {
	case hasSuffix(p, "_test.go"):
		return true
	case hasSuffix(p, ".test.ts"), hasSuffix(p, ".test.tsx"),
		hasSuffix(p, ".spec.ts"), hasSuffix(p, ".spec.tsx"),
		hasSuffix(p, ".test.js"), hasSuffix(p, ".spec.js"):
		return true
	case contains(p, "/test/"), contains(p, "/tests/"), contains(p, "/__tests__/"):
		return true
	}
	return false
}
