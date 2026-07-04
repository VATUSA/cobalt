package auth

import "time"

const ContextUserCID = "UserCID"

const ContextActorId = "ActorId"

const JWTCookieName = "vatusa-cobalt-token"

// SessionDuration bounds both the JWT's exp claim and the session cookie's
// lifetime, so a stolen cookie can't outlive the token it carries.
const SessionDuration = time.Hour * 24 * 7
