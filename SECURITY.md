# Security Policy

Atabeh deals with network configurations and connectivity testing, which means security considerations matter. This document explains how we handle security, what we expect from users, and how to report vulnerabilities.

## Understanding Atabeh's Security Context

Atabeh tests VPN and proxy configurations by making network connections. This is inherent to what the tool does—you can't test if a config works without trying to connect through it. However, this also means Atabeh handles potentially sensitive information and makes network requests that could be observed.

What Atabeh does:
- Tests connections to proxy and VPN servers
- Measures latency and availability
- Parses configuration data from various sources
- Provides an API for clients to request this information

What Atabeh doesn't do:
- Route your actual internet traffic (it's not a VPN client itself)
- Store your personal data or browsing activity
- Phone home or send telemetry without consent
- Share your configurations with third parties

## Supported Versions

We provide security updates for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 0.x.x   | :white_check_mark: |

As Atabeh is currently in pre-1.0 development, we focus security efforts on the current development version. Once we reach 1.0, we'll establish a more formal support policy for previous versions.

## Reporting a Vulnerability

If you discover a security vulnerability in Atabeh, please report it responsibly:

**DO NOT** create a public GitHub issue for security vulnerabilities.

Instead, use one of these methods:

1. **GitHub Security Advisories** (preferred): Go to the repository's Security tab and click "Report a vulnerability"

2. **Direct Email**: Contact the maintainers directly at the email address listed in the repository

3. **Encrypted Communication**: If the issue is particularly sensitive, request PGP keys for encrypted communication

### What to Include

When reporting a vulnerability, help us understand the issue by providing:

- A clear description of the vulnerability
- Steps to reproduce the issue
- Potential impact and attack scenarios
- Any suggested fixes or mitigations you've identified
- Your contact information for follow-up questions

### What to Expect

**Initial Response**: We'll acknowledge your report within 48 hours

**Investigation**: We'll investigate and validate the issue, which typically takes 1-7 days depending on complexity

**Updates**: We'll keep you informed about our progress and any challenges we encounter

**Resolution**: Once fixed, we'll work with you to coordinate disclosure timing

**Credit**: If you'd like, we'll credit you in the security advisory and release notes

## Security Considerations for Users

### Configuration Source Trust

Atabeh fetches configurations from sources you specify. We can't verify that these sources are trustworthy—that's on you. Only use configuration sources you trust. Remember that a malicious configuration could:

- Route your traffic through a compromised server
- Log your activity
- Inject malware or tracking into your connections
- Perform man-in-the-middle attacks

Atabeh tests if configs work, not whether they're safe to use for real traffic.

### Testing Network Impact

When Atabeh tests configurations, it makes actual network connections. This means:

- Your IP address is visible to the servers being tested
- If you're in a region with internet restrictions, testing certain servers might be observable
- Failed connection attempts might be logged by network infrastructure

Consider your threat model before running tests on configurations from unknown sources.

### Local API Security

If you expose Atabeh's API over a network (rather than just using it locally), understand the security implications:

- Anyone who can reach the API can trigger configuration tests
- Anyone who can reach the API can see your configuration list
- There's currently no authentication on the API (this is planned for future versions)

For now, only expose the API on localhost unless you're in a completely trusted network.

### Data Storage

Atabeh stores configuration data and test results locally. This data could be sensitive:

- Configuration details might include server addresses in restricted regions
- Test results show which configs you've tried
- Historical data shows patterns of what you're testing

Make sure your local storage is appropriately secured, especially on shared systems.

## Known Limitations

We want to be transparent about current limitations:

**Parser Security**: Parsers handle untrusted input from configuration sources. While we try to handle malformed data safely, parsers are complex and could have bugs that cause crashes or unexpected behavior.

**Network Isolation**: Tests run in the same process as the core engine. A malicious configuration that exploits a protocol implementation bug could potentially affect the entire application.

**Timing Attacks**: Timing measurements for testing could theoretically leak information about network topology or routing, though this is a minimal risk in practice.

**Dependency Security**: Like all software, Atabeh depends on external libraries. We monitor these for known vulnerabilities but can't guarantee they're perfect.

## Security Roadmap

We're working toward better security:

**Short term:**
- Improved input validation across all parsers
- Better sandboxing of test connections
- API authentication mechanisms
- Automated dependency vulnerability scanning

**Long term:**
- Process isolation for test execution
- Optional encrypted storage for sensitive configs
- Audit logging for security-relevant events
- Formal security audit of core components

## Responsible Disclosure Policy

We believe in coordinated disclosure. When a vulnerability is reported:

1. We'll work on a fix while keeping details private
2. We'll prepare a security advisory describing the issue and fix
3. We'll coordinate a disclosure timeline with the reporter (typically 90 days)
4. We'll release the fix and publish the advisory simultaneously
5. We'll credit the reporter (if they wish) and share any bug bounties or rewards that might apply

We expect security researchers to:

- Give us reasonable time to fix issues before public disclosure
- Not exploit vulnerabilities beyond what's needed to demonstrate the issue
- Not access or modify user data without permission
- Not perform testing that could harm the availability of services

## Questions?

If you have questions about Atabeh's security that aren't covered here, feel free to open a general discussion issue. For specific security concerns or potential vulnerabilities, use the reporting process described above.

Security is a collaborative effort. We appreciate everyone who helps make Atabeh safer for its users.