package auth

import (
	"context"
	"sync"
	"vatusa-cobalt/dbconn"
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
	mu sync.RWMutex
	// tokenActorMap is nil until the first Load completes. A plain nil map
	// is safe to read — it behaves like empty — so a request arriving before
	// that first Load just misses the cache instead of dereferencing a nil
	// pointer.
	tokenActorMap map[string]*TokenActor
}

func (c *TokenActorCache) Get(token string) (*TokenActor, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	tokenActor, ok := c.tokenActorMap[token]
	return tokenActor, ok
}

func (c *TokenActorCache) Load(ctx context.Context) error {
	tokens, err := dbconn.Queries().GetActiveActorTokens(ctx)
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
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tokenActorMap = newMap
	return nil
}
