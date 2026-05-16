# Security Policy

Tene takes security seriously. It encrypts secrets locally with
XChaCha20-Poly1305 and makes zero network calls by default.

## Reporting a Vulnerability

**Do not open a public GitHub issue for security vulnerabilities.**

Instead, email **security@tene.sh** (or contact `@agent-kay-it` privately on
GitHub if the alias is not yet live).

Please include:

- A clear description of the vulnerability
- Steps to reproduce (PoC code welcome)
- Affected version(s)
- Your proposed remediation, if any

We aim to acknowledge within 72 hours and fix critical issues within 14 days.

## Supported Versions

| Version | Supported |
| ------- | :-------: |
| 1.x     | ✅        |
| < 1.0   | ❌        |

## Security Model

- **Encryption**: XChaCha20-Poly1305 (256-bit keys, 192-bit nonces,
  secret name bound as AAD)
- **Key Derivation**: Argon2id (64 MiB memory, 3 iterations)
- **Key Storage**: OS native keychain (macOS Keychain, Linux libsecret,
  Windows Credential Vault)
- **Recovery**: 12-word BIP-39 mnemonic
- **Network**: Zero network calls from the CLI by default
- **Audit**: All encryption primitives live in `pkg/crypto/` and can be
  inspected under the MIT license.

## AI-Safe Design Properties

The CLI actively defends against AI agent secret exfiltration:

- `tene run -- <cmd>` is the primary workflow — secrets are injected as
  environment variables and never printed to stdout.
- `tene list` returns secret **names** only, never values.
- `tene get <KEY>` refuses non-TTY stdout by default to prevent accidental
  leakage into AI agent context windows, log aggregators, and shell history.
  Opt in with `--unsafe-stdout` or `TENE_ALLOW_STDOUT_SECRETS=1` when you
  truly need piped output.
- `.tene/` contents are never served by the landing site (robots.txt
  disallow).

## Security Disclosures Log

_None yet — first valid disclosure will be recorded here (and a CVE
requested if applicable)._

## Bug Bounty

There is no formal program at this time. Security researchers are credited
here when they submit valid reports.
