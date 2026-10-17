# scopeweaver-cli product

<!-- impeccable:product-schema 1 -->

## Platform
web

## Stack
The product is a Go command line application with its own engine and managed PostgreSQL. Its introduction is a static HTML/CSS/JavaScript page in landing/, with no backend or external assets.

## Users
Developers and coding agents performing source research and technical validation in locally isolated environments. The public introduction supports English and Korean.

## Product Purpose
Make ScopeWeaver's backend features discoverable and executable through named CLI commands, JSON output, schemas, explicit mutation confirmation, and persistent local state.

## Positioning
A separate standalone repository and portable folder. The first operational command downloads pinned runtime dependencies; subsequent core use needs no separate ScopeWeaver, PostgreSQL installation or OS service.

## Operating Context
Users unpack a folder they can write to. Model inference needs configured provider access. Subscription profiles use the official Codex CLI alongside the portable core. External integrations retain their own requirements.

## Capabilities and Constraints
The catalog contains 297 named operations and 857 parameters. Catalog parity is not evidence that every integration passed live testing. ChatGPT OAuth has passed built-in workflows in a Windows candidate; Windows R3 OAuth passed 125 installed calls; Windows R4 GLM-5.3-Flash/low passed 127 calls through chat, planner, worker, finding and report with a replacement Coding Plan key. Flagship GLM-5.3 timed out at the final planner call and its full workflow is not accepted. The dated catalog union snapshot is 284 invoked / 283 positive / 297 total, not all-feature acceptance. Final rebuilt-package checks remain pending. No public CLI release exists yet. See docs/CLI-VERIFICATION.md for dated evidence and limits. Usage restrictions prohibit online target testing, including owned or authorized targets.

## Brand Commitments
Retain ScopeWeaver's scope-grid mark, precise console language, restrained blue accents and English/Korean parity. Credit ARTEX and upstream ScopeWeaver under AGPL-3.0. The landing page uses the user-requested agy Gemini 4 Argon Medium design.

## Evidence on Hand
Source and schema, installed CLI and lifecycle receipts, independent agent checks, built-in OAuth E2E, and browser QA reports. Historical upstream GUI materials are preserved in docs/upstream-landing and are not the CLI landing surface. No testimonials, adoption metrics or independent security-effectiveness claims are supplied.

## Product Principles
Show real commands and clearly labeled examples. State first-use networking and provider requirements near setup. Distinguish completed evidence, historical tests and open release gates. Do not publish secrets, private targets or user conversations.

## Accessibility & Inclusion
Responsive desktop/mobile layout, keyboard tabs and copy controls, visible focus, reduced motion and local fonts. English and Korean share one page and persist the language choice.
