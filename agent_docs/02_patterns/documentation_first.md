# Documentation-First Engineering

This codebase treats documentation as an executable architectural contract. We never write code without first defining and aligning on the design.

---

## 1. The Design-First Contract

Before any code modification starts:
1. **Define the Goal**: Draft a structured design document under `agent_docs/04_plans/<feature-name>/design.md`.
2. **Align on Architecture**: Detail coordinate mathematics, API models, network schemas, and UI layouts inside this design document.
3. **Structure the Backlog**: Maintain an explicit, living implementation backlog inside the design document containing `## Pending`, `## Current`, and `## Completed` tasks.

---

## 2. Breakdown and Auditing

Once the design is finalized and approved:
1. Run the `breakdown` CLI tool to generate individual, step-by-step markdown task files:
   ```bash
   breakdown -v agent_docs/04_plans/<feature-name>/design.md agent_docs/04_plans/<feature-name>/breakdown-output
   ```
2. **Audit**: Step through each generated file to prune redundant setup tasks, ensure correct dependency sequencing, and guarantee strict testing bounds are defined.

---

## 3. The Execution Loop (RED $\rightarrow$ GREEN $\rightarrow$ REFRACTOR $\rightarrow$ UPDATE)

For each task in the breakdown checklist:
1. Read the task markdown file.
2. Formulate the test changes and code updates.
3. Propose them clearly to the product owner/user.
4. **Halt!** Wait for explicit permission to proceed.
5. Write the failing tests (RED).
6. Implement the feature code (GREEN).
7. Refactor the implementation and verify all tests pass.
8. Update the backlog in `design.md` immediately, moving the task from `Pending`/`Current` to `Completed`.
