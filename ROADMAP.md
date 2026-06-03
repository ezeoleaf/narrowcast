# narrowcast Roadmap

Personal News Radio — lightweight Go backend for RSS → filter → TTS (Raspberry Pi friendly).

## Done

- [x] Core pipeline: `config` → `ingest` → `filter` → `audio`
- [x] YAML/JSON config (topics, keywords, feeds)
- [x] Concurrent RSS fetch (`gofeed`)
- [x] Mock / espeak / Piper TTS backends
- [x] **macOS `say` TTS** (built-in, no Homebrew deps)
- [x] Daemon mode with `schedule.interval`
- [x] Seen-URL state file (skip repeats)
- [x] URL deduplication (newest wins)
- [x] HTML cleanup + entity unescape
- [x] Sort by published date + `max_per_cycle`
- [x] Context-aware TTS + utterance length cap
- [x] State TTL + max entries + atomic save
- [x] Summary sanitizer (HN/Reddit boilerplate)
- [x] Per-feed HTTP timeout + one parser per goroutine
- [x] Exclude keywords + `match_mode` + `max_age`
- [x] `-dry-run`, `pause_between`, `state.persist`, `-version`
- [x] **Word-boundary / regex matching** (`match_style: word`, `/regex/`)
- [x] **OPML feed import** (`opml:` path merges with `feeds`)
- [x] **Per-feed labels** + **health stats in logs**
- [x] **TTS fallback chain** (`engine: auto`, `audio.fallback`)
- [x] **systemd unit** (`deploy/narrowcast.service`)
- [x] **Makefile** (build, test, vet, `build-linux-arm64`)
- [x] **GitHub Actions CI** (build, golangci-lint, test — termagotchi-style)
- [x] **Release workflow** + `make release` (darwin/linux/windows + linux-arm64)
- [x] **Homebrew formula** (`homebrew-tap/Formula/narrowcast.rb`)

## Near term

- [ ] Word-boundary matching for non-ASCII (Unicode letters)
- [ ] Feed OPML export / subscription sync
- [ ] Structured logging (`slog`) with log levels

## Mid term

- [ ] Optional LLM summary shortening (offline model for Pi)
- [ ] HTTP API / webhook trigger for cycles
- [ ] Bluetooth speaker routing helpers on Linux
- [ ] Multi-profile configs (`profiles/work.yaml`)

## Long term / ideas

- [ ] Podcast-style intro/outro clips between segments
- [ ] Time-of-day scheduling (morning briefing only)
- [ ] Push notifications when high-priority keyword matches
- [ ] Web UI for config editing on the LAN

## How to use this file

Move items from **Near term** → **Done** as they ship.
