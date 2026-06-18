# Subagent Task: Dashboard UI Explorer

- Agent: `019ed7ce-e165-78d0-8d96-2b28daa6f875`
- Nickname: Archimedes
- Status: completed
- Outcome: Read-only inspection found static and partially stubbed UI across in-scope domains.

## Key Findings

- Orders page and shipping page use hard-coded arrays, non-wired filters/search/pagination/actions, and fixed KPI counts.
- Orders left sidebar links to missing routes and has configuration links pointing to `#`.
- Orders right sidebar is not contextual and always shows the same sample order.
- Channels page and sidebar are static; channel actions, filters, KPIs, and right sidebar are not API-driven.
- Teams page, left sidebar, and right sidebar are fully static.
- Settings general page is static and has no load/save path.
- Settings organization page is partially real via `/api/v1/organization/members`.
- Settings integrations page uses `setTimeout`, empty mock connections, and a hard-coded Unified workspace ID.
- Settings callback posts the connection ID, but the file has a syntax defect discovered separately by the main agent build run.
- Settings sidebar has multiple `href="#"` route gaps.

## Suggested Patches Captured

- Add backend GET/list/detail endpoints for OMS orders and sales channels.
- Add React Query hooks for orders/channels and replace static tables where possible.
- Wire row selection into right sidebars.
- Add settings GET/PUT endpoints and wire the General Settings form.
- Add `GET /api/v1/integrations/connections` and remove the integrations mock timeout.
