# Specification And Roadmap Readiness Checklist

Use this checklist before running an implementation cycle.

## Specification

- [ ] Status is `Accepted`.
- [ ] Scope and exclusions are explicit.
- [ ] The selected capability has a stable `CAP-*` identifier.
- [ ] Inputs and preconditions are exact.
- [ ] Success is observable.
- [ ] Output or response semantics are exact.
- [ ] Side effects and forbidden effects are exact.
- [ ] Negative behavior and errors are exact.
- [ ] Lifecycle, cleanup, timeout, or ordering rules are defined when relevant.
- [ ] Security and data-safety invariants are exact.
- [ ] Compatibility expectations are stated.
- [ ] Acceptance criteria use stable `AC-*` identifiers.
- [ ] Criteria are binary and testable.
- [ ] The production acceptance boundary is named.
- [ ] No blocking open question remains.

## Roadmap

- [ ] Status is `Accepted`.
- [ ] The source specification is the accepted version.
- [ ] The next slice has one mode.
- [ ] The next slice references exact `CAP-*` and `AC-*` identifiers.
- [ ] The slice produces one observable result.
- [ ] One primary acceptance boundary is named.
- [ ] Useful red or baseline-green evidence is defined.
- [ ] Focused verification is defined or has a deterministic discovery rule.
- [ ] A real-path exercise is defined.
- [ ] Broad verification is defined.
- [ ] Dependencies are explicit and satisfied.
- [ ] In-scope and out-of-scope work are explicit.
- [ ] Risks and stop conditions are explicit.
- [ ] Exit gate is binary.
- [ ] The slice is not horizontal scaffolding.
- [ ] The slice fits one reviewable implementation cycle.
- [ ] The slice keeps required context within roughly the first 30% of a typical
      Codex token window.

## Repository

- [ ] Applicable `AGENTS.md` files are present.
- [ ] `.agents/skills` is installed.
- [ ] Build and test entry points can be discovered.
- [ ] Existing baseline failures are known.
- [ ] The intended production boundary can be exercised.
- [ ] Required services, credentials, permissions, or hardware are available.
- [ ] User-created working-tree changes are identified and preservable.

Any failed applicable item must be resolved or recorded as a blocker before
implementation begins.
