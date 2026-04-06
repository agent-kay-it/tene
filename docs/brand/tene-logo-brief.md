# Tene Logo Design Brief

## 1. Brand Identity

### What is Tene?
- **Category**: Developer CLI tool — secret management for AI agents
- **Tagline**: "Secret management that AI agents understand"
- **Personality**: Minimal, technical, trustworthy, developer-native
- **Analogy**: "The Git of secrets" — local-first, open source, CLI-driven

### Core Brand Values
| Value | Expression |
|-------|-----------|
| **Security** | Encrypted, locked, protected — but not intimidating |
| **Simplicity** | One command to start. No complexity. |
| **Developer-native** | Terminal aesthetic, monospace, CLI culture |
| **AI-aware** | Built for the AI agent era, forward-looking |
| **Open source** | Transparent, community-driven, trustworthy |

---

## 2. Color Palette

### Primary Colors
| Role | Color | Hex | Usage |
|------|-------|-----|-------|
| **Background** | Near Black | `#0a0a0a` | Dark mode default, app background |
| **Foreground** | Light Gray | `#ededed` | Primary text |
| **Accent (Primary)** | Terminal Green | `#00ff88` | Key highlights, CTAs, brand mark |
| **Accent (Dim)** | Deep Green | `#00cc6a` | Hover states, secondary accents |

### Supporting Colors
| Role | Color | Hex | Usage |
|------|-------|-----|-------|
| **Surface** | Dark Gray | `#141414` | Cards, elevated surfaces |
| **Border** | Subtle Gray | `#2a2a2a` | Dividers, outlines |
| **Muted** | Mid Gray | `#888888` | Secondary text, captions |

### Color Rationale
- **Terminal Green (#00ff88)** is the brand signature — it references the classic terminal cursor/prompt color
- The dark background + green accent conveys "developer tool" and "security" simultaneously
- High contrast green on dark = excellent readability and instant recognition

---

## 3. Typography

| Context | Font | Style |
|---------|------|-------|
| **Logo wordmark** | Geist Mono (or similar monospace) | Bold, tracking normal |
| **UI headings** | Geist Sans | Bold |
| **Code/terminal** | Geist Mono | Regular |

The logo should feel like it belongs in a terminal. Monospace is essential.

---

## 4. Logo Concept Directions

### Direction A: Terminal Prompt (Recommended)
```
$ tene
```
- The `$` prompt symbol in accent green, followed by `tene` in white/light
- Minimal, immediately recognizable to developers
- Matches the landing page nav exactly: `$ tene`
- Can be used as icon (`$` in green circle) + wordmark (`tene`)

**Icon variant**: A rounded square or circle with `$` or `>_` in terminal green on dark background

### Direction B: Lock + Terminal
- A minimal lock icon where the keyhole is shaped like a terminal cursor `█` or underscore `_`
- Combines "security" (lock) + "terminal" (cursor) in one symbol
- Accent green lock outline on dark background

### Direction C: Shield + Code
- A minimal shield outline with `{ }` or `< >` brackets inside
- Represents "code protection" / "secret guarding"
- Clean geometric shape, works at small sizes

### Direction D: Key Symbol
- A stylized key where the handle/bow is a terminal prompt `>_`
- Or the key teeth form a binary/code pattern
- Represents "keys" (API keys) + "developer tool"

---

## 5. Logo Specifications

### Must Have
- **Icon mark**: Works at 16x16 favicon size
- **Wordmark**: `tene` in monospace, clean
- **Combined mark**: Icon + wordmark horizontal layout
- **Dark background primary**: Logo designed for dark backgrounds first
- **Light background variant**: For docs, GitHub README, light contexts

### Size Variants
| Context | Size | Format |
|---------|------|--------|
| Favicon | 16x16, 32x32 | SVG, ICO |
| Nav bar | 24px height | SVG |
| Social / OG Image | 1200x630 | PNG |
| GitHub README | 200px width | SVG, PNG |
| npm badge | 20px height | SVG |

### Clear Space
- Minimum clear space around logo = height of the `t` character
- No elements should intrude into this space

---

## 6. Usage Examples

### Nav Bar (Current Landing Page)
```
[$ tene]  Features  How it works  Security  GitHub
```
The `$` is in #00ff88 (accent green), `tene` is in #ededed (foreground white).

### Terminal Context
```
$ npm install -g @tene/cli
$ tene init
```
Logo appears naturally in terminal usage.

### GitHub README Header
```
[Icon] tene
Secret management that AI agents understand.
```

---

## 7. What to Avoid

- Overly complex symbols — must be readable at 16px
- Gradients or 3D effects — keep it flat and minimal
- Generic lock/shield icons without terminal personality
- Bright/saturated colors beyond the accent green
- Rounded, playful, or "cute" aesthetics — this is a serious security tool
- Serif fonts or decorative typography

---

## 8. Inspiration References

| Reference | What to Take |
|-----------|-------------|
| **Git logo** | Simple, iconic, works at any size |
| **Vercel logo** | Geometric minimalism, dark-first |
| **1Password logo** | Trust/security with modern feel |
| **Warp terminal** | Developer aesthetic, dark + accent color |
| **Linear logo** | Clean mark that works as favicon |

---

## 9. Deliverables

1. **Primary logo** (icon + wordmark) on dark background
2. **Icon only** on dark background
3. **Light background variant** of both
4. **Favicon** (16x16, 32x32)
5. Color palette swatches

---

## Summary

> Tene's logo should feel like it was born in a terminal.
> Minimal, monospace, dark background, terminal green accent.
> A developer should look at it and immediately think: "CLI tool, security, I trust this."
