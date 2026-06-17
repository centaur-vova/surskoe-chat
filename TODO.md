# TODO · Telegram Bot Platform

## 🔌 Distributed Plugin System

### Vision

Turn the bot from a monolithic app into a **pluggable platform** where plugins can be:

- Built as separate services
- Written in any language (Go, Python, PHP, Rust)
- Deployed independently
- Scaled horizontally

### Options under consideration

| Approach      | Pros                                                  | Cons                      |
| ------------- | ----------------------------------------------------- | ------------------------- |
| **NATS**      | Async, battle-tested in fast-atomic-flow, lightweight | Requires NATS cluster     |
| **gRPC**      | Fast, typed contracts, looks good in CV               | Sync, more boilerplate    |
| **HTTP REST** | Simple, debuggable                                    | Slower, harder to version |

### Current decision

**NATS** — because it's already familiar, async, and fits the event-driven model.

### Architecture (draft)

Telegram → Bot Core (NATS client)
↓
publish `message.received` to NATS
↓
Plugin (NATS subscriber) → process → publish `message.response`
↓
Bot Core → send to Telegram

### Plugin types

- **Filter** — intercepts messages (profanity, spam)
- **Command** — handles /commands (weather, bus, quote)
- **Background** — runs periodically (cleanup, reminders)

### Next steps

1. [ ] Add NATS client to bot core
2. [ ] Define message schemas (protobuf or JSON)
3. [ ] Move profanity filter to external NATS plugin
4. [ ] Add plugin discovery (via config or service registry)

---

## 📦 Planned Plugins

- [x] `profanity` — filter bad words (internal)
- [ ] `bus` — /bus schedule
- [ ] `welcome` — greet new users
- [ ] `ride` — ride sharing (TBD)
