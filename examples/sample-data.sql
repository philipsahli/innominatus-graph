-- Sample data for IDP Orchestrator demo
-- Run this after the main migrations to populate demo data

-- Insert demo application
INSERT INTO apps (name, description)
VALUES ('demo', 'Demo application showing IDP orchestration patterns')
ON CONFLICT (name) DO NOTHING;

-- Get the app ID for reference
DO $$
DECLARE
    demo_app_id UUID;
BEGIN
    SELECT id INTO demo_app_id FROM apps WHERE name = 'demo';

    -- Clear existing demo data
    DELETE FROM edges WHERE app_id = demo_app_id;
    DELETE FROM nodes WHERE app_id = demo_app_id;

    -- Insert demo nodes
    INSERT INTO nodes (id, app_id, type, name, description, properties) VALUES
        ('database-spec', demo_app_id, 'spec', 'Database Specification', 'PostgreSQL database configuration and requirements', '{"version": "13", "size": "medium", "replicas": "2"}'),
        ('api-spec', demo_app_id, 'spec', 'API Specification', 'REST API service specification', '{"framework": "gin", "version": "v1", "port": "8080"}'),
        ('frontend-spec', demo_app_id, 'spec', 'Frontend Specification', 'React frontend application specification', '{"framework": "react", "version": "18", "build": "webpack"}'),

        ('deploy-database', demo_app_id, 'workflow', 'Deploy Database', 'Terraform workflow to provision PostgreSQL database', '{"provider": "aws", "region": "us-east-1", "engine": "postgresql"}'),
        ('deploy-api', demo_app_id, 'workflow', 'Deploy API Service', 'Kubernetes deployment workflow for API service', '{"namespace": "production", "replicas": "3", "strategy": "rolling"}'),
        ('deploy-frontend', demo_app_id, 'workflow', 'Deploy Frontend', 'CDN deployment workflow for frontend assets', '{"provider": "cloudfront", "compression": "gzip", "caching": "aggressive"}'),
        ('run-migrations', demo_app_id, 'workflow', 'Database Migrations', 'Run database schema migrations', '{"tool": "flyway", "baseline": "1.0.0"}'),

        ('postgres-cluster', demo_app_id, 'resource', 'PostgreSQL Cluster', 'Production PostgreSQL database cluster', '{"endpoint": "prod-db.amazonaws.com", "port": "5432", "ssl": "required"}'),
        ('api-service', demo_app_id, 'resource', 'API Service', 'Production API service deployment', '{"url": "https://api.example.com", "health": "/health", "metrics": "/metrics"}'),
        ('frontend-app', demo_app_id, 'resource', 'Frontend Application', 'Production frontend application', '{"url": "https://example.com", "cdn": "cloudfront"}'),
        ('database-schema', demo_app_id, 'resource', 'Database Schema', 'Current database schema version', '{"version": "1.5.2", "tables": "15", "views": "8"}');

    -- Insert demo edges showing dependencies and relationships
    INSERT INTO edges (id, app_id, from_node_id, to_node_id, type, description, properties) VALUES
        -- Workflow dependencies on specs
        ('e1', demo_app_id, 'deploy-database', 'database-spec', 'depends-on', 'Database deployment needs specification', '{"validation": "required"}'),
        ('e2', demo_app_id, 'deploy-api', 'api-spec', 'depends-on', 'API deployment needs specification', '{"validation": "required"}'),
        ('e3', demo_app_id, 'deploy-frontend', 'frontend-spec', 'depends-on', 'Frontend deployment needs specification', '{"validation": "required"}'),

        -- Resource provisioning by workflows
        ('e4', demo_app_id, 'deploy-database', 'postgres-cluster', 'provisions', 'Workflow provisions the database cluster', '{"method": "terraform"}'),
        ('e5', demo_app_id, 'deploy-api', 'api-service', 'provisions', 'Workflow provisions the API service', '{"method": "kubernetes"}'),
        ('e6', demo_app_id, 'deploy-frontend', 'frontend-app', 'provisions', 'Workflow provisions the frontend app', '{"method": "s3-cloudfront"}'),
        ('e7', demo_app_id, 'run-migrations', 'database-schema', 'creates', 'Migrations create/update database schema', '{"tool": "flyway"}'),

        -- Service dependencies
        ('e8', demo_app_id, 'deploy-api', 'postgres-cluster', 'depends-on', 'API service needs database to be ready', '{"connection": "required"}'),
        ('e9', demo_app_id, 'deploy-frontend', 'api-service', 'depends-on', 'Frontend needs API service to be ready', '{"cors": "configured"}'),
        ('e10', demo_app_id, 'run-migrations', 'postgres-cluster', 'depends-on', 'Migrations need database cluster to be ready', '{"schema": "public"}'),

        -- Resource bindings
        ('e11', demo_app_id, 'api-service', 'postgres-cluster', 'binds-to', 'API service binds to database cluster', '{"connection_pool": "10", "ssl": "require"}'),
        ('e12', demo_app_id, 'frontend-app', 'api-service', 'binds-to', 'Frontend binds to API service endpoints', '{"base_url": "https://api.example.com/v1"}'),
        ('e13', demo_app_id, 'database-schema', 'postgres-cluster', 'binds-to', 'Schema is deployed to database cluster', '{"schema": "public", "owner": "app_user"'}');

END $$;

-- Insert sample graph run history
INSERT INTO graph_runs (app_id, version, status, started_at, completed_at, execution_plan, metadata)
SELECT
    apps.id,
    1,
    'completed',
    NOW() - INTERVAL '1 hour',
    NOW() - INTERVAL '30 minutes',
    '{"total_nodes": 11, "execution_order": ["database-spec", "api-spec", "frontend-spec", "deploy-database", "postgres-cluster", "run-migrations", "database-schema", "deploy-api", "api-service", "deploy-frontend", "frontend-app"], "duration_seconds": 1800}',
    '{"triggered_by": "admin", "environment": "production", "git_commit": "abc123def456"}'
FROM apps WHERE name = 'demo';

COMMIT;