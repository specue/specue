package stakeholders

// Stakeholder - the owner of the drive/interest
// used for match decisions with
// May be a humen/group/tole/team/system etc
#Stakeholder: {
	id:    string
	title: string
}

// The driver, or interest, or need
// of the stakeholder show who and what wants
// for next linking with the decision
// for irrational reasoning it
#Driver: {
	owner: #Stakeholder
	want:  string
}

// Context extension: drivers live in the decision's context.
// Apply to `context`, NOT to the decision — so the profile's openness
// does not reopen the decision's closed contract.
//   context: sh.#WithDrivers & {drivers: [...]}
#WithDrivers: {
	drivers: [#Driver, ...] & [_, ...]
	// Open for other context profiles to add their own keys.
	...
}
