package cli

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//specue:test:list-resources#kinds-listed-without-prior-knowledge
func TestGetListsResourcesWithoutSpecTree(t *testing.T) {
	// A bare `get` is discovery — it must work with no spec tree (run from a dir
	// that has none) so an agent can learn the resource set unconditionally.
	out, _, code := run("get", "-C", t.TempDir())
	assert.Equal(t, exitOK, code)
	assert.Contains(t, out, "usecase")
	assert.Contains(t, out, "need")
}

func TestGetResourcesJSON(t *testing.T) {
	out, _, code := run("get", "--json", "-C", t.TempDir())
	require.Equal(t, exitOK, code)

	var got struct {
		Resources []struct {
			Name    string   `json:"name"`
			Aliases []string `json:"aliases"`
			Type    string   `json:"type"`
		} `json:"resources"`
	}
	require.NoError(t, json.Unmarshal([]byte(out), &got))
	assert.Len(t, got.Resources, 7)
	assert.Equal(t, "usecase", got.Resources[0].Name)
	assert.Equal(t, []string{"uc"}, got.Resources[0].Aliases)
}

//specue:test:list-resources#nodes-of-a-kind
func TestGetFiltersByType(t *testing.T) {
	out, _, code := run("get", "usecase", "-C", walletSpec)
	require.Equal(t, exitOK, code)
	assert.Contains(t, out, "validate-graph")
	assert.Contains(t, out, "asserted")
}

func TestGetAliasResolves(t *testing.T) {
	byName, _, _ := run("get", "usecase", "-C", walletSpec, "--json")
	byAlias, _, _ := run("get", "uc", "-C", walletSpec, "--json")
	assert.Equal(t, byName, byAlias, "alias `uc` selects the same nodes as `usecase`")
}

func TestGetAllSpansTypes(t *testing.T) {
	out, _, code := run("get", "all", "-C", walletSpec)
	require.Equal(t, exitOK, code)
	// the TYPE column appears and nodes of different types are all listed
	assert.Contains(t, out, "TYPE")
	assert.Contains(t, out, "UseCase")
	assert.Contains(t, out, "Port")
	assert.Contains(t, out, "Container")
}

func TestGetAllJSONCarriesType(t *testing.T) {
	out, _, code := run("get", "all", "-C", walletSpec, "--json")
	require.Equal(t, exitOK, code)

	var got struct {
		Resource string `json:"resource"`
		Rows     []struct {
			Type string `json:"type"`
		} `json:"rows"`
	}
	require.NoError(t, json.Unmarshal([]byte(out), &got))
	assert.Equal(t, "all", got.Resource)
	require.NotEmpty(t, got.Rows)
	for _, r := range got.Rows {
		assert.NotEmpty(t, r.Type, "every all-row carries its type")
	}
}

func TestGetUnknownResourceIsActionable(t *testing.T) {
	_, errOut, code := run("get", "widget", "-C", walletSpec)
	assert.Equal(t, exitUsage, code)
	assert.Contains(t, errOut, "try:")
	assert.Contains(t, errOut, "usecase", "the fix lists the valid resources")
}

func TestGetSingleNodeNotFoundIsActionable(t *testing.T) {
	_, errOut, code := run("get", "usecase", "specue.test/example@v0:nope", "-C", walletSpec)
	assert.Equal(t, exitUsage, code)
	assert.Contains(t, errOut, "try:")
}
