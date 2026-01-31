package auth

import (
	"context"
	"log"
	"sync"
	"time"
	"vatusa-cobalt/db"
)

type TokenGrant = string

const (
	GrantAllPermissions        TokenGrant = "*"
	GrantViewPersonalDetails   TokenGrant = "view_personal_details"
	GrantViewEmailAddress      TokenGrant = "view_email_address"
	GrantViewTrainingRecords   TokenGrant = "view_training_records"
	GrantManageTrainingRecords TokenGrant = "manage_training_records"
	GrantManageRoster          TokenGrant = "manage_roster"
	GrantGenerateUserAuthToken TokenGrant = "generate_user_auth_token"
	GrantLegacySyncRoles       TokenGrant = "legacy_sync_roles"
)

type ActorType = string

const (
	ActorTypeInternalSystem ActorType = "system"
	ActorTypeFacilitySystem ActorType = "facility"
	ActorTypeExternalSystem ActorType = "external"
)

const GrantGlobalScopeFacility string = "*"

const ActorTokenHeader string = "X-Auth-Token"

const ContextTokenActor string = "TokenActor"

type TokenActor struct {
	Comment           string
	ActorName         string
	ActorType         ActorType
	RateLimitOverride int
	RateLimitBypass   bool
	Grants            []ActorGrant
}

type ActorGrant struct {
	Grant         TokenGrant
	ScopeFacility string
}

func (a *TokenActor) HasGrant(grant TokenGrant, scopeFacility string) bool {
	if scopeFacility == "" {
		return false
	}
	for _, g := range a.Grants {
		if g.Grant == GrantAllPermissions {
			return true
		}
		if g.Grant == grant {
			if g.ScopeFacility == scopeFacility ||
				g.ScopeFacility == GrantGlobalScopeFacility {
				return true
			}
		}
	}
	return false
}

func (a *TokenActor) HasGlobalGrant(grant TokenGrant) bool {
	return a.HasGrant(grant, GrantGlobalScopeFacility)
}

type TokenActorCache struct {
	sync.RWMutex
	tokenActorMap *map[string]*TokenActor
	lastRefresh   time.Time
}

func (c *TokenActorCache) Get(token string) (*TokenActor, bool) {
	c.RLock()
	tokenActor, ok := (*c.tokenActorMap)[token]
	c.RUnlock()
	return tokenActor, ok
}

func (c *TokenActorCache) Load(ctx context.Context, queries *db.Queries) error {
	tokens, err := queries.GetActiveActorTokens(ctx)
	if err != nil {
		return err
	}
	newMap := map[string]*TokenActor{}
	for _, token := range tokens {
		tokenActor := TokenActor{
			Comment:           token.Comment.String,
			ActorName:         token.Name,
			ActorType:         token.ActorType,
			RateLimitOverride: int(token.RateLimitOverride),
			RateLimitBypass:   token.RateLimitBypass,
			Grants:            make([]ActorGrant, 0),
		}
		grants, err := queries.GetACLGrantsForActor(ctx, token.ID)
		if err != nil {
			return err
		}
		for _, grant := range grants {
			ag := ActorGrant{
				Grant:         grant.Acl,
				ScopeFacility: grant.ScopeFacility,
			}
			tokenActor.Grants = append(tokenActor.Grants, ag)
		}
		newMap[token.Token] = &tokenActor
	}
	c.Lock()
	c.tokenActorMap = &newMap
	c.lastRefresh = time.Now()
	c.Unlock()
	log.Printf("Successfully loaded %d actor tokens.\n", len(*c.tokenActorMap))
	return nil
}
