# Binary Analysis — Platform Backend & Generated Service

> REA analysis of two Go ELF binaries: the running Lambra platform backend and a
> binary generated and deployed by the platform itself (`svc-test`).
> Both artifacts are shipped-artifact evidence; digests below pin the exact bytes.

---

## 1. Artifacts

| | Platform backend | Generated service (`svc-test`) |
|---|---|---|
| Local path | `backend/tmp/main` (Air build) | extracted from container `svc-test:/root/main` |
| SHA-256 | `7d054778f93ec2731c8929e20a5f27b285c58cdc57b3eceaf3035b8e8e7f9c28` | `9922e578438d26afd54580aa9a35c28b0cc741b37000027b38566e28ca2f5445` |
| Size | 17,303,130 bytes | 11,386,732 bytes |
| Format | ELF 64-bit, little-endian, ARM aarch64, statically linked | same |
| Symbols | not stripped, includes debug_info | same |
| Evidence | `ev_4293e88d40fca3e2d4edcea3f66e351408783580315e6eada6b98f01cac6bfbc` | `ev_446c9503c5bab46e84273bf9ad9d2bea8033227f4a3e6aefffdcd9c6e8151d02` |

## 2. Build Identity (REA inspect_go_binary)

| Property | Backend | svc-test |
|---|---|---|
| Go version | `go1.21.13` | `go1.21.13` |
| Main module | `github.com/yourusername/lambra` (`(devel)`) | `svc-test` (`(devel)`) |
| Command package | `github.com/yourusername/lambra/cmd/server` | `svc-test/cmd/server` |
| Build settings | `-buildmode=exe`, `-compiler=gc`, `CGO_ENABLED=0`, `GOOS=linux`, `GOARCH=arm64` | identical |

Findings:
- Both toolchains are identical (`go1.21.13`), confirming the platform builds generated
  services with the same pinned image (`golang:1.21-alpine`) it uses itself.
- `CGO_ENABLED=0` in both: fully static binaries — no libc attack surface from dynamic
  linking; matches the deployment template (`deployment_service.go` Dockerfile).
- The generated binary carries its **own module identity** (`svc-test`), i.e. the
  generator produces a self-contained Go module, not a fork of the platform.
- Neither binary embeds VCS stamping (`vcs.revision` absent from settings) — builds are
  not reproducibly tied to a git commit. Improvement opportunity for audit trail.
- Note: `(devel)` main-module version is normal for non-module-cache builds.

## 3. ELF Notes (REA inspect_elf_notes)

Both binaries expose exactly one note: `.note.go.buildid` (owner `Go`, descsz 83),
visible in both PT_NOTE and SHT_NOTE at file offset 3996.

| | Backend | svc-test |
|---|---|---|
| BuildID descriptor SHA-256 | `02bfcb8d2ea6fada0aed00f9e140e4827b10f3046504567cbfd2e869d810c09f` | `7ab297765681af9143f27114ef26b2b645bf2ac1c8387348c8a3f72105fd611f` |
| GNU build-id | absent | absent |

No GNU build-id and no `.note.ABI-tag` — expected for Go static executables; the Go
BuildID provides the reproducibility anchor instead.

## 4. Dependency Comparison

Backend: 22 modules; svc-test: 19 modules. Shared core: `gin-gonic/gin`,
`go-sql-driver/mysql`, `jmoiron/sqlx`, `google/uuid`.

**Backend-only (platform concerns):**
- `github.com/golang-jwt/jwt/v5 v5.2.1` — auth
- `github.com/gin-contrib/cors v1.5.0` — CORS
- `github.com/joho/godotenv v1.5.1` — env loading

**Version gaps in the generated service (transitive pins resolved by `go mod tidy`
inside the generated Dockerfile):**

| Module | Backend | svc-test | Note |
|---|---|---|---|
| `golang.org/x/crypto` | v0.17.0 | **v0.9.0** | security-relevant delta |
| `golang.org/x/net` | v0.18.0 | **v0.10.0** | security-relevant delta |
| `go-playground/validator/v10` | v10.15.5 | v10.14.0 | |
| `golang.org/x/text` | v0.14.0 | v0.9.0 | |
| `pelletier/go-toml/v2` | v2.1.0 | v2.0.8 | |
| `google.golang.org/protobuf` | v1.31.0 | v1.30.0 | |

### Finding F-1 (medium): generated services pin older dependencies than the platform

Root cause: the generator writes a minimal `go.mod` (direct deps only, no pinned
`go.sum` transitive floor) and relies on `RUN go mod tidy` at image-build time, so
transitive versions float to the minimum satisfying the resolved graph — which can lag
the platform's own audited pins by several minor versions (e.g. `x/crypto` v0.9.0 from
mid-2023 vs v0.17.0).

Impact: a known-vulnerable transitive dependency would persist in generated services
even after the platform itself is patched.

Recommendation: have `generateGoFiles` emit a `go.sum`-style floor (or `replace`
directives / `go mod tidy && go mod vendor` from the platform's own module cache at
generation time), so generated services inherit the platform's reviewed dependency set.
This is a strong "governed by construction" argument for the submission.

## 5. Supply-Chain Observations

- All dependencies observed are first-tier open-source Go modules; no unexpected or
  obfuscated modules appear in either build-info record.
- Build metadata is observed file content, not proof of compiler identity or dependency
  authenticity (REA limitation); for the hackathon scope the digests above anchor the
  analyzed bytes.
- Deep provider analysis (Ghidra/Hopper decompilation) was not available on this host;
  conclusions here derive from build metadata, ELF structure, and notes — appropriate
  for supply-chain/audit review of unstripped first-party binaries.
