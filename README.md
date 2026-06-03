# Adora Test

Backend Engineer Assignment — Subscription Reconciler

## How to Run

### Start Everything

```bash
docker compose up --build
```

API will run on localhost:8080

## API Example

### Store webhook ingestion

```bash
curl --location --request GET 'localhost:8080/webhooks/store' \
--header 'Content-Type: application/json' \
--data '{
    "eventId": "evt_u42_1",
    "userId": "u_42",
    "type": "INITIAL_PURCHASE",
    "eventTimeMs": 1716700000000,
    "productId": "premium_monthly"
}'
```

### Marketplace bulk revoke

```bash
curl --location 'localhost:3000/webhooks/marketplace/revoke' \
--header 'Content-Type: application/json' \
--data '{
    "userIds": [
        "u_42",
        "u_91",
        "u_133"
    ]
}'
```

### Entitlement read endpoint

```bash
curl --location 'localhost:3000/users/u_42/entitlement'
```

## Design Decision

### High Level Architecture

```er
HTTP API / Worker
      ↓
Entitlement Usecase
      ↓
Repository (PostgreSQL)
```

#### Union architecture

To have separate layers so tests can be done isolated from each other.

- improves testability
- avoids HTTP coupling
- isolates business rules

### Polling Flow

```er
Worker
  ↓
Claim batch (SELECT ... FOR UPDATE SKIP LOCKED)
  ↓
Carrier API call
  ↓
Update entitlement
```

#### SKIP LOCKED for polling

Since the constraint is to have more than one instance of worker, so I use: `FOR UPDATE SKIP LOCKED` to safely process concurrent workers.

- no distributed locking needed
- safe horizontal replicas scaling

## Tradeoffs considered

- inner layer (e.g.: repository) depends on ent go ORM

## What you would change if you had another week

- add db migration file
- implement proper audit log with event sourcing
- add more infra layer for structured logging for observability
- polling with event driven using other message broker
