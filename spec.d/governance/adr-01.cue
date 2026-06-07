// Package governance holds Specue's architecture decision records: the causal
// layer. An ADR is a code-unbound node; Contract elements cite one via decided_by
// to record why a contract is shaped as it is, kept out of the contract's own text.
//
// One ADR per file (adr-NN.cue); Plan records live in plans/.
package governance

import s "specue.io/schema@v0:spec"

adr01CUENativeResolution: s.#ADR & {
	slug:       "ADR-01"
	title:      "Cross-module references resolve through CUE, not a hand-written resolver"
	status:     "accepted"
	body: """
		The whole module set is stitched into one CUE value tree, and CUE resolves
		every cross-module reference, version pin and visibility rule. The previous
		generation interpreted its own mini-language of string references through a
		hand-written resolver, which became the system's bottleneck. Standing on CUE
		shifts a class of resolution bugs onto a mature implementation and lets the
		compiler do only what CUE cannot — domain constraints (statuses, cycles,
		coverage).
		"""
}
