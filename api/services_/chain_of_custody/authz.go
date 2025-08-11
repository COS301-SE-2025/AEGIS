package coc

import "context"

// SimpleAuthz is a permissive authorizer that allows all actions.
type SimpleAuthz struct{}

func (SimpleAuthz) CanLogCoC(ctx context.Context, actorID *string, caseID, evidenceID string, action Action) bool {
	return true
}
