package ai

import (
	"context"

	"github.com/cortezaproject/corteza/server/compose/service"
	"github.com/cortezaproject/corteza/server/compose/types"
)

// GenerateToolSchema fetches active object metadata and builds LLM JSON schemas
func GenerateToolSchema(ctx context.Context, nsID uint64, modSvc service.ModuleService) (string, error) {
	mods, _, err := modSvc.Find(ctx, types.ModuleFilter{NamespaceID: nsID})
	if err != nil {
		return "", err
	}
	// TODO: Transform mods into OpenAI JSON schema tool definitions
	_ = mods
	return `{"type": "function", "name": "create_record"}`, nil
}
