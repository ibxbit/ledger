# Design decisions

The reasoning behind the schema. These are the answers you give when a teammate asks "why did we do it this way?"

## 1. Why compute balance from entries instead of a `balance` column?

A `balance` column is a *cached copy* of the truth, and caches drift. If the app crashes between "subtract from sender" and "add to receiver", the column now lies and nothing can prove it. With double-entry, the entries ARE the truth: balance is `SUM(entries.amount)`, every movement has a full audit trail (who, when, which transfer), and the invariant `SUM(all entries) = 0` lets one query detect corruption. If the column and the entries ever disagree, you'd believe the entries anyway — so store only the entries.

(Later, for performance, real systems add a cached balance *on top* — but it's derived, rebuildable, and checked against the entries. Cache as optimization, never as source of truth.)

## 2. Why BIGINT minor units (cents), never FLOAT?

Floats are binary fractions and cannot represent most decimal amounts exactly: `SELECT 0.1::float8 + 0.2::float8` returns `0.30000000000000004`. Tiny errors compound across millions of transactions, and money that "almost" balances is money that is wrong. Integers are exact: $10.50 is stored as `1050` cents. BIGINT tops out around 92 quadrillion cents — enough for any realistic system. Rendering `1050` as `$10.50` is the presentation layer's job.

## 3. What stops a USD→EUR transfer? (trick question)

**Nothing. This is a real hole in the current schema.** `transfers` only checks amount > 0 and from ≠ to; it never compares the two accounts' currencies. A transfer between a USD and a EUR account would happily record 5000 "somethings" moving — silently treating 50 dollars as 50 euros.

Finding holes like this by reading your own schema is the skill. Fix options, in order of strength:
1. Enforce in the app layer inside the transfer transaction (weakest — every future code path must remember).
2. A DB trigger comparing currencies (strong, but hidden logic).
3. Redesign so it's structurally impossible (strongest — e.g. currency lives on the transfer and entries must match it).

We'll close this hole in Week 2 when transfers become real transactions — and we'll write a failing test for it first.
