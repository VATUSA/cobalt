package auth

import (
	"context"
	"sync"
	"time"
	"vatusa-cobalt/db"
)

type ActorType = string

const ActorTokenHeader string = "X-Auth-Token"

type TokenActor struct {
	ActorId           int
	Comment           string
	ActorName         string
	ActorType         ActorType
	RateLimitOverride int
	RateLimitBypass   bool
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
			ActorId:           int(token.ID),
			Comment:           token.Comment.String,
			ActorName:         token.Name,
			ActorType:         token.ActorType,
			RateLimitOverride: int(token.RateLimitOverride),
			RateLimitBypass:   token.RateLimitBypass,
		}
		newMap[token.Token] = &tokenActor
	}
	c.Lock()
	c.tokenActorMap = &newMap
	c.lastRefresh = time.Now()
	c.Unlock()
	return nil
}
