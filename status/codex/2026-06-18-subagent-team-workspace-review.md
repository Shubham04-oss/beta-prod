# Subagent Task: Team Workspace Implementation Review

- Agent: `019ed7e3-297e-73e1-a012-72e12f14f84b`
- Nickname: Curie
- Status: completed
- Outcome: Read-only review found compile and tenant-consistency risks in the newly added Team Workspace slice.

## Findings

- `handlers_team_workspace.go` needed a `context` import.
- `ai_tasks.team_id` needed composite tenant consistency with `human_ai_teams`.
- Team Workspace RLS should use `FORCE ROW LEVEL SECURITY` if the app role owns the table.
- Workspace reads and task creation should enforce both tenant and org boundaries.
- Missing or inaccessible teams should return `404`, not `500`.

## Main Follow-up

- Add composite `(id, tenant_id, org_id)` uniqueness and FK.
- Force RLS on both tables.
- Filter Team Workspace queries by tenant and org.
- Special-case `pgx.ErrNoRows` on task creation.
