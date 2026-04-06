package claudemd

const SectionHeader = "# Secrets Management"

const SecretsMdTemplate = `# Secrets Management

This project uses [tene](https://github.com/tomo-kay/tene) for secret management.

## Usage
- Get a secret: ` + "`tene get <KEY>`" + `
- List secrets: ` + "`tene list`" + `
- Run with secrets injected: ` + "`tene run -- <command>`" + `
- Set a secret: ` + "`tene set <KEY> <VALUE>`" + `

## Rules
- Never hardcode secret values in source code
- Access secrets via environment variables
- Do not create .env files -- use ` + "`tene run`" + ` instead
- Use ` + "`tene list`" + ` to see available secrets
`
