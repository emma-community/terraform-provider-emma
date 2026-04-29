package state

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/emma-community/terraform-provider-emma/internal/emma/common/async"
	"github.com/emma-community/terraform-provider-emma/internal/emma/common/logging"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// ResourceStateChecker checks the current state of a resource
type ResourceStateChecker func(ctx context.Context) (string, error)

// StateTransitionConfig configures state transition behavior
type StateTransitionConfig struct {
	ResourceType       string
	ResourceID         string
	StatusChecker      ResourceStateChecker
	TargetStates       []string
	TransitionalStates []string
	FailureStates      []string
	Timeout            time.Duration
	PollInterval       time.Duration
}

// StateTransitionManager handles resource state transitions
// Each manager instance is independent and does not share state with other instances
type StateTransitionManager struct {
	config StateTransitionConfig
}

// NewStateTransitionManager creates a new manager
// Each call creates an independent manager instance for parallel operation support
func NewStateTransitionManager(config StateTransitionConfig) *StateTransitionManager {
	return &StateTransitionManager{config: config}
}

// WaitForStableState waits for resource to reach a stable state
// This method supports context cancellation and logs state transitions
func (m *StateTransitionManager) WaitForStableState(ctx context.Context) error {
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Log the start of state waiting
	tflog.Debug(ctx, "Waiting for resource to reach stable state", map[string]interface{}{
		"resource_type":  m.config.ResourceType,
		"resource_id":    m.config.ResourceID,
		"target_states":  m.config.TargetStates,
		"timeout":        m.config.Timeout.String(),
		"poll_interval":  m.config.PollInterval.String(),
	})

	startTime := time.Now()

	// Check current state before polling (idempotent check)
	currentState, err := m.config.StatusChecker(ctx)
	if err != nil {
		tflog.Error(ctx, "Failed to check current state", map[string]interface{}{
			"resource_type": m.config.ResourceType,
			"resource_id":   m.config.ResourceID,
			"error":         err.Error(),
		})
		return fmt.Errorf("failed to check current state: %w", err)
	}

	// Log current state
	tflog.Debug(ctx, "Current resource state", map[string]interface{}{
		"resource_type": m.config.ResourceType,
		"resource_id":   m.config.ResourceID,
		"current_state": currentState,
	})

	// Idempotent check: if already in target state, return immediately
	if m.IsStableState(currentState) {
		tflog.Info(ctx, "Resource already in stable state, skipping wait", map[string]interface{}{
			"resource_type": m.config.ResourceType,
			"resource_id":   m.config.ResourceID,
			"current_state": currentState,
		})
		return nil
	}

	// Check if in failure state
	if m.IsFailureState(currentState) {
		tflog.Error(ctx, "Resource is in failure state", map[string]interface{}{
			"resource_type": m.config.ResourceType,
			"resource_id":   m.config.ResourceID,
			"current_state": currentState,
		})
		return fmt.Errorf("resource %s %s is in failure state: %s", m.config.ResourceType, m.config.ResourceID, currentState)
	}

	// Log that we're starting to poll
	tflog.Info(ctx, "Resource in transitional state, starting to poll", map[string]interface{}{
		"resource_type": m.config.ResourceType,
		"resource_id":   m.config.ResourceID,
		"current_state": currentState,
		"target_states": m.config.TargetStates,
	})

	// Create a wrapper status checker that logs state transitions
	lastState := currentState
	wrappedStatusChecker := func(ctx context.Context) (string, error) {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		state, err := m.config.StatusChecker(ctx)
		if err != nil {
			return "", err
		}

		// Log state transition if state changed
		if state != lastState {
			logging.LogStateTransition(ctx, m.config.ResourceType, m.config.ResourceID, lastState, state)
			lastState = state
		}

		return state, nil
	}

	pollerConfig := async.PollerConfig{
		Timeout:       m.config.Timeout,
		PollInterval:  m.config.PollInterval,
		StatusChecker: wrappedStatusChecker,
		TargetStates:  m.config.TargetStates,
		FailureStates: m.config.FailureStates,
	}

	poller := async.NewPoller(pollerConfig)
	err = poller.Poll(ctx)

	duration := time.Since(startTime)

	if err != nil {
		tflog.Error(ctx, "Failed to reach stable state", map[string]interface{}{
			"resource_type": m.config.ResourceType,
			"resource_id":   m.config.ResourceID,
			"duration":      duration.String(),
			"error":         err.Error(),
		})
		return err
	}

	// Log successful completion
	finalState, _ := m.config.StatusChecker(ctx)
	tflog.Info(ctx, "Resource reached stable state", map[string]interface{}{
		"resource_type": m.config.ResourceType,
		"resource_id":   m.config.ResourceID,
		"final_state":   finalState,
		"duration":      duration.String(),
	})

	return nil
}

// IsTransitionalState checks if current state is transitional (case-insensitive)
func (m *StateTransitionManager) IsTransitionalState(state string) bool {
	for _, ts := range m.config.TransitionalStates {
		if strings.EqualFold(state, ts) {
			return true
		}
	}
	return false
}

// IsStableState checks if current state is stable (case-insensitive)
func (m *StateTransitionManager) IsStableState(state string) bool {
	for _, ss := range m.config.TargetStates {
		if strings.EqualFold(state, ss) {
			return true
		}
	}
	return false
}

// IsFailureState checks if current state is a failure state (case-insensitive)
func (m *StateTransitionManager) IsFailureState(state string) bool {
	for _, fs := range m.config.FailureStates {
		if strings.EqualFold(state, fs) {
			return true
		}
	}
	return false
}

// CheckCurrentState checks the current state of the resource
// This method supports context cancellation and returns the current state
func (m *StateTransitionManager) CheckCurrentState(ctx context.Context) (string, error) {
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	tflog.Debug(ctx, "Checking current resource state", map[string]interface{}{
		"resource_type": m.config.ResourceType,
		"resource_id":   m.config.ResourceID,
	})

	state, err := m.config.StatusChecker(ctx)
	if err != nil {
		tflog.Error(ctx, "Failed to check resource state", map[string]interface{}{
			"resource_type": m.config.ResourceType,
			"resource_id":   m.config.ResourceID,
			"error":         err.Error(),
		})
		return "", err
	}

	tflog.Debug(ctx, "Current resource state retrieved", map[string]interface{}{
		"resource_type": m.config.ResourceType,
		"resource_id":   m.config.ResourceID,
		"current_state": state,
	})

	return state, nil
}

// IsInTargetState checks if the resource is already in one of the target states
// This is useful for idempotent operations that should skip if already in desired state
func (m *StateTransitionManager) IsInTargetState(ctx context.Context) (bool, string, error) {
	currentState, err := m.CheckCurrentState(ctx)
	if err != nil {
		return false, "", err
	}

	isTarget := m.IsStableState(currentState)
	
	if isTarget {
		tflog.Debug(ctx, "Resource is already in target state", map[string]interface{}{
			"resource_type": m.config.ResourceType,
			"resource_id":   m.config.ResourceID,
			"current_state": currentState,
			"target_states": m.config.TargetStates,
		})
	}

	return isTarget, currentState, nil
}
