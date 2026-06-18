# Subagent Task: Settings UI Wiring Review

- Agent: `019ed7e3-3bea-7481-b470-30aeb4c3535d`
- Nickname: Dalton
- Status: completed
- Outcome: Read-only review confirmed the General Settings page should bind four writable fields to `/api/v1/settings/tenant`.

## Writable Fields

- `inventory_allocation_model`: `HARD` or `SOFT`
- `auto_po_enabled`: boolean
- `default_low_stock_threshold`: non-negative integer
- `costing_method`: `WAC` or `FIFO`

## Risks Found

- `dashboard/src/lib/api.ts` uses `baseUrl`, but `ky` expects `prefixUrl`.
- `fetchAPI` attempts `.json()` even for plain-text backend errors.
- Some settings sidebar imports are unused.

## Main Follow-up

- Fix API client `prefixUrl` and error handling.
- Keep Settings page bound to the four writable backend fields.
