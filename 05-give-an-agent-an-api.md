# How to Give an Agent an API

## Learning objectives

After reading this chapter you will be able to:

1. Wire an existing REST endpoint into an agent as a declared operation.
2. Say where authority for the request shape lives, and who is allowed to change it.
3. Trace a request body from declared parameters through the input mapping to the wire.
4. Choose between carrying a value forward and publishing it under a label.
5. Bound an agent's network reach with a declared allowlist.

## The runtime is the client

An agent needs to call an API. Usually somebody writes a wrapper that
builds the request, sends it, checks the status, pulls out the fields
the agent needs, and registers itself under a name the model can call.
Multiply by every endpoint and the harness becomes a client library
with a model attached.

Under a declarative harness nobody writes that function. A file
describes the endpoint — method, path, what goes in the body, which
status counts as success, which parts of the response survive — and
one compiled client executes every operation described that way. A
tool is a declaration, and the runtime is the client.

That relocation is the whole chapter, and it leaves one question to
ask of any agent framework. When the API changes, what has to be
edited, and by whom?

## The shape of an operation

The swarm in this book reads and writes a Chroma collection over HTTP.
Its REST vocabulary lives in `examples/catalog/agents/knowledge-
manager/corpus-rest.yaml`, a file the book copies from upstream
without modification. One operation from it, the filtered query the
root uses to collect its workers' findings, is declared like this:

```yaml
query_records_filtered:
  method: POST
  path: /api/v2/.../collections/{collection}/query
  params:
        body_schema:
      type: object
      required: [query_embeddings, where, where_document, n_results]
      properties:
        query_embeddings: {type: array}
        where: {type: object}
        where_document: {type: object}
        n_results: {type: integer}
    body_source: previous_result
    input_mapping:
      collection: $.mapped.collection_id
      query_embeddings: $.carried.query_embeddings
      where: $.carried.where
  success: {status: [200], signal: QueryResponded}
  response:
    output:
      ids: $.ids
      documents: $.documents
      metadatas: $.metadatas
```

The full path and two of the five mapping lines are elided for the
page; nothing else is. Six clauses do the work, and not one of them is
code.

| Clause | What it settles |
|---|---|
| `method`, `path` | The address, with `{collection}` a hole filled at call time |
| `body_schema` | Types the request and marks four fields required |
| `body_source` | Where the values come from |
| `input_mapping` | Which incoming value fills which parameter |
| `success` | Which status codes pass, and the signal the machine sees |
| `response.output` | What survives into the word's result |

Two rows repay a second look. Marking those four fields required is
not tidiness. The body renderer substitutes a declared parameter that
was not supplied with its zero value instead of dropping the field, so
an optional `n_results` would arrive as a null result count and Chroma
would reject it. Declaring the filters required on a separate
operation is why the older unfiltered query still sends the request it
always sent.

The `response.output` row is where a reader is most likely to draw the
wrong conclusion. It selects what survives into the word's result, and
this operation selects all three fields, `documents` included. Nothing
here keeps document text away from a caller. That job belongs to the
calling word's own output schema, which is where Chapter 9 and the
swarm's handle discipline actually live.

## Who calls it

An operation says what the endpoint is. A word in a machine says when
to call it, and the binding is two lines:

<!-- listing: c5-1 source=large-context-swarm/agents/rlm-root/declarations.yaml -->

```yaml
    config:
      rest_ref: chroma
      operation: query_records_filtered
```

`rest_ref` names the client, `operation` names the entry in that
client's vocabulary, and there is no third line. No URL appears here,
and no HTTP knowledge of any kind. Changing those two lines is how
this word comes to call something else, and it is the whole of the
change.

Its declared parameters are the contract on the other side:

<!-- listing: c5-2 source=large-context-swarm/agents/rlm-root/declarations.yaml -->

```yaml
    parameters:
      type: object
      properties:
        collection: {type: string}
        query_embeddings: {type: array}
        where: {type: object}
        where_document: {type: object}
        n_results: {type: integer}
```

Read the two files together and the authority is unambiguous. Request
shape belongs to the REST file, so changing the wire format is one
edit in one place for every consumer. What this particular word
supplies belongs to its declaration. Neither owns both, which is what
keeps a shared vocabulary shared.

> **Good Practice.** Keep the REST file as the single authority even
> when only one word uses an operation today. The second consumer is
> the one that discovers whether the boundary was drawn in the right
> place, and by then moving it is a migration rather than an edit.

## Threading, and why word order is forced

Selectors in the input mapping are the part that surprises people.
`$.carried.query_embeddings` does not read a global. It reads the
*previous* word's result, and only the previous word's.

Previous-result threading is single-hop. A `$.carried` selector sees
what the word immediately before produced and nothing earlier, so a
value moving down a chain of words has to be carried at every step,
which is what `carry_forward` does. Word order in that chain is a
consequence rather than a preference: embedding runs before collection
resolution, which runs before the query, because each word carries the
payload the next one needs. Reorder them and the selectors point at
something that is not there.

The runtime offers a second route, and the swarm uses it twenty-five
times. A transition can attach a `label:` to what a word returns, and
any later word can address it as `$from(label).field` without anything
in between carrying it. The two mechanisms answer different questions.
Use the carried chain for a value the next word consumes; use a label
for a value many words need, which is why the root republishes the
request under one label in its first word and never threads those
fields again.

> **Common Error.** Assuming the request seed is readable everywhere.
> Only the first word of a run sees it, so a field that is not
> republished immediately is unreachable for the rest of the machine.
> The swarm's root spends its first word doing nothing but republishing
> request fields under a label, which looks like ceremony until the
> alternative is threading seven values through fifteen words by hand.

## The boundary

An agent that can call one API can call any API the network reaches,
unless something says otherwise. The same REST file says otherwise, in
a `limits:` block above the clients:

```yaml
network:
  schemes: [http]
  hosts: [127.0.0.1]
  ports: [8000, 11434]
```

Those three lines put the agent's reach at loopback on two ports. A
prompt injection that talks a model into exfiltrating a collection has
nowhere to send it, because the client refuses the connection before
the model's intent matters. The same block carries request and
response size caps, a timeout, and a redirect policy of `none`, so a
302 to somewhere interesting is an error instead of a hop.

The block is named, and each client opts into it by writing
`limits_ref: local_corpus`. That indirection is worth the extra line,
because one profile then bounds every client that names it, and an
operator can read the agent's entire network surface in one place
without reading any code.

## Summary

Giving an agent an API means describing the endpoint rather than
wrapping it: method, path, a typed body schema, the mapping from
threaded values into parameters, the status that counts as success,
and which response fields survive. One compiled client runs every
operation declared that way, so a wire-format change is one edit
rather than one edit per caller. A word binds to an operation in two
lines and carries no HTTP knowledge of its own. Selectors that fill
the request read the previous word's result, which forces the order of
any chain that carries a value forward, while a labelled result stays
addressable to every later word. A named limits profile bounds where
any of it can reach.

## Terms

| Term | Definition |
|---|---|
| **Operation** | A named, declared endpoint: method, path, body schema, success signal, and response mapping |
| **REST client reference** | The `rest_ref` naming which declared client a word invokes |
| **Body source** | Where the request values come from, such as the previous word's result |
| **Input mapping** | Which threaded value fills which declared parameter |
| **Carry forward** | Republishing a value so the next word can reach it, since threading is single-hop |
| **Response output** | The selection of response fields that survive into the word's result; anything unnamed is unreachable |
| **Network allowlist** | The declared schemes, hosts, and ports the client may contact, enforced before any request |
