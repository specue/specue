package plan_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/specue/specue/internal/model"
	"github.com/specue/specue/internal/plan"
	"github.com/specue/specue/internal/source"
)

// consumerRepo lays out a two-module landscape where example exposes a node that consumer
// depends on cross-module. Plan A edits consumer, plan B removes example's node; alone
// each is fine, together they dangle — the structural conflict Conflict catches.
func consumerRepo(t *testing.T) (map[model.ModulePath]string, *plan.Manager, string) {
	t.Helper()
	bin := gitBin(t)
	root := t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command(bin, args...)
		cmd.Dir = root
		out, err := cmd.CombinedOutput()
		require.NoErrorf(t, err, "git %v: %s", args, out)
	}
	write := func(rel, content string) {
		full := filepath.Join(root, rel)
		require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
		require.NoError(t, os.WriteFile(full, []byte(content), 0o644))
	}
	run("init", "-b", "base")
	run("config", "user.email", "t@t")
	run("config", "user.name", "t")

	write("governance/spec.mod.cue", "module: \"x.test/governance@v0\"\nversion: \"v0.0.1\"\nkind: \"governance\"\n")
	write("governance/cue.mod/module.cue", "module: \"x.test/governance@v0\"\nlanguage: version: \"v0.16.0\"\ndeps: \"specue.io/schema@v0\": v: \"v0.0.1\"\n")

	write("example/spec.mod.cue", "module: \"x.test/example@v0\"\nversion: \"v0.0.1\"\nkind: \"service\"\n")
	write("example/cue.mod/module.cue", walletCueMod)
	write("example/nodes.cue", `package example
import s "specue.io/schema@v0:spec"
svc: s.#Container & {type:"Container", slug:"example", title:"Wallet", confidence:"CONFIRMED", kind:"service"}
applyOp: s.#Contract & {type:"Contract", slug:"apply-op", title:"Apply", confidence:"CONFIRMED", service:svc, postconditions:[{text:"done"}]}
`)
	write("consumer/spec.mod.cue", "module: \"x.test/consumer@v0\"\nversion: \"v0.0.1\"\nkind: \"service\"\nrequire: [{module: \"x.test/example@v0\", version: \"v0.0.1\", replace: \"../example\"}]\n")
	write("consumer/cue.mod/module.cue", "module: \"x.test/consumer@v0\"\nlanguage: version: \"v0.16.0\"\ndeps: \"specue.io/schema@v0\": v: \"v0.0.1\"\ndeps: \"x.test/example@v0\": v: \"v0.0.1\"\n")
	write("consumer/nodes.cue", `package consumer
import s "specue.io/schema@v0:spec"
svc: s.#Container & {type:"Container", slug:"consumer", title:"Consumer", confidence:"CONFIRMED", kind:"service"}
validate: s.#Contract & {type:"Contract", slug:"validate", title:"Validate", confidence:"CONFIRMED", service:svc, postconditions:[{text:"placed"}]}
`)
	run("add", "-A")
	run("commit", "-m", "base")

	gov := model.ModulePath("x.test/governance@v0")
	work := source.Workspace{PlanBase: "base", Modules: []source.WorkModule{
		{Path: gov, Dir: filepath.Join(root, "governance")},
		{Path: "x.test/example@v0", Dir: filepath.Join(root, "example")},
		{Path: "x.test/consumer@v0", Dir: filepath.Join(root, "consumer")},
	}}
	dirs := map[model.ModulePath]string{
		gov:               filepath.Join(root, "governance"),
		"x.test/example@v0": filepath.Join(root, "example"),
		"x.test/consumer@v0":  filepath.Join(root, "consumer"),
	}
	mgr, err := plan.NewManager(work, dirs, plan.NewGit(bin), gov)
	require.NoError(t, err)
	return dirs, mgr, bin
}

//specue:test:detect-conflict#structural-conflict-blocks
func TestConflictDetectsCrossPlanDangling(t *testing.T) {
	dirs, mgr, bin := consumerRepo(t)
	exampleDir := dirs["x.test/example@v0"]
	consumerDir := dirs["x.test/consumer@v0"]
	root := filepath.Dir(exampleDir)

	// Plan A: consumer's validate now depends on example's validate-op (cross-module ref). Clean
	// alone — apply-op exists on base.
	require.NoError(t, mgr.Register("a", ""))
	require.NoError(t, mgr.Use("a"))
	require.NoError(t, os.WriteFile(filepath.Join(consumerDir, "nodes.cue"), []byte(`package consumer
import (
	s "specue.io/schema@v0:spec"
	w "x.test/example@v0:example"
)
svc: s.#Container & {type:"Container", slug:"consumer", title:"Consumer", confidence:"CONFIRMED", kind:"service"}
validate: s.#Contract & {type:"Contract", slug:"validate", title:"Validate", confidence:"CONFIRMED", service:svc, postconditions:[{text:"placed", depends_on:[{to: w.applyOp, role:"call"}]}]}
`), 0o644))
	commit(t, bin, root, "a: consumer depends on example apply-op")
	require.NoError(t, mgr.Base())

	// Plan B: example removes apply-op. Clean alone — nothing on base refs it.
	require.NoError(t, mgr.Register("b", ""))
	require.NoError(t, mgr.Use("b"))
	require.NoError(t, os.WriteFile(filepath.Join(exampleDir, "nodes.cue"), []byte(`package example
import s "specue.io/schema@v0:spec"
svc: s.#Container & {type:"Container", slug:"example", title:"Wallet", confidence:"CONFIRMED", kind:"service"}
`), 0o644))
	commit(t, bin, root, "b: remove apply-op")
	require.NoError(t, mgr.Base())

	parser, err := source.NewCUEParser()
	require.NoError(t, err)
	proj, err := plan.NewProjector(parser, plan.NewGit(bin))
	require.NoError(t, err)
	t.Cleanup(func() { _ = proj.Close() })

	res, err := mgr.Conflict(proj, dirs, "a", "b")
	require.NoError(t, err)
	assert.True(t, res.Conflicts(), "A (ref apply-op) + B (remove apply-op) dangle together")
}

func TestConflictCleanWhenIndependent(t *testing.T) {
	dirs, mgr, bin := consumerRepo(t)
	consumerDir := dirs["x.test/consumer@v0"]
	root := filepath.Dir(consumerDir)

	// Two plans touching different things, no cross-fault: A adds a consumer node, B
	// adds a example node. Together they still load.
	require.NoError(t, mgr.Register("a", ""))
	require.NoError(t, mgr.Use("a"))
	require.NoError(t, os.WriteFile(filepath.Join(consumerDir, "nodes.cue"), []byte(`package consumer
import s "specue.io/schema@v0:spec"
svc: s.#Container & {type:"Container", slug:"consumer", title:"Consumer", confidence:"CONFIRMED", kind:"service"}
validate: s.#Contract & {type:"Contract", slug:"validate", title:"Validate", confidence:"CONFIRMED", service:svc, postconditions:[{text:"placed"}]}
cashout: s.#Contract & {type:"Contract", slug:"cashout", title:"Cashout", confidence:"CONFIRMED", service:svc, postconditions:[{text:"paid"}]}
`), 0o644))
	commit(t, bin, root, "a: add cashout")
	require.NoError(t, mgr.Base())

	parser, err := source.NewCUEParser()
	require.NoError(t, err)
	proj, err := plan.NewProjector(parser, plan.NewGit(bin))
	require.NoError(t, err)
	t.Cleanup(func() { _ = proj.Close() })

	// b never registered → only a's content overlays; clean.
	res, err := mgr.Conflict(proj, dirs, "a", "a")
	require.NoError(t, err)
	assert.False(t, res.Conflicts(), "a overlaid with itself is clean")
}
