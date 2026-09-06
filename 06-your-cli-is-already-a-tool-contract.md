# Your CLI Is Already a Tool Contract

## Learning objectives

After reading this chapter you will be able to:

1. State which parts of the tool contract a conventional CLI already satisfies.
2. Name the parts a declaration has to add, and say why the CLI cannot carry them.
3. Declare an exec tool with argv mapping and bounded stdin.
4. Predict which signal a command emits from how it failed.
5. Recognize when a shell script has crept back into a machine, and what it costs.

## Conventions you already follow

A well-behaved command-line utility has a contract, and its author
probably never called it that. Flags are named, typed inputs. Standard
input is where the payload arrives when it is too big or too awkward
for argv. Standard output is the result. The exit code says whether it
worked. Those four conventions are old enough to be invisible, and
they are most of what a tool declaration needs.

Wiring a CLI into a declarative agent is therefore mostly a matter of
saying what is already true.

| Convention | What the declaration calls it |
|---|---|
| The command itself | `binary` |
| Fixed leading arguments | `args` |
| Named flags and positional arguments | `flag`, `positional`, `bool_flag` on a parameter |
| Payload on standard input | `stdin_source`, with `stdin_max_bytes` |
| Result on standard output | the word's `Result.Output`, raw or structured |
| Exit code | the emitted signal |

The swarm's round counter is the smallest possible example, and it is
a real word in a running machine:

<!-- listing: c6-1 source=large-context-swarm/agents/rlm-root/declarations.yaml -->

```yaml
    binary: jq
    args: ["-c", ". + 1"]
    stdin_source: $from(round_count).output
    stdin_max_bytes: 32
```

That is the entire declaration, and one line of it is a filter
program. To increment a number, the agent runs `jq`, which reads the
number on standard input and writes the successor to standard output.
No Go was written to make that happen, and the machine gained an
arithmetic step that an operator can read.

The `stdin_source` line points at a labelled earlier result rather
than at the previous word, so the counter can sit anywhere in the
machine and still be found. What arrives on the pipe is bounded, and
the bound is enforced by refusal. An input over `stdin_max_bytes`
fails before the process starts; the runtime does not truncate it into
something smaller that still parses.

> **Good Practice.** Set `stdin_max_bytes` to the size the input should
> be, not the size it might reach. Thirty-two bytes for a counter is a
> statement about what this word is for, and a run that violates it has
> a defect worth stopping on.

## What the exit code buys

Signal mapping is where a CLI's oldest convention pays for itself.
Three outcomes come out of the runtime's exec contract. A successful
exit maps to `ToolDone`, a non-zero exit or a failure while running
maps to `ToolFailed`, and a failure that prevents the command from
being constructed at all maps to `CommandError`.

The third case is the one to dwell on. If the `stdin_source` selector
names a label that does not exist, or resolves to an empty value, the
subprocess never launches and the machine sees `CommandError`. A `jq`
that ran and failed takes a different edge. A shell pipeline reports
both through one exit status.

The split is not universal. A failed precondition stops the launch
too, and so does a missing required parameter. Both of those still
emit `ToolFailed`. Selector resolution is where the two outcomes come
apart cleanly.

The swarm's counter uses this. Its declaration spells out one
condition for `CommandError`, covering a round counter that is absent,
malformed, empty, or over the declared bound. It leaves `ToolFailed`
to the general contract.

## What the CLI cannot tell you

Those conventions stop short in exactly the place Chapter 4 spends its
time. Argv carries no statement about what the command touches outside
the process, no classification of whether the effect can be reversed,
and no receipt for reversing it. A `rm` and a `cat` present identical
interfaces to a caller reading only flags and exit codes.

So the declaration adds what the CLI never said: a `side_effects`
block naming the kind and target, a reversibility classification, an
undo strategy, and the errors the word can emit with what each leaves
behind. None of that is discoverable from the binary, which is why the
contract is a file rather than an introspection.

The declaration also draws a boundary the model never sees. Every
argv-mapping extension is stripped from the tool schema sent to the
model, so a model may supply a parameter's value while the machinery
that turns it into `--flag value` stays on the machine's side of the
line.

## Where the analogy strains

The swarm's dispatch word is `binary: sh`, and its arguments are a
shell script:

<!-- listing: c6-2 source=large-context-swarm/agents/rlm-root/declarations.yaml -->

```yaml
        intent="$(cat)"
        id="finding-$(printf %s "$intent" | cksum | cut -d' ' -f1)"
        jq -cn --arg content "$intent" --arg id "$id" --slurpfile seed worker-seed.json \
          '$seed[0] + {content: $content, id: $id}' > worker-request.json
        exec "${RLM_AGENT_BIN:-agent}" \
          --profile "${RLM_APP_DIR:-.}/agents/rlm-worker/profile.yaml" \
          --request worker-request.json
    stdin_source: $from(intent).value
    stdin_max_bytes: 4096
```

Read it and the reason is legible. One intent arrives on standard
input, gets a deterministic id from a checksum, is merged with a seed
file into a request, and the merged request is handed to a child agent
— because a child agent reads its request from a file rather than a
pipe. A pipeline, a redirect, and a process handoff, all inside one
word.

The runtime did not invite this. Its exec contract is explicit that
`stdin_source` is not a pipeline authoring surface, and that one word
launches one binary with one argv. Nothing stops that binary from
being `sh`, and when the substrate has no word for "merge two JSON
documents and hand the result to a child process", a shell script is
what an author reaches for.

The cost is checkability, and it is worth stating plainly. Every other
word in this machine states what it does in fields a reader can
compare against what the runtime actually did. That comparison is
imperfect, as Chapters 4 and 9 both show, because a declaration can
misdescribe the word it belongs to. This one puts a seven-line program
inside a YAML string. A typo in it is a runtime failure rather than a
load-time one, and there are no fields to compare against anything.

> **Common Error.** Reaching for `binary: sh` because a step needs two
> commands. Two commands are usually two words with a transition
> between them, which is the arrangement the runtime's own guidance
> recommends: one exec word models one CLI verb, and multi-step work
> belongs in the transition table. The dispatch word earns its script
> because the middle of it is a file handoff to a child process, not
> because it is doing two things.

## Summary

A conventional CLI already supplies most of a tool contract: the
binary is the action, flags and positionals are typed parameters,
standard input is the payload, standard output is the result, and the
exit code is the outcome. A declaration says those things in a file.
It then adds what the command line has no way to express: which
effects reach outside the process, whether they can be reversed and
how, and what each failure leaves behind. Exit-code mapping gains what
a pipeline loses, separating a command that ran and failed from one
that was never constructible. Watch for `binary: sh`. It is available
and sometimes correct, and every line inside it is a line no check can
read.

## Terms

| Term | Definition |
|---|---|
| **Exec tool** | A word whose action is running one external command with one argv |
| **Argv mapping** | The extensions that turn declared parameters into command-line arguments: `flag`, `positional`, `bool_flag`, and the `default`, `position`, and `source` modifiers |
| **`stdin_source`** | A selector naming the earlier labelled value delivered on standard input |
| **`stdin_max_bytes`** | The byte bound on that input, enforced by rejection before launch |
| **`ToolDone` / `ToolFailed`** | The signals for a successful exit and for a command that ran and failed |
| **`CommandError`** | The signal for a command that could not be constructed, such as an unresolvable stdin selector |
