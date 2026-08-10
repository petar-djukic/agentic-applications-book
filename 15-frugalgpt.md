# The Cheap Model Answers Most Questions

**Papers:** Chen, Zaharia, Zou (2023). *FrugalGPT*. arXiv:2305.05176. Ong, Almahairi, et al. (2024). *RouteLLM*. arXiv:2406.18665.

*Chapter in progress — the article version is scheduled.*

Model routing in the papers is a classifier and a cost curve.

**Claim under test.** Routing in production is also the fallback path, and the machine version declares it: every way the router can fail routes to the fast tier, so a broken classifier degrades cost, never availability.
