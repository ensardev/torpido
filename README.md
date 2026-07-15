<div align="center">

# 🚢 torpido

**Battleship. In your terminal. Over SSH.**

```
ssh torpido.dev
```

No install. No signup. No mouse. Just you, a fleet, and questionable tactics.

[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Built with Bubble Tea](https://img.shields.io/badge/built%20with-Bubble%20Tea-ff69b4)](https://github.com/charmbracelet/bubbletea)
·
[**torpido.dev**](https://torpido.dev)

</div>

## Play

```
ssh torpido.dev
```

That's the whole thing — you land straight in the lobby. From there:

- **Warm up on a bot** — three difficulties waiting around the clock.
- **1v1 a friend** — create a room, text them the 4-letter code, settle it.
- **Quick match** — get paired with whoever's around.

Your SSH key is your identity, so your wins, losses and nickname follow you back every time — no account, no password.

## Controls

| | |
|---|---|
| **Placing your fleet** | arrows / `hjkl` to move · `r` to rotate · `enter` to drop |
| **Firing** | arrows / `hjkl` to aim · `enter` to fire |
| **Anywhere** | `q` steps back a screen · `ctrl+c` disconnects |

## Features

- ⚔️ **Real 1v1 over SSH** — invite codes, quick match, optional password rooms, rematches
- 🤖 **Three bot tiers** — Rookie (random), Admiral (hunt & target), Sea Wolf (probability-hunting, merciless)
- 🏆 **Persistent stats & leaderboard** — win/loss keyed to your SSH key, top 10 + your rank
- 💥 **Pure terminal juice** — rolling waves, exploding hits, a live battle log, pointed ships
- 🌍 **English & Turkish**

## How it works

torpido is a single Go binary that speaks SSH. [Wish](https://github.com/charmbracelet/wish) turns each incoming SSH session into a [Bubble Tea](https://github.com/charmbracelet/bubbletea) program, so the exact same code drives your local terminal and every remote player. No web server, no client to install — the SSH connection *is* the game.

## Self-host

```sh
go install github.com/ensardev/ssh-torpido@latest
ssh-torpido serve            # listens on :2222; set TORPIDO_ADDR=:22 to change
```

To run it as `ssh yourdomain` (no port), put torpido on port 22 and move admin SSH aside. The [`deploy/`](deploy/) folder has a ready `systemd` unit and a one-command deploy script.

## License

MIT © [Ensar Akkuzey](https://ensar.dev)
