-- Sample data for HelloWorld application with container, database, and cache components
-- Run this after the main migrations to populate helloworld demo data

-- Insert helloworld application
INSERT INTO apps (name, description)
VALUES ('helloworld', 'Simple HelloWorld application demonstrating container, database, and cache orchestration')
ON CONFLICT (name) DO NOTHING;

-- Get the app ID for reference
DO $$
DECLARE
    helloworld_app_id UUID;
BEGIN
    SELECT id INTO helloworld_app_id FROM apps WHERE name = 'helloworld';

    -- Clear existing helloworld data
    DELETE FROM edges WHERE app_id = helloworld_app_id;
    DELETE FROM nodes WHERE app_id = helloworld_app_id;

    -- Insert helloworld nodes
    INSERT INTO nodes (id, app_id, type, name, description, properties) VALUES
        -- Specifications
        ('helloworld-app-spec', helloworld_app_id, 'spec', 'HelloWorld App Specification', 'Container application specification for HelloWorld service', '{"image": "helloworld:latest", "port": "3000", "env": "production", "resources": {"cpu": "200m", "memory": "256Mi"}}'),
        ('helloworld-db-spec', helloworld_app_id, 'spec', 'HelloWorld Database Specification', 'PostgreSQL database configuration for HelloWorld app', '{"version": "14", "storage": "10Gi", "backup": "daily", "replicas": "1"}'),
        ('helloworld-cache-spec', helloworld_app_id, 'spec', 'HelloWorld Cache Specification', 'Redis cache configuration for HelloWorld app', '{"version": "7", "memory": "256Mi", "persistence": "false", "replicas": "1"}'),

        -- Workflows
        ('deploy-helloworld-container', helloworld_app_id, 'workflow', 'Deploy HelloWorld Container', 'Kubernetes deployment workflow for HelloWorld application container', '{"namespace": "helloworld", "replicas": "3", "strategy": "RollingUpdate", "healthCheck": "/health"}'),
        ('deploy-helloworld-database', helloworld_app_id, 'workflow', 'Deploy HelloWorld Database', 'Helm chart deployment for PostgreSQL database', '{"chart": "postgresql", "version": "12.1.9", "values": "custom-values.yaml"}'),
        ('deploy-helloworld-cache', helloworld_app_id, 'workflow', 'Deploy HelloWorld Cache', 'Redis deployment workflow with persistence disabled', '{"chart": "redis", "version": "17.3.7", "mode": "standalone"}'),
        ('setup-helloworld-schema', helloworld_app_id, 'workflow', 'Setup HelloWorld Schema', 'Initialize database schema and seed data for HelloWorld', '{"migrations": "flyway", "seedData": "true", "tables": ["users", "messages", "sessions"]}'),

        -- Resources
        ('helloworld-container', helloworld_app_id, 'resource', 'HelloWorld Container', 'Running HelloWorld application container in Kubernetes', '{"image": "helloworld:v1.2.3", "replicas": "3", "status": "Running", "endpoints": ["http://helloworld.example.com"]}'),
        ('helloworld-database', helloworld_app_id, 'resource', 'HelloWorld PostgreSQL Database', 'PostgreSQL database instance for HelloWorld application', '{"endpoint": "helloworld-db.default.svc.cluster.local:5432", "database": "helloworld", "status": "Ready"}'),
        ('helloworld-cache', helloworld_app_id, 'resource', 'HelloWorld Redis Cache', 'Redis cache instance for HelloWorld application', '{"endpoint": "helloworld-redis.default.svc.cluster.local:6379", "status": "Ready", "maxMemory": "256mb"}'),
        ('helloworld-db-schema', helloworld_app_id, 'resource', 'HelloWorld Database Schema', 'Database schema and initial data for HelloWorld', '{"version": "1.2.0", "tables": "3", "indexes": "5", "seedRecords": "100"}');

    -- Insert helloworld edges showing dependencies and relationships
    INSERT INTO edges (id, app_id, from_node_id, to_node_id, type, description, properties) VALUES
        -- Workflow dependencies on specs
        ('hw-e1', helloworld_app_id, 'deploy-helloworld-container', 'helloworld-app-spec', 'depends-on', 'Container deployment requires application specification', '{"validation": "dockerfile", "security": "scan"}'),
        ('hw-e2', helloworld_app_id, 'deploy-helloworld-database', 'helloworld-db-spec', 'depends-on', 'Database deployment requires database specification', '{"validation": "config", "backup": "enabled"}'),
        ('hw-e3', helloworld_app_id, 'deploy-helloworld-cache', 'helloworld-cache-spec', 'depends-on', 'Cache deployment requires cache specification', '{"validation": "config", "monitoring": "enabled"}'),

        -- Resource provisioning by workflows
        ('hw-e4', helloworld_app_id, 'deploy-helloworld-container', 'helloworld-container', 'provisions', 'Workflow provisions the HelloWorld container', '{"method": "kubernetes", "rollout": "blue-green"}'),
        ('hw-e5', helloworld_app_id, 'deploy-helloworld-database', 'helloworld-database', 'provisions', 'Workflow provisions the PostgreSQL database', '{"method": "helm", "persistence": "pvc"}'),
        ('hw-e6', helloworld_app_id, 'deploy-helloworld-cache', 'helloworld-cache', 'provisions', 'Workflow provisions the Redis cache', '{"method": "helm", "persistence": "none"}'),
        ('hw-e7', helloworld_app_id, 'setup-helloworld-schema', 'helloworld-db-schema', 'creates', 'Schema setup creates database structure and data', '{"tool": "flyway", "baseline": "1.0.0"}'),

        -- Infrastructure dependencies
        ('hw-e8', helloworld_app_id, 'deploy-helloworld-container', 'helloworld-database', 'depends-on', 'HelloWorld container needs database to be ready', '{"connection": "tcp", "timeout": "30s"}'),
        ('hw-e9', helloworld_app_id, 'deploy-helloworld-container', 'helloworld-cache', 'depends-on', 'HelloWorld container needs cache to be ready', '{"connection": "tcp", "timeout": "10s"}'),
        ('hw-e10', helloworld_app_id, 'setup-helloworld-schema', 'helloworld-database', 'depends-on', 'Schema setup needs database instance to be ready', '{"connection": "admin", "privileges": "create"}'),

        -- Service bindings
        ('hw-e11', helloworld_app_id, 'helloworld-container', 'helloworld-database', 'binds-to', 'HelloWorld app connects to PostgreSQL database', '{"connection_string": "postgresql://user:pass@helloworld-db:5432/helloworld", "pool_size": "10"}'),
        ('hw-e12', helloworld_app_id, 'helloworld-container', 'helloworld-cache', 'binds-to', 'HelloWorld app connects to Redis cache', '{"connection_string": "redis://helloworld-redis:6379/0", "timeout": "5s"}'),
        ('hw-e13', helloworld_app_id, 'helloworld-db-schema', 'helloworld-database', 'binds-to', 'Schema is deployed to PostgreSQL database', '{"schema": "public", "owner": "helloworld_user", "encoding": "UTF8"}');

END $$;

-- Insert sample graph run history for helloworld
INSERT INTO graph_runs (app_id, version, status, started_at, completed_at, execution_plan, metadata)
SELECT
    apps.id,
    1,
    'completed',
    NOW() - INTERVAL '2 hours',
    NOW() - INTERVAL '1 hour 45 minutes',
    '{"total_nodes": 11, "execution_order": ["helloworld-db-spec", "helloworld-cache-spec", "helloworld-app-spec", "deploy-helloworld-database", "helloworld-database", "deploy-helloworld-cache", "helloworld-cache", "setup-helloworld-schema", "helloworld-db-schema", "deploy-helloworld-container", "helloworld-container"], "duration_seconds": 900, "parallel_executions": ["deploy-helloworld-database", "deploy-helloworld-cache"]}',
    '{"triggered_by": "developer", "environment": "staging", "git_commit": "helloworld-v1.2.3", "deployment_type": "initial"}'
FROM apps WHERE name = 'helloworld';

-- Insert another graph run showing a recent update
INSERT INTO graph_runs (app_id, version, status, started_at, completed_at, execution_plan, metadata)
SELECT
    apps.id,
    2,
    'completed',
    NOW() - INTERVAL '30 minutes',
    NOW() - INTERVAL '15 minutes',
    '{"total_nodes": 11, "execution_order": ["helloworld-container"], "duration_seconds": 300, "updated_nodes": ["helloworld-container"], "reason": "rolling_update"}',
    '{"triggered_by": "ci_pipeline", "environment": "staging", "git_commit": "helloworld-v1.2.4", "deployment_type": "rolling_update", "previous_version": "v1.2.3"}'
FROM apps WHERE name = 'helloworld';