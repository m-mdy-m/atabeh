# Atabeh

**Status**: Core engine in development. Functional but incomplete. Clients coming soon.

**Fuck censorship. Fuck unreliable VPNs. Fuck testing configs manually.**

## Why This Exists

Every single day, I had to:
- Test dozens of VPN configs manually
- Watch half of them fail immediately
- Ping servers one by one like an idiot
- Switch between configs constantly
- Start the whole process over when something breaks
- Waste hours just trying to access basic internet services

I was going insane. Actually insane. So I decided to build something to automate this nightmare.

That's Atabeh.

## What It Does

Atabeh collects VPN and proxy configurations from wherever you point it—GitHub links, subscription URLs, local files, Telegram channels, wherever. Then it:

1. Parses all the different formats (VMess, VLess, Trojan, Reality, Clash, whatever)
2. Tests every single config for actual connectivity and speed
3. Ranks them by performance
4. Shows you which ones actually fucking work

No more guessing. No more manual testing. No more wasted time.

## The Name

"Atabeh" means **Aftabeh** (if y know! y know)

## Current Status

This is the core engine. It's **under active development**.

## For Developers

If you want to contribute, read [CONTRIBUTING.md](CONTRIBUTING.md). If you want to understand how it works, check the `docs/` folder.

The core is written in Go because:
- It handles concurrency well (needed for testing hundreds of configs at once)
- It compiles to native binaries without runtime dependencies
- It's fast enough for what we need

That's it. No other bullshit justification needed.

## Architecture

**Core (this repo)**: Does all the heavy work—fetching, parsing, testing, ranking.

**Clients (separate repos)**: Lightweight apps that talk to the core via API.

This separation means the core can run on a server or locally, and clients just display results. Build whatever client you want.

## Security

This tool tests network connections. It makes outbound requests to VPN servers. It doesn't send your data anywhere else, but obviously be smart about what sources you trust.

See [SECURITY.md](SECURITY.md) for reporting vulnerabilities.

## License

MIT License. Do whatever you want with it. If this helps you access information your government doesn't want you to see, that's exactly why it exists.

## A Note

Tools like Atabeh exist because governments think they can control what people see and say. They can't. They never could. The internet interprets censorship as damage and routes around it.

This is just another route.

## Dedication

To everyone in Iran fighting for basic freedoms. To everyone anywhere dealing with censorship and restricted internet. To everyone who's tired of authoritarian bullshit.
