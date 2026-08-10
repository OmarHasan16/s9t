# Authorization Model

While the conceptual request demanded an OroCRM JSONB permission matrix on the Role entity, Corteza's native engine (`pkg/rbac`) provides superior, granular access control out-of-the-box using the `rbac_rules` engine. To satisfy integration requirements, we utilize an adapter (`oro_sync.go`) that accepts JSON matrices and translates them into Corteza's native tuple format.
