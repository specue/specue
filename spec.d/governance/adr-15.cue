package governance

import s "specue.io/schema@v0:spec"

adr15SchemaHygiene: s.#ADR & {
	slug:       "ADR-15"
	title:      "A Contract's service is a Container, a Need's domain is a Domain, a dep is typed by its role, and binding is dropped"
	status:     "proposed"
	body: """
		The schema accreted three ungated or rudimentary spots, all surfaced while
		dogfooding M-elem.

		Ungated typed refs. `#Contract.service` and `#Need.domain` were typed
		`#Node` (the any-node union), so CUE accepted `service: someADR` and the
		compiler only checked the target resolved, not its type — while everywhere
		else the model gates its edges (`satisfies`→atom, binding→Contract). Tighten
		the CUE: `service!: #Container`, `domain!: #Domain`. (`#Port.schema` stays
		`#Node` for now — there is no dedicated IDL node type to point it at.)

		A dep ignored its own role. A plain dep (no `role`) is a logical dependency
		and should target a `#Contract`; an infra dep (`role` set) is a physical
		touch and should target a `#Port | #Container`. Both were `#Node`, and a
		mis-typed infra target was silently skipped by topology. This is expressible
		via the sibling-field idiom (the `#Port if kind ==` pattern): gate `to` on
		whether `role` is set. The `#invariant` option (G2, element-grained deps) is
		preserved on the plain branch.

		`binding` is a rudiment — dropped. The field offered required|optional|
		abstract, but `optional` was dead (treated as `required`) and `abstract` was
		authored by no node — only one synthetic test. Its single effect was one
		fixpoint branch: an abstract contract never blocks and is never a gap. But a
		contract with no promise anyone keeps is not a Contract — everything
		speculative is already its own node type (ADR, Plan, with status proposed).
		Drop `binding`; every Contract without code is now an honest gap. (A closed
		enum exists only when the tool reasons about it; after removing the abstract
		branch there is nothing left to reason about.)

		Breaking across schema, model, source, fixpoint, jsonir, diff and describe;
		ships in the pre-release window. Closes M4.
		"""
}
