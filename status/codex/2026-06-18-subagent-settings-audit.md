# Subagent Task: Settings/Audit/Developer Surfaces Explorer

- Agent: `019ed7de-6e4f-79e3-90d3-49e95e19167e`
- Nickname: Schrodinger
- Status: completed
- Outcome: Read-only inspection found bounded settings/audit/integration gaps and security risks.

## Key Findings

- Backend tenant settings API exists, but the Settings dashboard page is still static.
- Audit listing exists, but sensitive operations do not consistently emit audit events.
- Settings sidebar advertises Audit Logs and Developer/API-key surfaces with `#` links.
- API-key management is absent; full generate/rotate/revoke is not bounded without schema and secret-store work.
- Unified integrations callback lacks state validation, and provider fallback can be spoofed when Unified credentials are unavailable.
- Two Unified webhook paths behave differently; one starts Temporal workflows without tenant connection verification.
- Role handling is inconsistent across local auth, handlers, and RBAC middleware.

## Main Follow-up

- Wire Settings page to `GET/PUT /api/v1/settings/tenant`.
- Emit audit rows from settings and integrations mutations.
- Normalize RBAC role comparisons case-insensitively.
- Tighten Unified provider fallback outside production only.
- Keep API-key CRUD out of this slice until schema and Secret Manager strategy exist.
