# Lesson 6 — Concurrency: Go goroutines vs async Python

Both backends in this repo answer the same requests. `backend/` does it with
FastAPI on an asyncio event loop; `backend-go/` does it with the Go runtime.
The code looks different for one reason: **the two languages solve "many
requests at once" in fundamentally different ways.**

This note is only about that. Not speed — mechanism.

---

## 1. Three words that get mixed up

| Term            | Meaning                                                             |
| --------------- | ------------------------------------------------------------------- |
| **Concurrency** | many tasks *in progress* at once (they take turns)                   |
| **Parallelism** | many tasks *executing* at the same instant (needs several CPU cores) |
| **Blocking**    | a task holds its execution slot while doing nothing but waiting      |

asyncio gives concurrency on **one** core. Go gives concurrency **and**
parallelism. That single sentence is most of this lesson.

---

## 2. Four levels: core, process, thread, task

Before comparing the two runtimes, get the stack straight. Levels 1–3 are
identical for both languages. **All the difference lives in level 4 and in how
it maps onto level 3.**

| Level                | Owned by     | What it is                                                                            | Cost of one          |
| -------------------- | ------------ | -------------------------------------------------------------------------------------- | -------------------- |
| **1. Core**          | hardware     | executes instructions. Their number is a hard ceiling on *simultaneous* execution, in any language | fixed                |
| **2. Process**       | OS           | isolated program: own memory, own sockets, own DB pool. Processes cannot see each other's variables | heavy (MBs, ms)      |
| **3. Thread**        | OS           | line of execution **inside** a process; threads share memory. The OS schedules *these* onto cores | ~8 MB stack, µs switch |
| **4. Task**          | the language | coroutine / goroutine. **The OS does not know it exists** — the runtime switches it without entering the kernel | ~KBs, ns switch      |

### Python

```
N cores  (N = 4 in this sketch)
├─ uvicorn process #1
│    ├─ thread 1 ──▶ event loop ──▶ coroutines A, B, C, D…  (one at a time)
│    └─ threads 2..n ──▶ pool for plain `def` handlers
├─ uvicorn process #2   ← own interpreter, own memory, own DB pool
├─ uvicorn process #3
└─ uvicorn process #4
```

Every coroutine of one process lives in **one** thread. A thousand coroutines
still occupy one core. Using N cores means N processes — and N connection
pools.

### Go

```
N cores  (N = 4 in this sketch)
└─ one process
     ├─ thread 1 ─ P1 ─▶ goroutines G7, G12, G3 …
     ├─ thread 2 ─ P2 ─▶ goroutines G4, G9 …
     ├─ thread 3 ─ P3 ─▶ goroutines G1, G22 …
     ├─ thread 4 ─ P4 ─▶ goroutines G8 …
     └─ a few helper threads (GC, netpoller, timers)
```

One process, a handful of threads, thousands of goroutines. The runtime maps
goroutines onto threads; the OS maps threads onto cores.

### Seeing it yourself

The OS only ever shows you levels 2 and 3. Start either backend and ask for
its process and thread count:

```bash
ps -o pid,nlwp,rss,comm -p <pid>     # nlwp = OS threads, rss = resident memory
```

A sample run — 4-core Linux box, both services idle right after start, one
worker each:

|                     | processes | OS threads | resident memory |
| ------------------- | --------- | ---------- | --------------- |
| Go binary           | 1         | 6          | 11 MB           |
| FastAPI (1 worker)  | 1         | 6          | 96 MB           |

Those exact numbers travel badly — they move with the core count, the Python
version and what gets imported. Reproduce them rather than trusting them. What
*does* hold everywhere:

- **Similar thread counts, different meaning.** Go's 6 are execution slots (one
  per core) plus GC and netpoller helpers, and the slots run *your* code in
  parallel. Python's 6 are one event-loop thread plus helpers, and only one of
  them ever runs bytecode.
- **Tasks are invisible to `ps`.** Serving thousands of connections would show
  the same 6 threads in either runtime. That is the whole point of level 4:
  thousands of tasks, a few threads, N cores.
- **Memory is counted per process.** To use every core, the Python service
  needs one worker per core — on this sample box, `--workers 4` means roughly
  380 MB and four separate pools against PostgreSQL. The Go binary stays one
  process with one pool no matter how many cores it uses.

---

## 3. The GIL

CPython has one global lock. **Only one thread can execute Python bytecode at
a time, per interpreter.** The lock is released around I/O (socket reads, disk,
`time.sleep`) and inside C extensions that opt out of it.

Consequences:

- Threads in Python **do** help with I/O-bound work (the lock is released while
  waiting), and **do not** help with CPU-bound work — 4 threads computing on
  4 cores still run one at a time.
- Real CPU parallelism needs **separate processes** (`multiprocessing`, or
  `uvicorn --workers 4`). Each process has its own interpreter, its own GIL,
  its own memory, and — relevant here — its own database connection pool.
- Python 3.13 ships an experimental free-threaded build (PEP 703) without the
  GIL, but it is not the default and not what this project runs.

The GIL is a property of **CPython**, not of the language. Go has nothing
equivalent: any number of goroutines execute Go code simultaneously.

---

## 4. asyncio: one thread, explicit yield points

```
       [ one process, one thread, one core ]
       ┌───────────────────────────────────────────┐
       │  event loop                               │
       │    task A ──await db──▶ suspended         │
       │    task B ─────────────▶ running          │
       │    task C ──await db──▶ suspended         │
       └───────────────────────────────────────────┘
```

Rules:

- A task can only be suspended **at an `await`**, and only if what it awaits
  actually yields to the loop.
- Between two `await`s, the code runs to completion, uninterrupted.
- The GIL is barely relevant here: there is only one thread, so nothing
  contends for it. asyncio's limit is not the GIL — it is having one thread.

**Failure mode:** one blocking call inside `async def` freezes *every* request
in that process:

```python
@app.get("/slow")
async def slow():
    time.sleep(5)      # ← blocks the whole event loop, not just this request
    requests.get(...)  # ← same problem: sync HTTP client
```

The fix is to await a non-blocking equivalent (`asyncio.sleep`, `httpx`), or to
push the work off the loop (`run_in_executor`, `asyncio.to_thread`).

**Colored functions.** A function is either `async` or it is not, and the two
do not mix freely: you cannot `await` from a sync function, and calling an
async one without `await` gives you a coroutine object, not a result. This is
why the ecosystem is split in two — `psycopg` vs `asyncpg`, `requests` vs
`httpx`. Choosing wrong is silent; it just degrades throughput.

### Where FastAPI puts your handler

| Handler declared as | Runs where                                       |
| ------------------- | ------------------------------------------------ |
| `async def`         | on the event loop — must never block             |
| `def` (plain)       | in a worker **thread pool** — blocking is fine   |

Both endpoints in [`backend/app/routers/categories.py`](../../backend/app/routers/categories.py)
are `async def` with `await session.execute(...)`, so they live on the loop.

---

## 5. Go: goroutines and the M:N scheduler

Three entities:

- **G** — a goroutine. ~2 KB of stack that grows on demand. Cheap enough to
  have hundreds of thousands.
- **M** — an OS thread. Real, expensive, few.
- **P** — a logical processor: a slot that permits execution. Count defaults to
  the number of cores (`GOMAXPROCS`).

```
       [ one process, N cores — 4 here ]
       P1 ← M1 ← G7  G12 G3 …     ← each P owns a run queue
       P2 ← M2 ← G4  G9  …
       P3 ← M3 ← G1  G22 …
       P4 ← M4 ← G8  …            ← idle P steals work from a busy one
```

What the runtime does for you:

- **Multiplexing.** Thousands of goroutines share a handful of threads.
- **Work stealing.** An empty P takes goroutines from a loaded P, so the load
  balances itself.
- **The netpoller.** A goroutine waiting on a socket is parked and consumes no
  P at all. Waiting is free; only running costs a slot.
- **Syscall handoff.** If a goroutine enters a genuinely blocking syscall, the
  runtime detaches its thread from the P and puts another thread there. Other
  goroutines keep running — the asyncio failure mode above cannot happen.
- **Preemption.** Since Go 1.14 the scheduler can interrupt a goroutine that
  never makes a function call (a tight loop) via a signal, after ~10 ms. No
  goroutine can starve the others forever.

Because waiting does not occupy a P, `GOMAXPROCS` is not a request limit. A
service that mostly waits on PostgreSQL can hold far more in-flight requests
than it has cores.

### No colored functions

There is no `async` keyword. Any function can be called from anywhere; you
choose concurrency at the *call site* with `go f()`. `net/http` already does
this: **the server starts a goroutine per connection**, which is why
[`backend-go/health.go`](../../backend-go/health.go) is written as ordinary
blocking code:

```go
if err := pool.Ping(r.Context()); err != nil {   // just waits; no async/await
```

While that goroutine waits, the scheduler runs others. Nothing is marked.

---

## 6. Side-by-side

|                                | FastAPI (asyncio)                            | Go (`net/http`)                                |
| ------------------------------ | -------------------------------------------- | ---------------------------------------------- |
| Unit of concurrency            | coroutine (task)                             | goroutine                                       |
| Cost of one                    | ~KBs, cheap                                  | ~2 KB, cheap                                    |
| Scheduled by                   | event loop, **cooperatively**                | runtime, cooperatively + **preemptively**       |
| Yield points                   | explicit — `await`                           | implicit — calls, channels, syscalls            |
| Cores used per process         | **1**                                        | all of them                                     |
| To use N cores                 | N processes (N pools, N× memory)             | nothing to do                                   |
| One blocking call              | stalls **all** requests in the process       | stalls only that goroutine                      |
| Sync/async split               | yes — colored functions, split ecosystem     | none                                            |
| Shared mutable state           | mostly safe: one thread at a time            | genuinely unsafe: needs `sync.Mutex`, `-race`   |
| CPU-bound work                 | blocked by the GIL → separate processes      | just works                                      |

The last two rows are the trade. Go removes the event-loop footguns and gives
real parallelism; in exchange, **data races become possible**. In asyncio, code
between two `await`s cannot be interrupted, which rules out a whole class of
bugs. In Go two goroutines really can write the same variable at the same
instant on two cores. Nothing in `backend-go/` shares mutable state yet (a
`pgxpool.Pool` is safe for concurrent use), but the day it does, the tool is
`go run -race`.

---

## 7. What this means for ishopster

Neither model is a bottleneck for a product catalog: the time is spent waiting
for PostgreSQL, and both wait efficiently. The practical differences are
operational.

- **Process count.** `uvicorn --workers N` = N interpreters, N pools, N× the
  memory. The Go binary is one process using all cores with one pool. That also
  means Postgres sees fewer connections from Go — worth remembering when tuning
  `max_connections`.
- **Failure shape.** Under load, a Python service degrades if anything blocks
  the loop; a Go service degrades when it runs out of CPU or pool connections.
  Different symptoms, different fixes.
- **Contract, not concurrency.** Both backends must answer identically. The
  concurrency model must never be visible in the JSON.

---

## Reading exercise

1. `uvicorn --workers 4` runs 4 processes rather than 4 threads. Why does the
   thread version not help?
2. Every handler in `backend-go/` is plain blocking code with no `async`
   keyword. Where does the concurrency come from?
3. Someone adds `time.sleep(2)` to an `async def` FastAPI endpoint and the same
   `time.Sleep(2 * time.Second)` to a Go handler. What does each do to *other*
   users' requests?
4. `GOMAXPROCS` defaults to the number of cores — say 4. Can the Go service
   still serve 1000 requests concurrently? What actually limits it?
5. Two requests arrive at once and both read `pool`. Why is that safe, and what
   would have to be true for it to become a race?
6. `ps` shows a similar handful of OS threads for either service. Why does that
   number say nothing about how many requests each one can handle?

---

## Answers

**1.** The GIL. Four threads still execute Python bytecode one at a time, so a
CPU-bound handler gets no parallelism. Separate processes each get their own
interpreter and their own GIL. (For purely I/O-bound work threads *would* help
— the GIL is released while waiting — but that is what the event loop already
does, more cheaply.)

**2.** From `net/http` itself: its accept loop starts a goroutine per
connection, so every handler already runs in its own goroutine. Concurrency is
in how the handler is *invoked*, not how it is written — the opposite of
Python, where it must be declared with `async def`.

**3.** Python: the whole event loop stops for 2 seconds — every request in that
process is stalled, including health checks. Go: only that one goroutine
sleeps; the scheduler runs everything else, and other users notice nothing.

**4.** Yes, easily. Goroutines waiting on the network are parked by the
netpoller and occupy no P, so `GOMAXPROCS` limits only simultaneous *execution*
of Go code. The real ceiling is the **connection pool** (`pgxpool` defaults to
`max(4, runtime.NumCPU())` connections) and PostgreSQL itself — request 1000
waits for a free connection, not for a CPU.

**5.** `*pgxpool.Pool` is documented as safe for concurrent use — it guards its
internal state with locks, and each request checks out a connection for its own
exclusive use. It would become a race the moment we introduced shared mutable
state that is *not* internally synchronised: a plain `map` cache, a request
counter as a bare `int`, a config struct someone reloads at runtime. Those need
`sync.Mutex`, `sync/atomic`, or a channel — and `go run -race` to prove it.

**6.** Because requests are handled at level 4, not level 3. A goroutine
waiting on PostgreSQL is parked by the netpoller and holds no thread; a
suspended coroutine holds no thread either. Both services can have far more
requests in flight than they have threads, and the thread count barely moves.
What the number *does* tell you is how much can execute at the same instant: as
many of Go's threads as there are cores can run Go code in parallel, while
exactly one of Python's runs bytecode at a time.

---

Next: the remaining catalog endpoints in `backend-go/` — path parameters,
`QueryRow`, and mapping a missing row to 404.
