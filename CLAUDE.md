## Project Overview

This project is a centralized bookmark and notes manager designed to operate through Discord and a micro-services architecture. Instead of relying on a browser extension, users interact with the system via a Discord bot, capturing URLs, submitting notes, and requesting automated tasks directly from their servers.

The system uses the OpenAI SDK to automatically process, tag, and organize incoming information. Everything is indexed into sections to cultivate a "knowledge tree," which allows the system to autonomously draw connections between distinct topics. This provides an advanced, automated search and organization system where the user only needs to drop in a link or outline a thought, and the bot handles the heavy lifting.

Data persistence is handled locally via SQLite, and the modular micro-services approach ensures that the Discord interface, the AI processing, and the database operations can scale and be maintained independently.

## Rules for Coding and Contributing

- This project is open source. Make all code readable and documented. Every line should have a comment if it involves complex logic; follow standard Golang by Google coding guidelines for clean, maintainable syntax.
- Write code from the perspective of a Senior Engineer. Implement solutions in the simplest way possible with the minimal amount of code. Do not recreate the wheel or bloat the project with unnecessary dependencies.
- If you believe a new dependency is required, ask the human developer before installing it. There may be existing libraries in the micro-services or native Go packages that already solve the problem.
- Always write tests for your code. We use basic golang testing framework as our testing framework; you must adopt a test-driven development (TDD) approach.
- Always check the most recent documentation for the libraries you are using (especially `discordgo` and the `openai` SDK) before implementing changes or using new methods.
- Whenever you resolve an error, bug, or architectural issue, report it to the human developer. Once a fix is implemented, document it in the `ai-shared-memory.md` file at the root of the project so other AI agents can learn from the solution.
- We also use makefile to help manage and group individual tasks and projects, such as running and building. Please use the provided makefile to build and run the project, and update if needed

## Project Dependencies

- Golang
- discordgo (for the bot interface)
- OpenAI SDK (for LLM processing and automated indexing)
- SQLite (for the database) with go-sqlite3
