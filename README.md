# narrowcast

A lightweight **Personal News Radio** — fetch RSS feeds, filter by your interests, and read headlines aloud (TTS). Runs on **macOS** (built-in `say`) and **Raspberry Pi** (Piper / espeak-ng).

## Quick start

```bash
cp config.example.yaml config.yaml
make run
make test
```

```bash
go run . -once -dry-run
go run . -daemon
```

## Make targets

Run `make help` for the full list. Common targets:

| Target | Description |
|--------|-------------|
| `make build` | Build `bin/narrowcast` |
| `make run` | Build and run one cycle |
| `make test` | `go test -v ./...` |
| `make lint` | `golangci-lint run` |
| `make release` | Build `dist/*` binaries (for GitHub releases) |

Release binaries are published when you push a version tag (`v0.3.0`, etc.); see `.github/workflows/release.yml`.

## CLI flags

| Flag | Description |
|------|-------------|
| `-config` | Config path (default `config.yaml`) |
| `-once` / `-daemon` | Run mode |
| `-dry-run` | Log matches, no TTS |
| `-tts` | Override: `mock`, `say`, `espeak`, `piper`, `elevenlabs`, `kokoro`, `auto` |
| `-state-file` | Seen URLs (`none` disables) |
| `-version` | Print version |

## Config highlights

| Field | Description |
|-------|-------------|
| `match_style` | `substring` or `word` (avoids `go` in `cargo`) |
| `feeds` | URL strings or `{url, label}` |
| `opml` | Merge feeds from OPML file |
| `audio.engine` | `auto` tries `fallback` chain |
| `audio.fallback` | Platform default if empty; e.g. `[say, piper, espeak, mock]` |
| `audio.say` | macOS voice (`Samantha`) and speech rate |
| `audio.elevenlabs` | Cloud TTS (`voice_id`, `ELEVENLABS_API_KEY` env) |
| `audio.kokoro` | Local script path (stdin = text; for Pi) |

Terms can use regex literals: `/\braspberry\s+pi/i`

## ElevenLabs (cloud)

```bash
export ELEVENLABS_API_KEY=your_key
```

```yaml
audio:
  engine: elevenlabs
  elevenlabs:
    voice_id: YOUR_VOICE_ID
    model_id: eleven_turbo_v2_5
    play_binary: afplay   # mpv on Linux
```

## Kokoro (local script on Pi)

Point `audio.kokoro.script` at a shell script that reads text from **stdin** and plays audio. See `deploy/kokoro-speak.sh.example`.

```yaml
audio:
  engine: kokoro
  kokoro:
    script: /home/pi/bin/kokoro-speak.sh
```

## macOS

No extra TTS install needed — macOS includes `say`:

```bash
# List voices
say -v ?

# config.yaml
audio:
  engine: say          # or auto with fallback: [say, mock]
  say:
    voice: Samantha
    rate: 180

go run . -once -tts say
```

With `engine: auto`, defaults are `say → elevenlabs → …` on macOS and `kokoro → piper → espeak → mock` on Linux (unconfigured engines are skipped).

## Raspberry Pi install

```bash
# On your dev machine (or download narrowcast-linux-arm64 from GitHub Releases)
make release
scp dist/narrowcast-linux-arm64 pi@raspberrypi.local:narrowcast/bin/narrowcast

# On the Pi
chmod +x ~/narrowcast/bin/narrowcast
sudo cp deploy/narrowcast.service /etc/systemd/system/
# Edit User, paths, and ExecStart in the unit file
sudo systemctl enable --now narrowcast
```

## Homebrew

```bash
brew tap ezeoleaf/tap
brew install narrowcast
```

After a new release, update `sha256` in `homebrew-tap/Formula/narrowcast.rb` (build with `make release` and `shasum -a 256 dist/narrowcast-darwin-amd64`).

Install TTS deps on the Pi:

```bash
sudo apt install espeak-ng alsa-utils
# Optional: piper + voice model for audio.engine auto
```

## CI

GitHub Actions (`.github/workflows/build.yml`) runs build, `golangci-lint`, and tests on push/PR to `main`. Tag pushes (`v*`) trigger release binaries via `release.yml`.

## Roadmap

See [ROADMAP.md](ROADMAP.md).

## License

MIT (add LICENSE when publishing).
