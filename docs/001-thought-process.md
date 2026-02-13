# Address Validation Service - Thought Process

## Problem

Your task is to design and implement a backend API that validates and standardizes property addresses. The API should expose a single endpoint (ie POST /validate-address) that accepts a property address in free-form text and returns a structured, validated version of the address, including street, number, city, state, zip code. The service can restrict its output for US addresses only. Your solution should handle edge cases (ie partial addresses, typos, etc) gracefully and indicate whether the address is valid, corrected or unverifiable.

## Assumptions

1. Use an external service to validate the address and guarantee of existence
2. Well formatted inexistent address should be returned as `invalid`
3. Typos can be returned as `corrected` if the external service provides the real information, or `unverifiable` if not
4. Valid address should return as `valid`

## Architectural Approaches

### 1) API with synchronous processing

More: @docs/003-sync-approach-chart.md

1. Client -> Call `POST /v1/validate-address`
2. API -> Perform fail-fast validations
  1. Maybe return `invalid`
3. API -> Parse the free-form address (manual or using library (i.e. `libpostal`))
	1. Maybe return `invalid` or `unverifiable`
4. API -> Call external service (i.e. `Google Address Validation API`)
	1. Maybe return `invalid` or `unverifiable`
5. API -> Transform to API contract
	1. Return `valid` or `corrected`

#### Pros

- Simplicity, fast to implement and test

#### Cons

- External dependencies can fail (i.e. downtimes, rate-limiting)
- Multiple responsibilities on a single request

### 2) API with asynchronous processing (Chosen)

More: @docs/004-async-approach-chart.md

1. Client -> Call `POST /v1/validate-address`
2. API -> Perform fail-fast validations
	1. Maybe return `invalid`
3. API -> Normalize the free-form address
4. API -> Check if the normalized address exist on Cache (Redis)
	1. If exist, return the result
5. API -> Parse the free-form address (manual or using library (i.e. `libpostal`))
	1. Maybe return `invalid` or `unverifiable`
6. API -> Publish the normalized address on a Queue/Event
7. API -> Wait for the result by checking if the normalized address exist on Cache (Redis)
	1. If exist, return the result
	2. If not, wait until reach the timeout
	3. If timeout, return time
8. Worker -> Receive the normalized address from Queue/Event
9. Worker -> Call external service (i.e. `Google Address Validation API`)
10. Worker -> Write the address inside Cache (Redis) with the status (`invalid`, `unverifiable`, `valid`, `corrected`)

#### Pros

- Separation of responsibilities
- API can handle more requests by just publishing in the Queue/Event
- Scale service resources independently

#### Cons

- Increased client-side complexity: Polling or long-waiting (with timeouts) can frustrate users expecting instant results, potentially needing webhooks for true async callbacks.
- Eventual consistency risks: Cache misses during waits lead to partial failures; stale data if TTLs aren't tuned properly.
- Higher operational overhead: Managing queues (e.g., RabbitMQ dead-lettering), workers, and cache invalidation adds monitoring and debugging needs.

## Tech Stack

- API: Go 1.25 (GoFr - https://gofr.dev)
- Worker: Go 1.25 (GoFr - https://gofr.dev)
- Cache: Redis
- Event: Google PubSub
- External Service: Google Address Validation API
- Address Parsing: libpostal (gopostal)

## Comparison

| Aspect      | Synchronous                                  | Asynchronous                                     |
| ----------- | -------------------------------------------- | ------------------------------------------------ |
| Latency     | Low for successes; blocks on failures nylas​ | Variable; fast cache hits, else delayed webpeak​ |
| Scalability | Limited by external deps nylas​              | High; independent scaling nylas​                 |
| Complexity  | Low dev/ops linkedin​                        | Higher (queues, polling) linkedin​               |
| Reliability | Single failure point developers.google​      | Resilient with retries dzone​                    |

## Improvements and next steps

- Add system and business metrics (cpu, memory, cache hit/miss)
- Add fallback for external services with circuit breaker

## External tools and libraries

- [Google Address Validation API](https://developers.google.com/maps/documentation/address-validation/overview)
- [gopostal](https://github.com/openvenues/gopostal)
