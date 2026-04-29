package state

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// StateManager provides utilities for common state operations
type StateManager struct {
	ctx context.Context
}

// NewStateManager creates a new StateManager instance
func NewStateManager(ctx context.Context) *StateManager {
	return &StateManager{ctx: ctx}
}

// RemoveFromState removes a resource from Terraform state
func (sm *StateManager) RemoveFromState(state *resource.ReadResponse) {
	state.State.RemoveResource(sm.ctx)
}
