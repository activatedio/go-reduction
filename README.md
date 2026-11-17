# go-reduction

Go implementation of reducer based state tracking system. Useful for developing
an api for multi-step operations.

## Install

```
go get github.com/activatedio/go-reduction
```

## Use

A state type, actions with a reducer each, and optionally `Init` (the first
state) and `Refresh` (run on every read of a saved state):

```go
r := reduction.New(reduction.NewMemoryStore(20 * time.Minute))
r.Builder().State(reflect.TypeFor[Cart]()).
	Init(func(ctx context.Context) (*Cart, error) { return &Cart{}, nil }).
	Action(reflect.TypeFor[AddItem](), func(ctx context.Context, s *Cart, a *AddItem) (*Cart, error) {
		if a.Qty <= 0 {
			return nil, reduction.Invalid("qty", "must be more than 0")
		}
		return &Cart{ItemCount: s.ItemCount + a.Qty}, nil
	})

router := mux.NewRouter()
err := rmux.Mount(router, r, rmux.MountOptions{
	RootPath: "/api",
	KeyFunc:  func(r *http.Request) (string, error) { return signedInPerson(r) },
})
// GET /api/cart, POST /api/cart/add_item, GET /api/swagger.json
```

A state type with an `Export(ctx) (*View, error)` method is answered as the view.

### Whose state

`MountOptions.KeyFunc` names whose state a request reads and writes, such as
the signed-in person from a bearer. No cookie is needed; an error or an empty
key answers 401. Without a `KeyFunc`, the fx module's session middleware keeps
a random key in an HMAC-signed cookie, as before. Outside HTTP, set the key
with `reduction.WithKey(ctx, key)`.

### Errors

| Returned by a reducer, Init or Refresh | Answer |
|---|---|
| `reduction.Invalid(field, message)` (`*FieldError`) | 400 `{"error": message, "field": field}` |
| `reduction.NotAllowed(message)` (`*NotAllowedError`) | 403 `{"error": message}` |
| any error with `HTTPStatus() int` (`StatusError`) | that status |
| anything else | 500 |

Errors may be wrapped (`errors.As`). A body that is not valid JSON for the
action answers 400.

### Stores

`reduction.Store` keeps a state per key and state type:

- `reduction.NewMemoryStore(expiry)`: in this process, for tests and a single
  process.
- `github.com/activatedio/go-reduction/store/valkey`: Valkey or Redis, its own
  module, with a TTL and a key prefix.
- Your own, over your database: implement `Get` and `Set`. If it also
  implements `reduction.TxStore` (`InTx`), each read and action runs inside
  `InTx`, and reducers get its ctx, so their own writes and the saved state
  commit together.
