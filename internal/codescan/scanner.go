package codescan

import (
	"io/fs"
	"strings"

	"github.com/specue/specue/internal/compiler"
	"github.com/specue/specue/internal/model"
)

// Scanner gathers code facts from source trees. An interface, like the other
// layers, so callers depend on the capability and can substitute it in tests.
type Scanner interface {
	Scan(targets []ScanTarget) ([]compiler.CodeFact, error)
}

// ScanTarget is one source tree to scan: the filesystem to read, the dir within it
// to walk, and the module attribution. There is no per-target scanner kind — a code
// module is multilingual (P20), so the scanner picks how to read each FILE by its
// own extension (langOf) and whether it is a test by its NAME (isTestPath). Module
// is the module whose code this is; Candidates are sibling modules the same source
// also implements (a deploy repo serving several), letting a bare annotation fall
// back to whichever owns the slug.
//
// Files optionally restricts the walk to an explicit set (relative to Root) — this
// is how a caller feeds `git ls-files` output so only tracked files are scanned
// (untracked / .gitignore'd trees never reached). Empty Files = walk the whole tree.
type ScanTarget struct {
	FS         fs.FS
	Root       string // dir within FS to walk ("." for the whole tree)
	Files      []string
	Module     model.ModulePath
	Candidates []model.ModulePath
}

type scanner struct{}

// NewScanner returns a Scanner.
func NewScanner() Scanner { return scanner{} }

// Scan reads every target and returns the code facts found.
//
//specue:req:scan-code
func (scanner) Scan(targets []ScanTarget) ([]compiler.CodeFact, error) {
	var facts []compiler.CodeFact
	for _, t := range targets {
		tf, err := scanTarget(t)
		if err != nil {
			return nil, err
		}
		facts = append(facts, tf...)
	}
	return facts, nil
}

// scanTarget reads one target's files and extracts facts. An explicit Files list
// (git ls-files) is scanned directly; otherwise the whole tree under Root is walked.
// Either way only scannable source files (a known language) are read.
func scanTarget(t ScanTarget) ([]compiler.CodeFact, error) {
	if len(t.Files) > 0 {
		return scanListed(t)
	}
	return scanWalk(t)
}

// scanListed reads exactly the files the caller named (relative to Root).
func scanListed(t ScanTarget) ([]compiler.CodeFact, error) {
	var facts []compiler.CodeFact
	for _, rel := range t.Files {
		path := model.FilePath(rel)
		if !IsScannable(path) {
			continue
		}
		raw, err := fs.ReadFile(t.FS, joinFS(t.Root, rel))
		if err != nil {
			return nil, err
		}
		facts = append(facts, scanFile(path, raw, t.Module, t.Candidates)...)
	}
	return facts, nil
}

// scanWalk walks the whole tree under Root, reading every scannable file.
func scanWalk(t ScanTarget) ([]compiler.CodeFact, error) {
	var facts []compiler.CodeFact
	err := fs.WalkDir(t.FS, t.Root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		path := model.FilePath(p)
		if d.IsDir() || !IsScannable(path) {
			return nil
		}
		raw, err := fs.ReadFile(t.FS, p)
		if err != nil {
			return err
		}
		facts = append(facts, scanFile(path, raw, t.Module, t.Candidates)...)
		return nil
	})
	return facts, err
}

// quotedInsideComment reports whether the annotation at start is prose inside an
// earlier comment rather than a comment of its own. The discriminator is what the
// line itself opens with: if the trimmed line already starts with a comment marker
// (// or #), the comment began earlier and this annotation is text within it — the
// doc-comment-mentioning-syntax case (`// … //specue:req:slug …`). If the line
// starts with code, the first marker on it opens a real trailing annotation, even
// when an earlier // sits inside a string literal (url := "http://x" //specue:…).
// So an annotation is quoted iff the line is already a comment AND the annotation
// is not the very marker that opens it.
//
//specue:req:scan-code#ignored-by-comment-context
func quotedInsideComment(line string, start int) bool {
	trimmed := strings.TrimLeft(line, " \t")
	if !strings.HasPrefix(trimmed, "//") && !strings.HasPrefix(trimmed, "#") {
		return false // line starts with code → first marker is a real (trailing) annotation
	}
	// Line is itself a comment. The annotation is genuine only if it sits exactly
	// where that comment opens; anything further right is prose inside it.
	openAt := len(line) - len(trimmed)
	return start != openAt
}

// joinFS joins an fs.FS-style path ("." root means the rel as-is).
func joinFS(root, rel string) string {
	if root == "" || root == "." {
		return rel
	}
	return root + "/" + rel
}

// scanFile extracts every annotation in one file into a CodeFact. Test context is
// decided by the file's name (isTestPath), independent of language.
func scanFile(path model.FilePath, raw []byte, mod model.ModulePath, candidates []model.ModulePath) []compiler.CodeFact {
	isTest := isTestPath(path)
	var facts []compiler.CodeFact
	for i, line := range strings.Split(string(raw), "\n") {
		for _, idx := range annRe.FindAllStringSubmatchIndex(line, -1) {
			// idx: [matchStart matchEnd v0 v1 ref0 ref1]. Skip an annotation that is
			// not itself the comment but text quoted inside an earlier one — e.g. a
			// doc comment mentioning `//specue:req:slug` as syntax. The real
			// annotation's comment marker opens the comment; if another marker sits
			// to its left on the line, it is already inside a comment, so it is prose.
			if quotedInsideComment(line, idx[0]) {
				continue
			}
			ra := rawAnnotation{verb: line[idx[2]:idx[3]], ref: line[idx[4]:idx[5]], file: path, line: i + 1, isTest: isTest}
			if f, ok := toFact(ra, mod, candidates); ok {
				facts = append(facts, f)
			}
		}
	}
	return facts
}

// toFact parses a raw annotation and stamps it with the module attribution.
func toFact(ra rawAnnotation, mod model.ModulePath, candidates []model.ModulePath) (compiler.CodeFact, bool) {
	verb, target, ok := ra.parse()
	if !ok {
		return compiler.CodeFact{}, false
	}
	return compiler.CodeFact{
		Module:     mod,
		Candidates: candidates,
		Verb:       verb,
		Target:     target,
		File:       ra.file,
		Line:       ra.line,
		IsTest:     ra.isTest,
	}, true
}
