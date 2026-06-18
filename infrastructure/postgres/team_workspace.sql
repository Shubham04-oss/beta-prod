-- Team Workspace schema for human-managed AI squads and delegated tasks.

CREATE TABLE IF NOT EXISTS human_ai_teams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    owner_user_id UUID REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    cadence_minutes INTEGER NOT NULL DEFAULT 30 CHECK (cadence_minutes > 0),
    approval_policy JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'human_ai_teams_id_tenant_org_unique'
    ) THEN
        ALTER TABLE human_ai_teams
            ADD CONSTRAINT human_ai_teams_id_tenant_org_unique UNIQUE (id, tenant_id, org_id);
    END IF;
END $$;

ALTER TABLE human_ai_teams ENABLE ROW LEVEL SECURITY;
ALTER TABLE human_ai_teams FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_human_ai_teams ON human_ai_teams;
CREATE POLICY tenant_isolation_human_ai_teams ON human_ai_teams
    USING (
        tenant_id = current_setting('app.current_tenant', true)::uuid
        AND org_id = current_setting('app.current_org', true)::uuid
    );

CREATE INDEX IF NOT EXISTS idx_human_ai_teams_tenant ON human_ai_teams(tenant_id, status)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS ai_tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    team_id UUID NOT NULL,
    created_by UUID REFERENCES users(id),
    title VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    priority VARCHAR(50) NOT NULL DEFAULT 'NORMAL',
    input_context JSONB NOT NULL DEFAULT '{}'::jsonb,
    proposed_output JSONB NOT NULL DEFAULT '{}'::jsonb,
    requires_approval BOOLEAN NOT NULL DEFAULT false,
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMP WITH TIME ZONE,
    due_at TIMESTAMP WITH TIME ZONE,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'ai_tasks_team_tenant_org_fk'
    ) THEN
        ALTER TABLE ai_tasks
            ADD CONSTRAINT ai_tasks_team_tenant_org_fk
            FOREIGN KEY (team_id, tenant_id, org_id)
            REFERENCES human_ai_teams(id, tenant_id, org_id)
            ON DELETE CASCADE;
    END IF;
END $$;

ALTER TABLE ai_tasks ENABLE ROW LEVEL SECURITY;
ALTER TABLE ai_tasks FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation_ai_tasks ON ai_tasks;
CREATE POLICY tenant_isolation_ai_tasks ON ai_tasks
    USING (
        tenant_id = current_setting('app.current_tenant', true)::uuid
        AND org_id = current_setting('app.current_org', true)::uuid
    );

CREATE INDEX IF NOT EXISTS idx_ai_tasks_tenant_status ON ai_tasks(tenant_id, status)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ai_tasks_team ON ai_tasks(team_id, created_at DESC)
    WHERE deleted_at IS NULL;
