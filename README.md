# Atabeh

**Atabeh** is a cross-platform VPN/Proxy configuration aggregator, normalizer, and evaluator designed to help you find the best connection path before you connect. It's not just another VPN client—it's the gateway to informed connectivity choices.

## What is Atabeh?

In regions where internet access faces restrictions or where VPN configurations are scattered across countless channels and sources, Atabeh serves as your intelligent assistant. It collects proxy and VPN configurations from diverse sources, tests them for quality and availability, and presents you with the best options ranked by performance.

Think of Atabeh as a quality filter for the chaotic world of VPN configs. Instead of manually trying dozens of configurations hoping one works, Atabeh does the heavy lifting for you.

## Core Philosophy

The name "Atabeh" suggests a threshold or doorway—a fitting metaphor for what this tool represents. Before you step through to connect to the internet, Atabeh shows you which door leads to the best path. It doesn't force you through any particular door; it simply illuminates your options with real data.

## Key Features

**Multi-Source Support**: Atabeh can pull configurations from GitHub raw links, subscription URLs, local files, JSON/YAML documents, and in the future, Telegram channels and bots. It doesn't care where your configs come from—it just wants to help you find the good ones.

**Format Agnostic**: Whether you're working with VMess, VLess, Trojan, Reality, Clash, or standard subscription links, Atabeh speaks your language. It normalizes everything into a unified data model so you can compare apples to apples.

**Intelligent Testing**: Each configuration goes through connection tests measuring ping, latency, and availability. Atabeh doesn't just tell you a config exists—it tells you if it actually works and how well.

**Smart Ranking**: Configurations are scored based on real performance metrics. The fastest, most stable connections rise to the top, saving you time and frustration.

**Modular Architecture**: The core is built in Go for speed and reliability, while clients for Android, iOS, Windows, Linux, and macOS communicate through a clean API. This separation means you can build new clients or integrate Atabeh into existing tools without touching the core logic.

## Architecture Overview

Atabeh follows a clear separation of concerns:

**Core Engine (Go)**
- Configuration fetching and parsing
- Network testing and latency measurement
- Data normalization and storage
- API/gRPC service for client communication

**Platform Clients**
- Independent applications for each platform
- Communicate with core via standardized API
- Handle UI/UX specific to each platform
- Can be developed independently of core

This modular design means the heavy computation happens once in the core, and lightweight clients simply display results. It also means Atabeh can work in resource-constrained environments or on devices with limited processing power.

## Use Cases

**Daily Users**: You just want your VPN to work reliably. Atabeh finds the best configs so you don't waste time testing them manually.

**Network Administrators**: Managing multiple proxy configurations across a team becomes simpler when you can see which ones actually perform well.

**Researchers**: Studying VPN availability and performance in different regions benefits from Atabeh's systematic testing approach.

**Developers**: Building apps that need VPN capabilities can integrate Atabeh's core to provide users with working configurations out of the box.

## Current Status

This repository contains the core engine. Platform-specific clients are under development and will be released in separate repositories. The core is functional and can be used via command-line interface or integrated into your own applications through the API.

## Getting Started

Right now, Atabeh core is in active development. If you're interested in contributing, testing, or just watching the project evolve, check out the [Contributing Guidelines](CONTRIBUTING.md) and [Code of Conduct](CODE_OF_CONDUCT.md).

For build instructions and technical details, see the documentation in the `docs/` directory.

## Why Go?

The core is written in Go for several practical reasons: exceptional concurrency support (essential for testing hundreds of configs simultaneously), cross-platform compilation without dependencies, strong networking libraries, and a balance of performance and development speed. Go lets Atabeh be fast where it matters and maintainable where it counts.

## Security Considerations

Atabeh tests network configurations, which means it makes outbound connections. All testing happens in isolated contexts, and no configuration data is sent to external services without explicit consent. See [SECURITY.md](SECURITY.md) for details on reporting vulnerabilities and our security practices.

## License

Atabeh is released under the MIT License. See [LICENSE](LICENSE) for details.

## Community

We're building Atabeh in the open because we believe tools for internet freedom should be transparent and accessible. Whether you want to contribute code, documentation, translations, or just share feedback, you're welcome here. Every threshold needs people to walk through it.

## Acknowledgments

This project stands on the shoulders of the many developers who've created the VPN protocols, parsers, and tools that make modern internet privacy possible. Atabeh aims to make their work more accessible to everyone who needs it.