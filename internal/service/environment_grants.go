package service

import (
	"context"
	"slices"

	"github.com/Hikyo-Org/hikyo/internal/authz"
	"github.com/Hikyo-Org/hikyo/internal/domain"
)

// releaseEnvironmentGrants is part of authorized environment deletion, not an
// independent grant-revocation surface. An environment's scoped authority ends
// with the environment, including every origin that held that authority alive.
// Project, organization and sibling-environment grants remain untouched.
func releaseEnvironmentGrants(ctx context.Context, az *authz.TxAuthorizer, scope domain.Scope) error {
	lines, err := az.GrantLinesInProject(ctx, string(scope.Org), string(scope.Project))
	if err != nil {
		return err
	}
	var scoped []authz.GrantLine
	var principals []domain.PrincipalID
	for _, line := range lines {
		if line.Grant.Scope != scope {
			continue
		}
		scoped = append(scoped, line)
		principals = append(principals, line.Principal)
	}
	slices.Sort(principals)
	principals = slices.Compact(principals)
	for _, principal := range principals {
		if err := az.LockTargetPrincipal(ctx, principal); err != nil {
			return err
		}
	}
	for _, line := range scoped {
		// Re-read origins after acquiring the principal lock: another grant
		// writer may have attached an origin since the initial census.
		origins, err := az.GrantOriginsFor(ctx, line.ID)
		if err != nil {
			return err
		}
		for _, origin := range origins {
			if _, err := az.ReleaseGrantOrigin(ctx, line.ID, line.Principal, origin); err != nil {
				return err
			}
		}
		if _, err := az.DeleteGrantRow(ctx, line.ID, line.Principal); err != nil {
			return err
		}
	}
	for _, principal := range principals {
		if err := invalidateGrantChange(ctx, az, principal); err != nil {
			return err
		}
	}
	return nil
}
