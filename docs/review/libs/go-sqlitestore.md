# go-sqlitestore — adjacent library audit (closeout T24)

**Verdict:** transitive-only, correctly: zero webphone imports — the module rides in through cqrs-htmx (usermgmt v4.18.0, loginpage v4.13.0) and go-cqrs-lite/projectionhost. Webphone's own sqlite story has two deliberate homes (the hand ladder in `internal/store/db.go` for the app database; metaengine's sqliteengine for readmodels.db), and a third store abstraction would be a second home for the same fact.
**Pin:** v0.1.1 (indirect) · **Latest:** v0.1.1 · **Currency:** current at the pin · **Status:** indirect-only
**Audited:** 2026-11-04 · **Prior reports:** none.
**Run all commands from the webphone checkout root** (repo-relative paths).

## Ruling

NO-GO as a direct dependency. The consumers that pull it own their own database lifetimes through it; webphone's schemas, migrations, and probes are already owned by the two named homes. **Refresh trigger:** usermgmt/loginpage exposing a migration surface webphone must configure (then the library becomes an integration seam, not an adoption); a third database appearing in the product.

## How to verify

```console
rg -n "go-sqlitestore" go.mod                        # indirect pin v0.1.1
rg -n "go-sqlitestore" --type go internal/ cmd/ | wc -l   # expect 0 imports
nix develop -c go mod graph | grep -c "go-sqlitestore@"   # expect >= 3 (the transitive consumers)
```
