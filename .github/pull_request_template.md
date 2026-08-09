<!--
zbridge-asm-lab PR template. The reasoning behind each section is in docs/team/charter.md §5.
Delete sections that genuinely do not apply — do not delete them to avoid answering.
-->

## What and why
<!-- The change, and the problem it solves. "Why" is not optional on this project. -->

Ticket:

## Which unknown does this retire?

- [ ] **U1** — Go assembly correctness / Go ABI on a big-endian 64-bit target
- [ ] **U2** — WTO parameter list, `SVC 35` acceptance
- [ ] **U3** — `GOOS=zos`, `Malloc31`, AMODE 31↔64, extended WPL, USS
- [ ] **None** — and here is why it is still worth doing:

## Claims check

Every factual claim about IBM, z/OS, MVS or z/Architecture behaviour in this PR is:

- [ ] **cited** — primary source with form number, or a Go source path
- [ ] **evidenced** — traceable to a file in `docs/evidence/`
- [ ] **registered** — an explicitly open assumption inside a hypothesis
- [ ] no such claims in this PR

There is no fourth category. Hedged guessing is the specific thing this box exists to stop.
When you do not know, the correct output is a research brief or a hypothesis.

## Verification — run, not reasoned about

<!-- Paste actual output. "Should work" fails review. -->

- [ ] `go vet ./...`
- [ ] `go test ./...`
- [ ] `GOOS=linux GOARCH=s390x go build ./...`  *(n/a for `add/` — it is not supposed to)*
- [ ] not applicable, because:

```
<paste output here>
```

## If this touches `docs/evidence/`

- [ ] Provenance header complete: machine, guest OS, architecture, emulator version, `speaks_to`
- [ ] **No emulated result is presented as a z/OS result** (ADR 0001 §7)
- [ ] **No performance number taken under an emulator** — QEMU and Hercules implement `TR` as
      a software loop, so timing it measures the emulator

## If this touches assembly

- [ ] Hardware `R15` and the pseudo-register `name+0(SP)` are not conflated
- [ ] Argument offsets re-derived — strings are 2 words, slices are 3 words
- [ ] Any `EXRL` target block is `NOSPLIT|NOFRAME`
- [ ] No `UNDEF` stub was replaced with a no-op or a zero-returning body

## Documentation

- [ ] `docs/implementations/YYYY-MM-DD-<slug>.md` written, or genuinely not warranted because:
- [ ] `memory/MEMORY.md` still accurate after this change

## Anything deliberately left undone

<!--
Scope is delivered, not narrowed. If part of the ticket was left out, say so here and why.
Scaling work down is the lead's decision, not the contributor's.
-->

---

### Reviewer routing (see charter §6)

| This PR touches | Approval required |
|---|---|
| `docs/decisions/**` | **W5 / lead — mandatory** |
| `ebcdic/tables.go`, any `LICENSES.md` | **W5** — BSD-3-Clause attribution must survive |
| `console/wpl.go` `LayoutVerified` | **W5** — it is a claim about reality, not a flag |
| an `UNDEF`-bearing stub | **W5** |
| `docs/evidence/**` | W4 |
| `zbridge/**` public API | W2 + 1 reviewer |
| any `.s` file | 1 reviewer, after `go-asm-reviewer` |
