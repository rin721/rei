# ADR 021: Business And Delivery Declarative Lifecycle

Date: 2026-04-21

## Status

Accepted

## Context

Infrastructure lifecycle composition was already declarative, but `business` and `delivery` slices still used hand-written step arrays and direct method lists.

That meant:

- mode coordinators still depended on slice-local sequencing details
- delivery bootstrap, start, reload, and shutdown were only partially modeled as named runtime behavior
- business and delivery lifecycle composition did not yet follow the same extension pattern as infrastructure

## Decision

### 1. Business and delivery slices now declare lifecycle capabilities

`internal/app` now models business and delivery runtime concerns as named slice capabilities instead of inline step arrays.

Current examples:

- business:
  - `business-modules`
- delivery:
  - `router`
  - `http-server`

### 2. Delivery runtime behavior is now profile-driven

Delivery lifecycle execution now resolves declaratively for:

- bootstrap
- start
- reload
- shutdown

`http-server` now participates through the delivery capability registry instead of being wired directly into handwritten lifecycle lists.

### 3. A shared lifecycle ordering kernel is now reused

Lifecycle ordering now shares one dependency-aware ordering helper across:

- infrastructure capabilities
- business slice capabilities
- delivery slice capabilities

This keeps deterministic ordering rules and reverse shutdown semantics aligned across app-owned lifecycle registries.

## Consequences

- Mode coordinators now know less about slice-local ordering and focus more on slice-to-slice coordination.
- Business and delivery slices now have a clearer extension point when new runtime concerns are introduced.
- The app shell is more internally consistent because infrastructure, business, and delivery now all compose lifecycle behavior declaratively.
- The next refinement is to move mode execution itself toward declarative mode plans so runtime coordinators select named slice profiles instead of calling slice methods directly.
