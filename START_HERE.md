# START HERE

**This document assumes you know nothing about mainframes, and it means that
literally.** No macro, no acronym, no IBM term is used here without being defined the
first time it appears. If you're reading this to prepare to explain this project to
someone else — your mentor, a committee, anyone — read it top to bottom once, then use
it as your map back into the rest of the documentation whenever you need the full
depth on something.

Every technical claim in this document and everything it points to is either something
this project ran and captured, or something read directly from a primary source. Where
something is *not* settled, this document says so in plain words rather than smoothing
it over — because presenting an open question honestly is safer than getting caught
overstating something to someone who will ask a sharp follow-up.

---

## 1. The whole story, in five minutes

**What we set out to build.** A single function, in the Go programming language, that
puts a line of text on the screen a mainframe operator watches. That's the whole
functional goal. It sounds small. It is not small, for a specific reason explained in
§1.2.

**Why anyone would want this.** Companies that run IBM mainframes increasingly also
want to run modern software — written in modern languages like Go — near their
mainframe data, sometimes on the mainframe itself. When that modern software needs to
tell a human operator "something happened," it needs a way to reach the operator's
screen. Today, the *only* documented way to do that from outside IBM's own assembly
language is to go through a **C** programming language layer that IBM provides. Nobody
has published a way to do it in pure Go, without C. That gap — not "can it be done" but
"has anyone shown how to do it without C" — is what this project fills.

**Why it's hard: there is no test system.** Normally, when you build software, you test
it against the real thing. This project's "real thing" — an actual IBM mainframe
running its current operating system, z/OS, with a Go compiler that can target it —
does not exist in a form we can access yet. Two separate things are missing: (1) nobody
has given us permission to use a real z/OS system, and (2) even if they did, the
standard Go compiler cannot produce programs that run on z/OS at all — that requires a
special, IBM-maintained version of the compiler that we also don't have yet.

**The idea that made progress possible anyway.** Instead of treating "we have no
mainframe" as one big blocker, we broke the actual engineering problem into three
separate, independent questions, and asked, for each one: *is there something short of
a real mainframe that can answer this?* Two of the three turned out to have a "yes":

1. **"Does our low-level code actually run correctly on this kind of computer chip?"**
   — answered using a widely available, free **emulator** (a program that pretends to
   be a different kind of computer chip) called **QEMU**. This doesn't need any special
   permission or mainframe at all, because the *chip family* mainframes use also runs
   ordinary Linux, which is freely available.
2. **"Do we know the exact, byte-for-byte format the operator-message feature expects,
   and does it accept what we built?"** — answered by running a real, working,
   **legally free-to-run 1981 ancestor** of the modern mainframe operating system,
   under a different emulator called **Hercules**. This old system is old enough that
   IBM's original licensing restrictions on it have expired, so anyone can legally run
   it. It's not the current system, but it's close enough in family lineage to be
   genuinely informative — and we don't just assume that, we tested specific claims
   about it one at a time (§4 explains exactly how, and exactly where we're honest that
   it might not carry over perfectly).
3. **"Does a Go program actually run on the real current mainframe operating system,
   with all its specific memory and addressing rules?"** — this one, honestly, **still
   has no answer.** Nothing free or emulated can substitute for it. It needs real
   access we don't have yet. We say so plainly, everywhere in this project, rather than
   guessing.

**What actually happened.** Both of the answerable questions got answered, with real,
captured evidence — not just "it should work," but an actual old-mainframe emulator
accepting real bytes that our Go code built, and printing our message on a screen.
That's this project's central, provable result. The third question — actually running
on a real, current mainframe — remains open, and everyone involved has been told that
plainly, including in writing to the mentor.

**Where we are right now, as of this writing.** The technical result above happened
about two weeks before this document was written. Since then, this project scaled from
one person to a five-person team, and — the reason this document exists — the entire
system was documented from the ground up: what it is, how every piece works, how every
test proves what it claims to prove, and exactly how to run the whole thing live, on a
clean computer, as a demo. That demo is real and repeatable; instructions are in
[`RUN.md`](RUN.md).

### 1.1 The one sentence to lead with

*"We built the first published proof that a modern programming language can talk
directly to an IBM mainframe's operator console without going through C — and we
proved the hardest, riskiest part of it works, using free tools, before spending a
minute of actual mainframe access."*

### 1.2 Why "without C" is the actual point, not a footnote

This is worth being able to defend in one breath, because it's the single most likely
"so what" question. IBM's own supported way to reach the operator console from a
high-level language is to write a small C wrapper around a function IBM provides, and
call that C code from your other language. Every example anyone could find of another
language doing this — Python, Java — goes through exactly that same C layer
underneath. **This project's Go code never touches C, ever, anywhere.** It builds the
exact low-level bytes an IBM mainframe instruction expects, and hands them to that
instruction directly, in Go's own low-level "assembly" dialect (§2 defines this).
Skipping the C layer isn't a style preference — it removes an entire dependency (a C
compiler, and IBM's C support runtime) that a lot of modern, container-style,
minimal-footprint software specifically wants to avoid.

---

## 2. The words you need — read this before anything else, or keep it open in another tab

Grouped by topic, not alphabetically, because the topics build on each other.

### 2.1 The machine and its operating system

| Term | Plain-English meaning |
|---|---|
| **Mainframe** | A large, IBM-built computer designed to run thousands of business transactions per second for decades without going down, with several completely separate programs safely sharing one physical machine. Not "a bigger PC" — a different design philosophy entirely. |
| **IBM Z** | IBM's current mainframe product line — the actual hardware. |
| **z/OS** | IBM's current, flagship mainframe *operating system* (the software that runs on IBM Z hardware). This is the actual target of the whole project — but we don't have access to a real one yet. |
| **MVS 3.8j** | A specific, much older (1981) IBM mainframe operating system. It's a direct ancestor of z/OS — same family, decades earlier. Because it's so old, its original licensing restrictions have lapsed, and hobbyists have legally run it for over 20 years. This project uses it as a free stand-in for testing, **but it is explicitly not z/OS**, and every document in this project says so whenever it's relevant. |
| **s390x** | The name Go (and Linux) use for the *processor family* IBM mainframes use — a specific instruction set, independent of which operating system runs on top of it. |
| **z/Architecture** | IBM's name for the same processor family (the 64-bit modern version specifically). |
| **big-endian** | A way of ordering the bytes of a multi-byte number in memory, where the most significant byte comes first. Mainframes are big-endian; the laptop/desktop chips almost everyone uses day to day (Intel/AMD) are the opposite ("little-endian"). This distinction matters here because code that works on one can be silently wrong on the other if you're not careful — one of the things this project specifically tested for. |

### 2.2 Talking to the machine: jobs, and the people who run it

| Term | Plain-English meaning |
|---|---|
| **Operator** | The human who watches a running mainframe and is told when something needs attention. |
| **Operator console** | The screen the operator watches. Historically a physical terminal; in this project, entirely emulated and captured to a log file. |
| **Console message / console prefix** | Any line of text a program puts on the operator's screen. A leading `+` on a message (as opposed to no prefix, or `@`, or `*`) is the system's way of marking *"this came from an ordinary, non-privileged program,"* not an error indicator. |
| **JCL (Job Control Language)** | The scripting language used to hand a mainframe a unit of work (a "job") — think of it as the mainframe equivalent of a shell script that says "run this program, with these inputs." |
| **Job** | One unit of work submitted via JCL — e.g., "assemble this program, link it, then run it." |
| **JES2** | "Job Entry Subsystem 2" — the part of the operating system that receives submitted jobs, queues them, runs them in order, and returns their output. Functionally, a task queue plus a task runner. |

### 2.3 Building a program for the mainframe

| Term | Plain-English meaning |
|---|---|
| **Assembler** | A program that translates human-written, low-level source code (where each line is close to one actual CPU instruction) into the raw bytes a CPU executes. This project's mainframe-side code uses an assembler called **IFOX00** (also called "Assembler XF"), which is an older, more limited assembler than IBM's modern **HLASM** — it lacks some convenience features HLASM has, which occasionally shapes how the example programs in this project had to be written. |
| **Linker** | A program that takes assembled pieces and combines them into one runnable program. This project's linker is called **IEWL**. |
| **CSECT** | Short for "control section" — roughly, the mainframe assembly-language equivalent of naming a block of code as one unit, similar to how you'd name a function or a file in a modern language. |
| **Macro** | A named shortcut in assembly source that expands into a longer block of actual instructions/data when assembled — similar in spirit to a macro in C, or a code-generating template. IBM provides a macro called `WTO` that expands into everything needed to call the operator-message feature; this project's central technical discovery came from reading exactly what that macro expands into (see §2.4 and §4). |
| **`DC` / `DS`** | Assembly-language directives meaning "Define Constant" (reserve space *and* fill it with a specific value) and "Define Storage" (just reserve space). If you see `DC X'0016'` in this project's documentation, read it as "a constant, written in hexadecimal, with the value 0x0016." |
| **`PRINT GEN`** | An instruction to the assembler: "when you print your listing of what you assembled, include the full expansion of every macro, not just the line where I invoked it." This is *how* this project found out what a macro actually builds, byte for byte — see §4. |

### 2.4 The specific mainframe feature this project targets

| Term | Plain-English meaning |
|---|---|
| **`SVC`** ("Supervisor Call") | A single CPU instruction that means *"stop what you're doing, and let the operating system do something privileged on my behalf."* This is the mainframe's version of what a "syscall" is on Linux/Windows/Mac. Different numbers mean different privileged services. |
| **`SVC 35`** | The specific supervisor call number that means "Write To Operator" — put a message on the console. This is the exact instruction the whole project exists to reach. |
| **WTO ("Write To Operator")** | The name of the overall feature/service that `SVC 35` provides. |
| **WTOR ("Write To Operator with Reply")** | A harder variant: put a message on the console *and wait for the operator to type something back*. Not attempted yet in this project — it needs machinery (§2.5) that isn't available. |
| **WPL ("WTO Parameter List")** | The block of bytes you have to build and hand to `SVC 35` so it knows what message to display. Getting this exactly right, byte for byte, was the single hardest and most important technical challenge in the whole project — see §4 for the full story. |
| **MCS flags** | ("Multiple Console Support" flags) — a small set of on/off switches inside the WPL that control things like which console groups see the message. For the simplest form of message this project sends, every one of these flags is off (zero). |
| **Return code / RC** | A number a program or service reports back meaning "here's how it went" — `0000` conventionally means "no problem." One of this project's real findings: on the old system used for testing, this particular feature reports back *nothing at all* for the simple message form used here, which is different from reporting back zero — see §4.4. |

### 2.5 Registers, addressing, and the trickiest low-level detail in the project

| Term | Plain-English meaning |
|---|---|
| **Register** | A small, extremely fast storage slot built directly into the CPU — there are 16 of them, numbered R0 through R15, and different conventions assign different meanings to specific ones for specific purposes. |
| **R1** | By mainframe convention, when calling a service like `SVC 35`, this register must hold the memory address of the parameter list (the WPL) you built. |
| **R13, R14, R15** | Also carry specific meanings under mainframe calling convention (a save-area address, a return address, and an entry-point/return-code register, respectively) — **and two of the three collide with meanings Go's own compiler already assigns those same registers for completely different purposes.** This collision is, genuinely, the single most dangerous technical detail in this whole codebase — full explanation in `docs/architecture/zbridge-module.md` §5. |
| **AMODE ("Addressing Mode")** | Which of several addressing schemes a mainframe program is currently using — essentially, how large an address space it can currently reach, and how. |
| **"Below the bar"** | A specific address range restriction some mainframe services (including, likely, this one on the real current system) impose: the data you hand them has to live in memory below a specific address boundary. Handling this correctly needs help from IBM's own runtime library — the *only* piece of non-Go, IBM-provided software this project's design allows itself, and only for this one narrow purpose. |
| **Language Environment (LE)** | IBM's standard runtime library that most mainframe programs rely on for things like the below-the-bar memory allocation just described. This project uses exactly one function from it (for memory allocation) and nothing else — deliberately, to keep the "no C, minimal dependencies" story as clean as possible. |
| **USS (Unix System Services)** | A Unix-like environment that runs inside z/OS, providing things that feel more familiar to non-mainframe programmers (a shell, files, and so on). Relevant because it's likely where a Go program would actually execute, once it can be built for z/OS at all. |

### 2.6 Character encoding — the one place mainframes are flatly different

| Term | Plain-English meaning |
|---|---|
| **EBCDIC** | The mainframe's native way of representing text as bytes — completely different from the ASCII/UTF-8 essentially everything else uses today. The letter `A` is a different number in each system. |
| **IBM-1047** | The specific variant of EBCDIC that z/OS's Unix-like environment (USS) uses. Every message this project sends has to be translated into this encoding before it can be displayed correctly. |
| **Code page** | A general term for "a specific mapping between numbers and characters" — EBCDIC IBM-1047 is one code page; ASCII is another. |

### 2.7 The tools that stand in for a real mainframe

| Term | Plain-English meaning |
|---|---|
| **Emulator** | Software that pretends to be a different kind of computer (or a whole different computer system), closely enough that real software written for the real thing runs on it. |
| **Hercules** | A free, open-source emulator of full IBM mainframe hardware — CPU, storage, and the mainframe's specific way of handling input/output devices. Real, unmodified mainframe operating systems can run on it, believing they're on real hardware. |
| **TK5** | A ready-to-run package that bundles Hercules together with MVS 3.8j, already installed and configured, so you don't have to build a mainframe system from scratch — you just unpack it and start it. |
| **DASD ("Direct Access Storage Device")** | The mainframe's term for a hard disk. In this emulated setup, a DASD is just a (large) file on your computer's real disk that Hercules treats as if it were a mainframe disk drive. |
| **IPL ("Initial Program Load")** | The mainframe's word for "boot" — starting the operating system. |
| **QEMU** | A different, more general-purpose emulator. This project uses it in a lighter-weight mode that translates individual programs' CPU instructions on the fly, without emulating a whole machine or running any operating system underneath — much faster to use than Hercules, but it can't tell you anything about mainframe-*operating-system* behavior, only about whether raw CPU instructions are correct. |

### 2.8 Go-specific terms

| Term | Plain-English meaning |
|---|---|
| **Plan 9 assembly** | Go's own, portable dialect of assembly language. It's *not* the mainframe's native assembly syntax — it's Go's own notation, which Go's compiler then translates into the real machine instructions for whatever target you're building for. |
| **`GOOS` / `GOARCH`** | Environment settings that tell the Go compiler which operating system and which processor family to build for — e.g., `GOOS=linux GOARCH=s390x` means "build a program that runs on Linux, on mainframe-family hardware." |
| **cgo** | Go's built-in mechanism for calling C code from a Go program. This project deliberately never uses it — see §1.2. |
| **`go vet`** | A built-in Go tool that checks code for a range of correctness issues without running it — including, critically for this project, whether hand-written assembly code correctly declares how much stack space and how many argument bytes it uses (the "frame contract"). A mismatch here is one of the most common and dangerous mistakes in this kind of code, and `go vet` catches it mechanically, every time, for free. |

### 2.9 This project's own internal vocabulary

These words won't mean anything outside this project — they're vocabulary this team
invented to keep track of its own progress honestly.

| Term | Plain-English meaning |
|---|---|
| **U1 / U2 / U3** | The three separate unknowns described in §1 — "does the low-level code run right," "is the message format correct," and "does it all work on the real current system." Every piece of work in this project is labeled with which of these three it addresses. |
| **E-ladder** | The sequence of small, incremental proof steps (named E0, E1, E2, E3) this project climbed *without* a real mainframe, using the emulators from §2.7. Fully explained in `docs/architecture/evidence-ladder.md`. |
| **T-ladder** | The equivalent sequence of proof steps planned for *real* mainframe hardware, once access exists. Not started yet. |
| **Rung** | One step on either ladder. |
| **ADR ("Architecture Decision Record")** | A written record of one specific decision this project made, including what was considered and rejected, and — importantly — what would cause the team to reverse it. |
| **Hypothesis** | A specific, testable claim this project wrote down and committed to *before* gathering evidence for or against it, precisely so nobody can quietly move the goalposts after seeing the result. |
| **Evidence (file)** | A captured, dated record of something this project actually ran, with enough detail (what machine, what software versions, what exact output) that someone else could check the claim. |

---

## 3. The guided reading path — every document, in the order to read it

Each entry says what it covers, why it exists, and gives you one sentence you could
say out loud even if you only read the summary below and never opened the file.

### The core path (read all of these before presenting)

1. **This document.** You're here. Re-read §1 and §2 until you could say them without
   notes.
2. **[`docs/architecture/README.md`](docs/architecture/README.md)** — the front door
   to the technical documentation. Covers the same ground as §1–2 above but goes
   deeper into *why* the three-unknowns idea works and how the whole repository is
   organized. *One sentence: "This is the map of everything else."*
3. **[`docs/architecture/zbridge-module.md`](docs/architecture/zbridge-module.md)** —
   a tour of the actual Go code, package by package, including *why several parts are
   deliberately incomplete* and how they fail honestly instead of guessing. *One
   sentence: "Here's exactly what's finished, what isn't, and why that's not a
   problem."*
4. **[`docs/architecture/evidence-ladder.md`](docs/architecture/evidence-ladder.md)**
   — exactly how the two solved unknowns got solved, rung by rung, with what was
   actually run and what it actually proved. *One sentence: "Here's the receipts."*
5. **[`docs/architecture/wpl-svc35-mechanism.md`](docs/architecture/wpl-svc35-mechanism.md)**
   — the single most technical document, tracing a message from a Go string all the
   way to bytes on the wire, one step at a time. This is where the WPL byte layout
   (§4 below) is explained in full. *One sentence: "Here's the byte-level proof that
   we're not guessing."*
6. **[`docs/architecture/emulation-harnesses.md`](docs/architecture/emulation-harnesses.md)**
   — precisely what QEMU and Hercules can and can't tell you, and why two different
   tools were needed. *One sentence: "Here's exactly where our evidence's authority
   runs out."*
7. **[`docs/architecture/c4/README.md`](docs/architecture/c4/README.md)** — the whole
   system redrawn as four diagrams, zooming from "what is this system" down to actual
   function calls. *One sentence: "Here's the picture version of everything above."*
8. **[`docs/architecture/c4/sequence-state-object-diagrams.md`](docs/architecture/c4/sequence-state-object-diagrams.md)**
   — three more diagrams: the live demo's timeline across all five systems it
   touches, every state the emulated mainframe can be in (and the one mistake that
   loses work), and the real byte values from one actual captured run. *One sentence:
   "Here's what the demo looks like happening, and here's the actual numbers."*
9. **[`docs/architecture/testing.md`](docs/architecture/testing.md)** — the seven
   different things this project calls "a test," including the very unusual one where
   the expected answer comes from a real machine's output rather than a guess. *One
   sentence: "Here's how we know any of this actually works, mechanically."*
10. **[`RUN.md`](RUN.md)** — how to actually run everything, including the live demo
    script with exactly what to say at each step. *One sentence: "Here's the button
    to press."*

### The deeper path (for when you need to defend a specific claim)

Read these on demand, when a question needs a primary source, not up front.

| If asked about... | Read |
|---|---|
| Why Hercules/emulation at all, and its limits | [ADR 0001](docs/decisions/0001-emulation-strategy-hercules-two-track.md) |
| Why the production code is shaped the way it is | [ADR 0003](docs/decisions/0003-production-bridge-module-architecture.md) |
| The three corrections to the original plan, and why cgo was ruled out entirely | [ADR 0004](docs/decisions/0004-roadmap-corrections-and-cgo-scope-closure.md) |
| Whether the old test system is a valid stand-in for the real one | [Hypothesis H001](docs/hypotheses/001-mvs38j-svc35-wto-oracle.md) |
| Whether the s390x port is behaviorally correct | [Hypothesis H002](docs/hypotheses/002-s390x-port-equivalence.md) |
| Raw, captured proof of any specific result | `docs/evidence/E0-tk5-boot-2026-07-26.md`, `E1-E3-wto-layout-and-svc35-2026-07-26.md`, `E-L-s390x-port-qemu-2026-07-25.md` |
| What's left, and what it's blocked on | [`docs/roadmap-2026-27.md`](docs/roadmap-2026-27.md) |
| Corrections to the original roadmap PDF | [`docs/roadmap-errata.md`](docs/roadmap-errata.md) |

---

## 4. The one technical story worth being able to tell in detail: the WPL

If your mentor only asks about one thing in depth, it will likely be this, because
it's the actual engineering result. Here it is at the level of detail you need to
field follow-ups.

### 4.1 The problem

To send a message, you must hand `SVC 35` a block of bytes (the WPL) in *exactly* the
right shape. Get it wrong, and at best nothing happens; at worst you get garbled or
truncated output on a real operator's screen. **Nobody could find where this exact
byte layout is written down in prose, in any IBM manual** — not because it's secret,
but because mainframe parameter lists like this one are conventionally built by a
**macro** (§2.3), and IBM documents the macro's *keywords*, not the raw bytes it
expands into.

### 4.2 The insight that solved it

Every mainframe assembler can be told to print the *fully expanded* source it actually
assembled, keyword macros and all — that's `PRINT GEN` (§2.3). So instead of hunting
for a manual that describes the format, the project simply **assembled one minimal
message using IBM's own macro, with that print option turned on, and read the exact
bytes IBM's own tooling produced.** That listing *is* the specification — nothing was
inferred or guessed.

### 4.3 The layout, in plain terms

Four pieces, back to back, in this exact order:

1. **Two bytes: how long the whole thing is** — but this trips people up, so be
   precise: it's not the length of just your message text. It's your message's
   length, **plus 4**, because it also counts these next two fields' own bytes.
2. **Two bytes: on/off switches (MCS flags)** — all zero, for the simplest message
   form used in this project.
3. **Your message text**, translated into EBCDIC (§2.6).
4. *(Not used in this project's current form, but worth knowing it exists: if you ask
   for extra features like "send this to a specific group of consoles," two more
   fields get appended after the text — and, this is the trap — the length field from
   step 1 does **not** grow to include them. This project's code refuses to build that
   extended form at all right now, specifically because the exact meaning of every one
   of its on/off switches hasn't been independently confirmed from a citable source
   yet — it would rather say "I can't do that yet" than guess.)*

### 4.4 The one thing that's still genuinely open, and you should say so plainly if asked

After `SVC 35` runs, can you read back a status code saying "it worked"? **On the old
test system, the honest answer is no — nothing is returned at all for this simple
message form.** Whether the real, current mainframe operating system behaves
differently is a documented, open question — the project's own internal notes record
it as unresolved with a citable source still needed, and the code is deliberately
written so that "nothing came back" is a real, representable outcome rather than being
silently confused with "it came back as zero, meaning success." **If asked "so how do
you know it will report failures correctly," the honest answer is: that specific
piece is not yet proven, and the code is written not to assume an answer either way.**
That kind of answer — precise about exactly what is and isn't known — is a strength in
front of an experienced reviewer, not a weakness.

### 4.5 The number to have ready

For a message that's fits within the accepted length, the formula is always: **byte
length written in the parameter list = number of characters in your message, plus
4.** If you remember only one number from this whole document, make it that one — it's
concrete, it's checkable, and it's exactly the kind of specific detail that
demonstrates real understanding rather than a rehearsed summary.

---

## 5. The demo itself

The live demo is real, and it's simple to narrate once you've run it a couple of
times. Full script with exact commands: [`RUN.md`](RUN.md) §7. The shape of it, so you
can describe it before you even open that file:

1. You type a message.
2. A Go program, running on an ordinary laptop, builds the exact bytes described in §4.
3. Those bytes get submitted to the old test mainframe as a small program.
4. The mainframe assembles it, links it, and runs it — running it means executing the
   `SVC 35` instruction.
5. The message appears on the (emulated) operator's console, and every step reports
   back "no problem" (return code zero).

Boot-up (getting the emulated mainframe from off to ready) takes 30–90 seconds and is
the only slow part — start it early and talk over it, per `RUN.md` §7's suggested
script.

---

## 6. One-page cheat sheet

Print this, or keep it on a second screen.

| Question | Answer |
|---|---|
| What does this project build? | `WTO(message string) error` — a Go function that writes to a mainframe operator console, with no C anywhere in the path |
| What's the novel contribution? | The first published proof this can be done in pure Go assembly, without going through C — checked, nobody else has published this |
| What are the three unknowns? | U1: does our low-level code run correctly? U2: is the message format byte-correct? U3: does it work on the real, current system? |
| What's resolved? | U1 (via QEMU) and U2 (via a free, legal 1981 ancestor system under Hercules) — both **without** real mainframe access |
| What's open? | U3 — needs real, entitled access to a current mainframe, which doesn't exist yet for this project |
| What's the central proof? | Bytes built entirely by Go code were accepted by a real `SVC 35` instruction and displayed on a console — captured, reproducible, and regression-tested on every code change since |
| The one formula to know | Message length field = number of characters + 4 |
| The one open technical question | Whether the real system returns a status code the way the old test system doesn't |
| Where's the demo script? | [`RUN.md`](RUN.md) §7 |
| Where's the full reading path? | §3 of this document |
