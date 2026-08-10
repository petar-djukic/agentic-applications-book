# ReAct Is Fifteen Lines. The Other Forty-Five Are the Job.

**Paper:** Yao, Zhao, Yu, Du, Shafran, Narasimhan, Cao (2022). *ReAct: Synergizing Reasoning and Acting in Language Models*. arXiv:2210.03629.

*Chapter in progress — the article version is scheduled.*

ReAct's interleaving of reasoning and acting is real and it works. The loop is about fifteen lines.

**Claim under test.** The paper's loop trusts the model to exit honestly. The machine version routes the model's "done" into a build-lint-test chain it cannot talk its way past — and that chain, not the loop, is where the engineering lives.
