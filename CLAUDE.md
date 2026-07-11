## Project Overview

This is a project designed to take an AI harness and remix it to work within a web browser extension (works in both Chrome and Firefox). This browser extension is designed to run as a note-taker and automated task runner whenever a user goes to a web page or requests for an action to be done. Everything runs locally via WebLLM (in the future, cloud-based models can also be used, but as of now, only local models are supported). The key feature of this note-taker is that we index everything into sections to allow for a knowledge tree to grow and gather connections between topics. This allows for better searching and organizing of notes, fully automated, with the user only needing to outline what they want.

This extension also acts as a quick way to go between tabs and ideas via our indexing method. It is similar to macOS Spotlight, but with a more advanced search and organization system.

## Rules for Coding and Contributing

- This project is open source, so please make all code readable and documented. Every line should have a comment if it is confusing or be written in a clean way; follow the coding guidelines for TypeScript as an example.
- Whenever you write code, write as if you are a Senior Engineer. You must do things in the simplest way possible and with the least amount of lines of code. Don't recreate the wheel or add on additional dependencies to the project.
- If you do need to add a dependency, please ask the human developer before doing so, as there may be existing libraries that can be used or code that already exists.
- Be sure to add tests for your code. Please use the testing framework vitest; you must always do test-driven work.
- Always Google the most recent documentation on the library you are using before making any sort of change to the code.
- Whenever you find an error, bug, or issue, please report it to the human developer for review. Once there is a fix, please add a note to the `ai-shared-memory.md` file at the root of the project so other AI agents can also benefit from the fix.

## Project Dependencies

- `web-extension-docs.md`: explains what template we are using for this extension
- pnpm
- Node.js
- Vite
- TypeScript
- React
- WebLLM
- Tailwind CSS
- webextension-polyfill
- vitest
