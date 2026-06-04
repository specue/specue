package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// hintRef matches a backtick-quoted command suggestion like `specue work add …`
// or `specue init …`, capturing the first verb after the program name. The verb
// starts with a letter, so a flag (`specue --help`) is not mistaken for a verb.
var hintRef = regexp.MustCompile("`specue ([a-z][a-z-]*)")

// TestHintedCommandsExist guards the recurring failure where an error's fix or a
// help text names a command that does not exist (e.g. `specue new` after new was
// never built, or a renamed verb). It scans the package source for backtick command
// suggestions and asserts every first verb is a real registered command — so a
// stale hint fails the build, not the user.
func TestHintedCommandsExist(t *testing.T) {
	verbs := registeredVerbs(t)

	files, err := filepath.Glob("*.go")
	require.NoError(t, err)
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, err := os.ReadFile(f)
		require.NoError(t, err)
		for _, m := range hintRef.FindAllSubmatch(src, -1) {
			verb := string(m[1])
			assert.Containsf(t, verbs, verb,
				"%s: hint references `specue %s`, which is not a registered command (verbs: %v)", f, verb, sortedKeys(verbs))
		}
	}
}

// TestCommandConstantsResolve guards the other half of the name/usage tie: every
// name constant must resolve to a real command via findCommand. If a command's
// `Use:` first word is renamed without updating the constant, usage()/cmdPath()
// would silently fall back to a bare string — this catches that.
func TestCommandConstantsResolve(t *testing.T) {
	cases := [][]string{
		{cmdValidate}, {cmdGet}, {cmdDescribe}, {cmdDiff}, {cmdPlan}, {cmdInit}, {cmdContext},
		{cmdPlan, subList},
		{cmdContext, subCreate}, {cmdContext, subUse}, {cmdContext, subList},
		{cmdContext, subModule, subAdd}, {cmdContext, subModule, subList},
	}
	for _, path := range cases {
		assert.NotNilf(t, findCommand(path...), "command path %v does not resolve — a constant drifted from its Use:", path)
	}
}

// registeredVerbs builds the root command and returns the set of top-level command
// names (and aliases) actually wired up.
func registeredVerbs(t *testing.T) map[string]bool {
	t.Helper()
	var g Globals
	var out, errb bytes.Buffer
	code := exitOK
	root := newRootCmd(&g, &out, &errb, &code)
	verbs := map[string]bool{}
	for _, c := range root.Commands() {
		verbs[c.Name()] = true
		for _, a := range c.Aliases {
			verbs[a] = true
		}
	}
	return verbs
}

func sortedKeys(m map[string]bool) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}
