# Velarvo Desktop

Velarvo Desktop is a local-first developer workspace for SSH, SFTP, databases
and APIs. It is built with [Wails 3](https://v3.wails.io) (Go) and Svelte 5.

All data lives in an encrypted SQLite vault on the user's machine, unlocked
with a master password. The app does not talk to any backend.

## Requirements

| Tool          | Version                                                    |
| ------------- | ---------------------------------------------------------- |
| Go            | 1.27 (toolchain pinned in `go.mod`)                        |
| Bun           | see `.bun-version`                                         |
| Wails CLI     | same version as `github.com/wailsapp/wails/v3` in `go.mod` |
| golangci-lint | v2.15+                                                     |

Install the Wails CLI at the version the app uses:

```sh
go install github.com/wailsapp/wails/v3/cmd/wails3@$(go list -m -f '{{.Version}}' github.com/wailsapp/wails/v3)
wails3 doctor
```

On Linux, Wails 3.1 targets GTK4 and WebKitGTK 6.0
(`libgtk-4-dev libwebkitgtk-6.0-dev pkg-config build-essential` on Debian/Ubuntu).

## Getting started

```sh
wails3 task install   # root tooling + frontend dependencies
wails3 dev            # run with hot reload
```

All project tasks live in `Taskfile.yml` and run through the Task runner
bundled with the Wails CLI. `wails3 task --list` shows them all.

| Task                   | What it does                                              |
| ---------------------- | --------------------------------------------------------- |
| `wails3 dev`           | Run the app with hot reload                               |
| `wails3 build`         | Build for the host platform into `bin/`                   |
| `wails3 package`       | Build a distributable (`.app`, NSIS installer, …)         |
| `wails3 task bindings` | Regenerate `frontend/bindings` after changing Go services |
| `wails3 task check`    | Everything CI runs: format, lint, types, build            |
| `wails3 task fmt`      | Format Go and frontend code                               |
| `wails3 task audit`    | Dependency vulnerability audit                            |

The husky git hooks call `wails3 task precommit` and `wails3 task prepush`.

## Layout

```
main.go                 entry point: logging, database, services, window, menu
internal/db             SQLite connection and migrations (migrations/*.sql)
internal/vault          data key, encryption, setup/unlock/lock
internal/projects       projects and their icons
internal/workspaces     workspaces of a project
internal/sshconn        SSH connections and keys, terminals, SFTP
internal/files          local filesystem and file previews
internal/settings       user preferences
internal/apperr         errors with a code the frontend translates
internal/logger         slog setup and the frontend log bridge
frontend/               Svelte 5 + Vite + Tailwind 4
frontend/bindings/      generated TypeScript bindings (do not edit)
build/                  platform build assets and Taskfiles
```

Records are encrypted with the vault key (`vault.Vault.Seal` / `Open`) and
stored with their SQL inline in each package.

### Services and bindings

Every `Service` (plus `sshconn.Terminal`, `sshconn.SFTP` and
`logger.Frontend`) is registered as a Wails service in `main.go`. Its exported
methods become typed TypeScript functions in `frontend/bindings`, imported
through the `@bindings/*` alias. Keep helpers that must not be callable from
the frontend unexported, or on a type that is not registered (like
`vault.Vault`).

Methods return `(T, error)`. A Go error rejects the promise; errors created
with `apperr.New` carry a code in `error.cause.code`, which `describeError`
translates through `errors.<CODE>` in the i18n files:

```ts
import { Service as Vault } from "@bindings/vault";
import { describeError } from "@/lib/errors";

try {
  const state = await Vault.Unlock(password);
} catch (error) {
  message = describeError(error);
}
```

A method that takes a `context.Context` is cancelled when the frontend calls
`.cancel()` on its promise (SFTP transfers use this).

Events pushed from Go (terminal output, transfer progress) are registered with
`application.RegisterEvent` in `internal/sshconn`, which gives them typed
listeners:

```ts
import { Events } from "@wailsio/runtime";
import { SftpTransferProgress } from "@events";

Events.On(SftpTransferProgress, ({ data }) => console.log(data.done));
```

CI fails if `frontend/bindings` is out of date, so run `wails3 task bindings`
and commit the result after changing a service.

### Frontend conventions

The frontend is Svelte 5 in runes mode throughout:

- Components use `$props`, `$state`, `$derived`, callback props (`onClose`,
  `onRenameNode`) instead of `createEventDispatcher`, snippets instead of
  slots, and `{@attach}` instead of `use:` actions.
- Shared state lives in classes with `$state` fields in `src/lib/*.svelte.ts`
  (for example `vault`, `projectsState`, `editor`). Data that is only ever
  replaced, never mutated, uses `$state.raw`. There are no `svelte/store`
  stores left.
- `t()` from `@/lib/i18n.svelte` is reactive, so it can be called directly in
  templates.

### Upgrading Wails

The Go module, the `@wailsio/runtime` npm package and the `wails3` CLI must be
on the same version. Upgrade them together:

```sh
go get github.com/wailsapp/wails/v3@vX.Y.Z && go mod tidy
bun add --cwd frontend @wailsio/runtime@X.Y.Z
go install github.com/wailsapp/wails/v3/cmd/wails3@vX.Y.Z
wails3 task bindings
```

## Releasing

Work lands on `dev`, which runs CI and security checks only. Releases are cut
from `main` with [release-please](https://github.com/googleapis/release-please):

1. Every push to `main` opens or updates a release PR. It bumps the version from
   the Conventional Commits since the last release (`feat` → minor, `fix` →
   patch, breaking changes → minor while below 1.0) and updates `CHANGELOG.md`
   and the version in `build/` (the lines marked `x-release-please-version`).
2. Merging the release PR tags `vX.Y.Z` and creates a draft GitHub release.
3. The release workflow runs CI, builds the packages and publishes the release
   with these assets and a `SHA256SUMS.txt`:

   | Platform | Assets                                          |
   | -------- | ----------------------------------------------- |
   | macOS    | universal `.dmg` and zipped `.app`              |
   | Windows  | NSIS `-installer.exe` and portable `.zip` (x64) |
   | Linux    | AppImage, `.deb` and `.rpm` (x86_64)            |

Run the **Package** workflow manually to build the same packages as workflow
artifacts without releasing. `scripts/set-version.sh <version>` stamps a version
locally.

### Repository setup

- Settings → Actions → General: allow GitHub Actions to create and approve pull
  requests.
- Optional `RELEASE_PLEASE_TOKEN` secret (fine-grained PAT or GitHub App token
  with contents and pull-request write access). Without it the release PR is
  opened by `GITHUB_TOKEN`, and GitHub does not run CI on it.

### Code signing

Unsigned builds work, but macOS Gatekeeper and Windows SmartScreen warn about
them. Signing turns on when these secrets are set:

| Secret                                                                         | Purpose                                                     |
| ------------------------------------------------------------------------------ | ----------------------------------------------------------- |
| `MACOS_CERTIFICATE`                                                            | Base64 of the "Developer ID Application" certificate `.p12` |
| `MACOS_CERTIFICATE_PASSWORD`                                                   | Password of that `.p12`                                     |
| `APPLE_API_KEY`                                                                | Contents of the App Store Connect API key `.p8`             |
| `APPLE_API_KEY_ID`, `APPLE_API_ISSUER`                                         | Key ID and issuer ID of that key, for notarization          |
| `AZURE_TENANT_ID`, `AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET`                    | Service principal for Azure Artifact Signing                |
| `AZURE_SIGNING_ENDPOINT`, `AZURE_SIGNING_ACCOUNT`, `AZURE_CERTIFICATE_PROFILE` | Artifact Signing account and profile                        |

`wails3 update build-assets` regenerates the files in `build/` and drops the
`x-release-please-version` markers; restore them afterwards.
