# Fixes Applied During Full-Scale Chrome E2E

## PostgreSQL delivery scale

Symptom: the 8,000-row remote PostgreSQL upsert stalled because conflict-aware delivery was effectively issuing one insert per row.

Fix:

- Compile bounded multi-row PostgreSQL `INSERT ... ON CONFLICT` statements.
- Cap delivery chunks at 1,000 rows and reduce them automatically for very wide
  SaaS tables so each statement remains below PostgreSQL's bind-parameter limit.

Result: the initial 8,000-row load and the 2,025-row incremental upsert both completed successfully.

## Authoritative SaaS catalog caching

Symptom: Stripe and HubSpot source discovery succeeded in the UI, but transformation validation could not build source schema hints from the discovered catalog.

Fix: persist authoritative Stripe/HubSpot ELT discovery metadata in the control-plane schema cache used by validation.

Result: source-aware validation, preview, and publishing worked for the discovered SaaS streams.

## Stripe dynamic fields

Symptom: an unrestricted account projection contained dynamic runtime properties outside the pre-created destination contract.

Fix: publish an explicit stable account projection.

Result: all 34 streams executed and the representative destination model delivered successfully.

## HubSpot strict coverage and PostgreSQL compatibility

Symptoms:

- One model could not satisfy the one-model-per-selected-stream invariant.
- Some HubSpot property names exceeded PostgreSQL's 63-character identifier limit.
- Sparse runtime records did not always contain every discovered property.
- Pipeline-stage resources did not expose a synthetic `id`.

Fixes:

- Create, validate, preview, and publish one model/table for each of the 10 selected streams.
- Use a sparse-safe compatible-name projection for companies, deals, and quotes.
- Use `(pipeline_id, stage_id)` composite keys for deal and ticket pipeline stages.

Result: the final 10-stream HubSpot run delivered 17 rows successfully.

All diagnostic failures and successful reruns remain retained. No secrets are included.
