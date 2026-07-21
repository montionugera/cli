# promptctl

Git-backed authoring CLI for **Priompt** — versioned prompt serving over gRPC.
Prompts live as `*.prompt` files in a git repo; versioning, history, and lineage
are plain git, embedded via go-git (no external git binary). promptctl adds
validation, the semantic propagation diff, and environment promotion. A prompt's
URI is its path — `acme/support/agent.prompt` → `priompt://acme/support/agent`;
slots are read from the template's `{placeholders}`.

## Install

```sh
go build -o promptctl .        # Go 1.25+
```

## Usage

```sh
promptctl commit -m "msg"                     # validate every *.prompt, then commit changes
promptctl diff [ref]                           # semantic diff vs ref (default HEAD)
promptctl log acme/support/agent.prompt        # version lineage of one prompt
promptctl promote acme/x.prompt dev prod       # bring a prompt from branch dev onto prod
promptctl push [-server host:port]             # validate, git push, then publish to the server
promptctl pull                                 # sync with origin
promptctl publish -server host:port [paths…]   # publish to the server + notify subscribers
```

- **commit / push** validate first — a malformed prompt blocks the whole commit.
- **diff** compares each changed `*.prompt` between a ref and the working tree
  using the `PRIOMPT_EMBED_*` model (offline lexical by default).
- **promote** models dev → staging → prod as branches: it copies a validated
  prompt from one branch onto another and commits it.
- **push / pull** use go-git and do not read git's credential helper — set
  `PRIOMPT_GIT_TOKEN` for HTTPS remotes; SSH uses your ssh-agent.

`promptctl push -server …` git-pushes *and* publishes each prompt, so subscribers
are notified in the same step. Publishing is idempotent (`PublishPrompt` is a
no-op when the version hash is unchanged), so only changed prompts fire a
notification. Git stays the source of truth; the server is the live serving copy.

The gRPC stubs (`gen/`), `internal/semdiff`, and `internal/validate` are copies
from the core repo — regenerate there and sync when the proto or engines change.
