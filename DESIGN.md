---
name: scopeweaver-cli landing
description: A portable command line product presented through concrete commands and evidence.
colors:
  bg-base: "#070b14"
  bg-surface: "#0d1424"
  bg-elevated: "#121c31"
  text-primary: "#f8fafc"
  text-secondary: "#cbd5e1"
  text-muted: "#94a3b8"
  accent-blue: "#38bdf8"
  accent-blue-strong: "#2563eb"
  border-subtle: "#1e293b"
  status-pass: "#34d399"
  status-warn: "#fbbf24"
typography:
  body:
    fontFamily: "Schibsted Grotesk, Segoe UI, Malgun Gothic, sans-serif"
    lineHeight: 1.6
  code:
    fontFamily: "JetBrains Mono, Cascadia Code, monospace"
rounded:
  sm: "6px"
  md: "10px"
  lg: "14px"
---

# scopeweaver-cli design

## Overview

The current visual source is landing/styles.css and the verified agy-generated page. It presents a dark console with readable technical examples, restrained blue interaction cues and a faint scope-grid motif. The inherited GUI landing is archived under docs/upstream-landing.

## Colors

The CSS custom properties are authoritative. Blue identifies links, focus and interaction; green and amber distinguish recorded success and limitations. Neutral surface levels separate commands, explanations and evidence without implying a live application.

## Typography

Schibsted Grotesk carries prose and headings; JetBrains Mono carries commands and identifiers. Both Latin variable fonts are self-hosted with their OFL licenses. Korean falls back to local Hangul fonts and keeps words together while long identifiers may wrap.

## Layout

The main container is at most 1200px. Split sections stack on narrow screens; grid columns use shrinkable tracks and code scrolls within its own container. At widths up to 980px the header participates in normal flow. Primary controls retain at least 44px touch height.

## Elevation & Depth

Dark tonal layers and hairline borders separate the terminal, requirements and evidence. Shadows support the existing terminal framing; visual changes should preserve the checked composition.

## Shapes

The scope-grid mark is the identity. Surfaces use the CSS 6px, 10px and 14px corner tokens. Code and factual labels remain visually distinct from action controls.

## Components

The hero contains POSIX and PowerShell command tabs with keyboard selection and copy. Language controls toggle EN/KO and preserve the choice. Copy success is announced; clipboard failure selects the visible code and announces manual-copy guidance. Architecture JSON is explicitly an illustration, not captured stdout. Release actions show unreleased status until publication is verified.

## Do's and Don'ts

Keep both languages paired, paths relative, fonts local and requirements explicit. Preserve no-network static rendering and reduced-motion behavior. Do not present historical evidence as a final-package pass, invent release links, expose credentials or imply a live hosted testing service.
