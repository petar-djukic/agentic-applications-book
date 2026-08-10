# Retrieval That Admits It Found Nothing

**Papers:** Yan, Gu, Zhu, Ling (2024). *Corrective Retrieval Augmented Generation*. arXiv:2401.15884. Jeong, Baek, Cho, Hwang, Park (2024). *Adaptive-RAG*. arXiv:2403.14403.

*Chapter in progress — the article version is scheduled.*

CRAG's evaluate-then-correct and Adaptive-RAG's routing both run here as declared transitions.

**Claim under test.** Retrieval should grade its own evidence. The honest improvement is a terminal state the papers never define: when refinement still finds nothing, the machine says so instead of generating over weak evidence.
