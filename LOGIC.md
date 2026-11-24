# Tamagochi business logic

## Structure

Routes are declared only in `backend/internal/app/http/routes/router.go`.
Handlers decode HTTP requests and call domain services. Services implement rules;
repositories perform PostgreSQL queries and transactions.

All public endpoints use the `/api/v1` prefix.

## Authentication and users

- `POST /register` creates a `USER`, its empty profile, and an access/refresh token pair.
- `POST /login` verifies the password and issues a new token pair.
- `POST /refresh` consumes a single-use refresh-session ID and issues a rotated pair.
- `PUT /users/me` updates only the authenticated user. A password is re-hashed only when supplied. The user's role is never accepted from the request.
- `DELETE /users/me` deletes only the authenticated user. PostgreSQL prevents deletion when referenced orders exist.
- `GET/POST/PUT/DELETE /profiles/me` operates only on the current profile. Profiles are created at registration; POST and PUT save the same profile, while DELETE clears bio and address instead of breaking the required user-profile relation.

## Catalog and favorites

- Products and product images are public to read; writes require `ADMIN`.
- Product price is a decimal string, never a floating-point value.
- Images can only be changed for an existing product and require a non-empty storage key and a non-negative sort order.
- `POST /likes/{productID}` is idempotent: adding an already liked existing product returns the existing like.
- `DELETE /likes/{productID}` is idempotent: removing a missing like succeeds.
- A favorite can be created only for an existing product and belongs only to the authenticated user.

## Cart

`PENDING` is the active cart status. Migration `000004` ensures one pending cart per user.

- `GET /cart` reads the active cart and its items.
- `POST /cart/items` creates a cart when necessary. Re-adding a product increases the existing item quantity.
- `PUT /cart/items/{id}` replaces a quantity; `DELETE /cart/items/{id}` removes a line.
- Item prices are copied from the product by PostgreSQL. `total_price` is recalculated in the same transaction as each item change.
- Only `PENDING` orders accept item changes.

## Checkout, payment and delivery

- `POST /cart/checkout` requires a non-empty cart and a delivery address from 5 to 500 characters. The address is copied to the order and cannot be affected by later profile changes.
- `POST /orders/{orderID}/payment` is a deterministic university-project payment simulator. `{ "approved": true }` marks the order paid; false marks it failed. A failed payment can be retried.
- Users can see only their own orders. A request for another user's order returns `404`.
- Users may cancel `PENDING`, `AWAITING_PAYMENT`, or `PAYMENT_FAILED` orders. They cannot mark an order paid or delivered.
- Administrators move paid orders through delivery statuses.

```text
PENDING --checkout--> AWAITING_PAYMENT --payment approved--> PAID
                                      --payment rejected--> PAYMENT_FAILED
PAYMENT_FAILED --payment approved--> PAID
PAID --> DELIVERING --> DELIVERED
PENDING | AWAITING_PAYMENT | PAYMENT_FAILED --> CANCELLED
```

`DELIVERED` and `CANCELLED` are terminal statuses. Illegal transitions return HTTP `409 Conflict`.

## Database migrations

- `000004_one_pending_cart_per_user`: partial unique index for one active cart.
- `000005_checkout_and_payment`: adds `delivery_address` and `AWAITING_PAYMENT`.

Apply them with `make migrate-up` before starting the API against an existing database.
