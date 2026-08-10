# The Paper That Says What My Whole Repo Says

**Paper:** Wu, Yue, Zhang, Wang, Wu (2024). *StateFlow: Enhancing LLM Task-Solving through State-Driven Workflows*. arXiv:2403.11322.

*Chapter in progress — the article version is scheduled.*

StateFlow argued that LLM task-solving improves when control flow lives in a state machine instead of a prompt. Every machine in this book is that argument taken further than the paper went.

**Claim under test.** States own control flow; the model fills states. Implementing it for real forces four decisions the paper never specifies: a closed signal alphabet, explicit budgets, distinguished terminal states, and who owns the exit.
