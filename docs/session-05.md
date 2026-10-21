# Session 5 — Who is calling

Session 4 left a hole on purpose: the JSON body chose `user_id`. This session closes it.

Authentication answers "which user is this?". A missing or bad token is 401. Authorization answers "may this user do this?". A student with a valid token who hits the kitchen route is 403. Those are different failures. Do not collapse both into 401.

By the end:

- `POST /v1/orders` and `GET /v1/orders/{id}` require `Authorization: Bearer`
- the public JSON has no `user_id`; the gateway copies the token subject onto the RPC
- a student token on accept is 403, and a kitchen token is not
- a student cannot read another student's order
- `/healthz` has no token check
- `JWT_SECRET` comes from the environment
- the signing method is HMAC-SHA256, and other methods are rejected

The order service does not parse the JWT. It trusts the user id the gateway put on the request. That is only true if nothing except the gateway can dial `order`. Do not publish `:9090` on a shared machine. Session 8 binds it for debugging and you will be explicit about that.

## What the token is

A JWT here is a signed blob: `sub` (user id), `role`, and `exp`. You sign it with the lab secret in `.env.example`, `dev-only-change-me`. That secret is not a production secret. Do not commit a real one, and do not log the token. Log `sub` and `role`.

The production shape is the same check with a different key: the university identity provider signs the token, and the gateway verifies it with their public key. This lab uses HMAC so you do not stand up an identity provider. An access token lives 15 minutes. There is no refresh token and no session table. A token cannot be revoked before `exp`; the short lifetime is the revocation story.

```bash
go get github.com/golang-jwt/jwt/v5
```

Claims:

```go
type Claims struct {
    Role string `json:"role"`
    jwt.RegisteredClaims
}
```

Mint a student token and a kitchen token from a tiny `cmd/token` program. `sub` is the seed user id.

```go
token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
    Role: role,
    RegisteredClaims: jwt.RegisteredClaims{
        Subject:   userID,
        ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
    },
})
signed, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
```

Ada is `11111111-1111-1111-1111-111111111111`, role `student`. The kitchen user is `22222222-2222-2222-2222-222222222222`, role `kitchen`.

## Middleware

Parse the bearer token in gateway middleware. Refuse any signing method other than HMAC-SHA256. `alg: none` and a token that swaps algorithms are the usual broken-JWT bugs. Returning the secret for every method would accept them.

```go
func Authenticate(secret []byte, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        raw, ok := bearer(r.Header.Get("Authorization"))
        if !ok {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        claims := &Claims{}
        _, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
            if t.Method != jwt.SigningMethodHS256 {
                return nil, fmt.Errorf("unexpected signing method")
            }
            return secret, nil
        })
        if err != nil || claims.Subject == "" {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        ctx := context.WithValue(r.Context(), userKey{}, User{ID: claims.Subject, Role: claims.Role})
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

`userKey` is an unexported empty struct type, not a string. A string key can collide with another package.

```go
mux.Handle("GET /healthz", http.HandlerFunc(live))
mux.Handle("POST /v1/orders", Authenticate(secret, http.HandlerFunc(h.createOrder)))
mux.Handle("GET /v1/orders/{id}", Authenticate(secret, http.HandlerFunc(h.getOrder)))
mux.Handle("POST /v1/kitchen/orders/{id}/accept",
    Authenticate(secret, RequireRole("kitchen", http.HandlerFunc(h.accept))))
```

`createOrder` no longer decodes `user_id`. It sets `CreateOrderRequest.UserId` from the user on the context. Delete the field from the public struct and from the OpenAPI request body.

`RequireRole` returns 401 when no user is on the context and 403 when the role does not match. Wire it inside `Authenticate`, as above, so the role check always sees a user.

Gateway `main` fails to start when `JWT_SECRET` is empty. `order` does not need the secret.

## Accept, for this session only

`accept` checks the role and returns 204. The status change and the outbox row are session 6, in one transaction. A 204 here means "this caller is allowed", not "the noodles are cooking". A courier token is also 403. Add that case if you mint a third token; the seed has no courier yet.

## Reading someone else's order

Middleware does not know which order belongs to whom. `GetOrder` in the order service compares the row's `user_id` with `caller_id`. A kitchen caller may read any order. Anyone else who does not own it gets `codes.PermissionDenied`.

```go
if order.UserID != req.CallerId && req.CallerRole != "kitchen" {
    return nil, status.Error(codes.PermissionDenied, "forbidden")
}
```

The gateway fills `caller_id` and `caller_role` from the token. Map `PermissionDenied` to 403 in `writeGRPCError`. A student who guesses a UUID must not see the order.

## OpenAPI

Add `bearerAuth` as an HTTP bearer scheme on the order routes and on accept. Do not put it on `/healthz`.

## Check

1. No `Authorization` header → 401.
2. Student token, create, JSON without `user_id` → 201, and the row's `user_id` is Ada's id.
3. Student token on accept → 403.
4. Kitchen token on accept → 204.
5. Change one character of the token → 401.
6. Student token, `GET` an order id that belongs to someone else → 403.
7. `GET /healthz` still returns 200 with no token.

## Homework

A courier role cannot accept. One sentence in `docs/architecture.md`: the order service does not verify the JWT.

## What this session is not

A password table, a refresh-token store, or Redis sessions. Roles do not come from JSON. The health check does not send a bearer token, because a process probe will not have one.
