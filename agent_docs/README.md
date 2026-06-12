# Agent Docs Framework

This directory houses the structured, AI-readable, and human-friendly documentation suite for the Scrollable RTS Tilemap Engine. It adheres to the [**`agent_docs` framework specifications**](https://github.com/jefflunt/agent_docs).

## 🗂️ Folder Structure

- **`01_orientation/`**: System-level high-level context, domain model definitions, and onboarding steps.
  - `architecture.md`: Visual layout of client/server components and network sequencing.
  - `domain.md`: Mathematical translation formulas (Grid $\leftrightarrow$ World $\leftrightarrow$ Screen $\leftrightarrow$ Minimap) and camera viewport culling strategies.
  - `onboarding.md`: Native library dependency setup and runtime commands.
- **`02_patterns/`**: Engineering guidelines, styling patterns, and QA standards.
  - `coding_conventions.md`: Style, package organization, interface requirements, and naming structures.
  - `testing_patterns.md`: Mocking protocols and headless UI testing patterns.
  - `documentation_first.md`: Design-first contract and continuous backlog alignment loop.
  - `context_management.md`: Progressive context disclosure and LLM token optimization strategy.
- **`03_deep_dives/`**: Highly focused analyses of complex technical subsystems.
  - `minimap_caching.md`: $O(1)$ pre-rendered HUD textures and dragging math.
  - `network_handshake_and_protocols.md`: Synchronous TCP validation and JSON/Gob dual socket codecs.
- **`04_plans/`**: Living feature backlogs, design documents, and automated breakdown tasks.
  - `scrollable-tilemap/`: Epic plan for the original tilemap rendering.
  - `minimap/`: Epic plan for the interactive overlay minimap.
- **`templates/`**: Boilerplate markdown forms to standardise pattern, deep dive, and feature planning submissions.

---

## 🧭 How to Navigate

For any developer (human) or AI implementation agent starting in this codebase:
1. **Always begin with `01_orientation/README.md`** to load the system mental model.
2. **Review `02_patterns/README.md`** to understand coding conventions and testing policies.
3. **Reference `03_deep_dives/`** when modifying or implementing features touching the minimap texture caches or TCP sockets.
4. **Use `04_plans/`** to check active tasks or log a new feature design document.
