// Package planning holds the contracts for the intent axis: diff between two
// snapshots, plans as branches, overlaying a plan on the current spec, detecting
// conflicts between plans, and accepting a plan.
package planning

import (
	s "specue.io/schema@v0:spec"
	root "specue.io/service@v0:service"
	agent "specue.io/domain/agent@v0:agent"
	govaud "specue.io/domain/governance@v0:governance"
	gov "specue.io/governance@v0:governance"
)

diffRefs: s.#Contract & {
	slug:        "diff-refs"
	title:       "Report the typed delta between the spec at two versioned points"
	service:     root.specue
	trigger:     "the caller asks for the difference between two refs of a module"
	invariants: [{
		id:   "typed-over-the-spec-graph"
		text: "The delta is over Contracts, Needs, Ports and their elements."
		satisfies: [agent.review.frs."fr-01"]
	}, {
		id:   "every-change-named"
		kind: "returns"
		text: "Each change is labelled added, removed, modified or rewired."
		satisfies: [agent.review.frs."fr-02"]
	}, {
		id:   "two-snapshots"
		text: "The diff is computed between two snapshots produced from the refs the caller named."
	}, {
		id:   "returns-delta-with-refs"
		kind: "returns"
		text: "The delta is returned together with the two refs it was computed against."
	}]
}

registerPlan: s.#Contract & {
	slug:        "register-plan"
	title:       "Open a new Plan as a Plan record plus branches"
	service:     root.specue
	trigger:     "the caller asks to register a new Plan"
	invariants: [{
		id:   "plan-is-a-branch-set"
		text: "A new Plan creates identically-named branches in every module it touches and a Plan record in the governance module pointing at them."
		satisfies: [
			agent.planner.frs."fr-01",
			govaud.decisionKeeper.frs."fr-02",
		]
		decided_by: [gov.adr07PlansAsBranches]
	}, {
		id:   "governance-required"
		kind: "rejects"
		when: "there is no governance module in the current context"
		satisfies: [agent.planner.frs."fr-06"]
		decided_by: [gov.adr07PlansAsBranches]
	}, {
		id:   "no-overwrite"
		kind: "rejects"
		when: "a Plan with that name is already taken"
	}, {
		id:   "from-base-only"
		kind: "rejects"
		when: "registering from any branch other than the landscape's base branch (so a Plan always forks from a known base, never another Plan)"
	}, {
		id:   "record-names-branches"
		kind: "returns"
		text: "The Plan record names the branches it points at and the modules they live in."
	}]
}

usePlan: s.#Contract & {
	slug:        "use-plan"
	title:       "Switch the working tree into a Plan"
	service:     root.specue
	trigger:     "the caller asks to work on a Plan"
	invariants: [{
		id:   "checks-out-every-branch"
		text: "Every module the Plan touches is checked out onto the Plan's branch in a single step."
		decided_by: [gov.adr07PlansAsBranches]
	}, {
		id:   "refuses-on-dirty-tree"
		kind: "rejects"
		when: "any affected module's working tree carries uncommitted changes"
	}, {
		id:   "authoring-lands-on-plan"
		text: "Subsequent authoring lands on the Plan's branches until the caller returns to base."
	}]
}

returnToBase: s.#Contract & {
	slug:        "return-to-base"
	title:       "Leave a Plan and return to the base branch"
	service:     root.specue
	trigger:     "the caller asks to leave the current Plan"
	invariants: [{
		id:   "every-module-returns"
		text: "Every module that was switched into the Plan is checked out back to the base branch."
		decided_by: [gov.adr07PlansAsBranches]
	}, {
		id:   "refuses-on-dirty-tree"
		kind: "rejects"
		when: "any affected module's working tree carries uncommitted changes"
	}, {
		id:   "authoring-lands-on-base"
		text: "Subsequent authoring lands on the base branch until another Plan is used."
	}]
}

dropPlan: s.#Contract & {
	slug:        "drop-plan"
	title:       "Abandon a Plan without accepting it"
	service:     root.specue
	trigger:     "the caller asks to drop a Plan"
	invariants: [{
		id:   "branches-and-record-removed"
		text: "The Plan record is closed and every branch it pointed at is removed."
		decided_by: [gov.adr07PlansAsBranches]
	}, {
		id:   "dropped-until-reregistered"
		text: "Once dropped the Plan cannot be used again under the same name until it is registered again."
	}]
}

pendingOverlay: s.#Contract & {
	slug:        "pending-overlay"
	title:       "Show a Plan against the current spec without switching the working tree"
	service:     root.specue
	trigger:     "the caller asks to view a Plan against the current spec"
	invariants: [{
		id:   "viewed-without-checkout"
		text: "The Plan is projected onto the current spec by reading its branches through git."
		satisfies: [agent.planner.frs."fr-02"]
		decided_by: [gov.adr07PlansAsBranches]
	}, {
		id:   "base-side-read-through-git"
		text: "The base side of the overlay is read through git from the base branch; the overlay is the same regardless of which branch is currently checked out."
		decided_by: [gov.adr07PlansAsBranches]
	}, {
		id:   "overlay-is-a-spec"
		text: "The overlay result is a spec graph with the same shape as the live one, so any read verb works against it."
	}, {
		id:   "returns-overlay-with-refs"
		kind: "returns"
		text: "The overlay is returned with the refs and the modules it composed."
	}]
}

detectConflict: s.#Contract & {
	slug:        "detect-conflict"
	title:       "Report conflicts between two open Plans"
	service:     root.specue
	trigger:     "the caller asks whether two Plans conflict"
	invariants: [{
		id:   "structural-conflict-blocks"
		when: "overlaying both Plans together produces a graph that cannot resolve (a removed node is referenced, the same edge is rewired two ways)"
		text: "the pair is reported as blocking."
		satisfies: [agent.planner.frs."fr-03"]
		decided_by: [gov.adr07PlansAsBranches]
	}, {
		id:   "co-touch-surfaces-for-review"
		when: "two Plans touch the same Contract or Port but both apply cleanly"
		text: "they are reported as advisory for human or agent review, not blocked."
		satisfies: [agent.planner.frs."fr-04"]
	}, {
		id:   "conflict-names-plans-and-node"
		kind: "returns"
		text: "Each conflict names the two Plans, the shared Contract or Port, and whether it is blocking or advisory."
	}]
}

acceptPlan: s.#Contract & {
	slug:        "accept-plan"
	title:       "Apply a Plan to the current spec and close it"
	service:     root.specue
	trigger:     "the caller asks to accept a Plan"
	invariants: [{
		id:   "merge-only-if-valid"
		text: "The Plan is accepted only when overlaying it on the current spec produces a graph that validates."
		satisfies: [agent.planner.frs."fr-05"]
		decided_by: [gov.adr07PlansAsBranches]
	}, {
		id:   "refuses-invalid-merge"
		kind: "rejects"
		when: "overlaying the Plan fails validation, or a merge conflict arises"
		text: "acceptance is refused and the caller is told which gate or conflict blocked it."
		satisfies: [agent.planner.frs."fr-05"]
		decided_by: [gov.adr07PlansAsBranches, gov.adr14OneInvariantKind]
	}, {
		id:   "branches-merged-everywhere"
		text: "Acceptance merges the Plan's branches into the base branch in every module it touches."
		satisfies: [agent.planner.frs."fr-05"]
		decided_by: [gov.adr07PlansAsBranches]
	}, {
		id:   "plan-record-closes"
		text: "Once merged, the Plan's record in the governance module is closed."
		satisfies: [agent.planner.frs."fr-05"]
	}, {
		id:   "works-from-anywhere"
		text: "Acceptance succeeds regardless of which branch the caller is currently on: a repo found on the Plan's branch is switched to base before merging, so the caller does not have to leave the Plan to land it."
	}, {
		id:   "tags-the-landing"
		text: "Acceptance marks the merge commit of every affected repo with a tag named after the Plan, so a reader of git history can enumerate landed Plans without parsing the commit graph."
	}]
}
