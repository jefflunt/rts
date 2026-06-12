# Context Management & Progressive Disclosure

As codebases scale, loading every single file, history, and dependency into an AI agent's active memory causes "context bloat," causing high latency, cost, and hallucinated errors. This project enforces **Progressive Disclosure** and **Continuous Alignment** to keep context highly targeted.

---

## 1. Core Strategies

### 🔍 Progressive Disclosure
An agent or developer should only load information at the level of detail currently required:
- **Level 1 (Orientation)**: Read `01_orientation/` to understand *what* the project is and *how* components interact at a system level.
- **Level 2 (Patterns)**: Read `02_patterns/` to understand the *rules of the road* (styling, TDD, interfaces).
- **Level 3 (Deep Dives)**: Read `03_deep_dives/` *only* if the feature touches a complex subsystem (e.g. minimap cache or network protocols).
- **Level 4 (Atomic Plans)**: Focus solely on a single, isolated markdown task inside `04_plans/<feature>/breakdown-output/`.

---

## 2. Best Practices for Developers and Agents

1. **Avoid Bulk Reads**: Never run broad commands like `cat **/*.go` or read every file in full. Use targeted searches (`grep` or `glob`) and read files with specific line limits/offsets when possible.
2. **Atomic Commits & Edits**: Work on one small subtask at a time. Run unit tests immediately to verify changes before editing subsequent modules.
3. **Keep Docs as Index**: The `agent_docs/` folder serves as a living, high-fidelity external index of the codebase. By maintaining it continuously, we ensure that newly spawned agents or developers can get up to speed in seconds.
