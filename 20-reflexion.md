# The Retry That Remembers Why It Failed

**Paper:** Shinn, Cassano, Gopinath, Narasimhan, Yao (2023). *Reflexion: Language Agents with Verbal Reinforcement Learning*. arXiv:2303.11366.

*Chapter in progress — the article version is gated on the runtime addition.*

Reflexion is memory across retries: the agent writes down why the last attempt failed and reads it before the next one.

**Claim under test.** On this substrate it is a data-flow edit — label the failure evidence, include it at prompt assembly — which turns the paper's architecture diagram into a two-line diff and turns the design question into what, exactly, deserves remembering.
