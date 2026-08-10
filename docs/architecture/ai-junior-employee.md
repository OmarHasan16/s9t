# AI Junior Employee Flow

The AI Agent acts as a "Junior Employee" connected to the CRM.
- **Discovery:** It calls `ai.GenerateToolSchema` to read the active metadata in `compose`.
- **Constraint:** The LLM does NOT mutate data. It returns structured JSON mapped to `compose` module fields.
- **Approval:** Outputs are stored as standard `AiPendingTask` records in Corteza, requiring a human Admin approval workflow to finalize the transaction.
