# PaperViz Design System

## Overview

PaperViz's one job is transformation: dense academic text becomes readable text, without losing meaning[cite: 1]. The design system makes that transformation itself the visual language — not editorial print styling, not a generic reading-app template[cite: 1]. Every structural device should reinforce "verified clarity," because the product's actual mechanism (claim-diff verification) is a trust feature, not just plumbing[cite: 1].

---

## Colors

### Base & Tokens
*   **--surface-base** (`#FAFAF9`): Page background (cool off-white, not cream)[cite: 1].
*   **--surface-raised** (`#FFFFFF`): Cards, dropzone, chart cards[cite: 1].
*   **--surface-inverse** (`#14171F`): Dark-mode base[cite: 1].
*   **--ink-primary** (`#1B1F27`): Primary text[cite: 1].
*   **--ink-secondary** (`#5B6472`): Secondary/muted text, captions[cite: 1].
*   **--border-default** (`#E4E7EC`): Default borders, dividers[cite: 1].

### Accent & States
*   **--accent-verified** (`#2F6F5E`): Primary accent — verification success, links, primary actions[cite: 1].
*   **--accent-verified-soft** (`#E7F2EE`): Accent background tint (badges, subtle highlight)[cite: 1].
*   **--state-warning** (`#B5710B`): verification_failed banner — amber, distinct from error[cite: 1].
*   **--state-error** (`#C4453B`): Upload/processing failures[cite: 1].
*   **--state-warning-soft** (`#FBF2E3`): Warning banner background[cite: 1].
*   **--state-error-soft** (`#FCEEEC`): Error banner background[cite: 1].

> **Dark Mode Tokens:** `--surface-base` (#14171F), `--surface-raised` (#1C202A), `--ink-primary` (#F0F1F3), `--ink-secondary` (#9AA3B1), `--border-default` (#2A2F3A)[cite: 1]. Accents shift brighter: `--accent-verified` (#4A9684), `--state-warning` (#D89440), `--state-error` (#E37168)[cite: 1].

---

## Typography

*   **UI Face**: Inter — neutral, modern, legible at small sizes for chrome[cite: 1].
*   **Reading Face**: Source Serif 4 — used only for the body of simplified/original paper text[cite: 1].
*   **Mono Face**: JetBrains Mono — used only for extracted numeric/tabular data[cite: 1].

### Typography Scale
*   **text-hero**: Inter 40px, 600 weight, 1.15 line height (Upload page headline only)[cite: 1].
*   **text-h2**: Inter 24px, 600 weight, 1.25 line height (Section headers)[cite: 1].
*   **text-h3**: Inter 18px, 600 weight, 1.3 line height (Chart card titles, component labels)[cite: 1].
*   **text-reading**: Source Serif 4, 18px, 400 weight, 1.7 line height (Simplified/original body text)[cite: 1].
*   **text-body**: Inter 15px, 400 weight, 1.5 line height (UI copy, buttons, form labels)[cite: 1].
*   **text-caption**: Inter 13px, 500 weight, 1.4 line height (Metadata, timestamps, chip labels)[cite: 1].
*   **text-mono**: JetBrains Mono 13px, 400 weight, 1.5 line height (Extracted data values)[cite: 1].

---

## Spacing

Base unit: **8px** (utility reading tool, not a specimen sheet)[cite: 1].
*   **space-1**: 4px — Icon-to-label, tight inline gaps[cite: 1].
*   **space-2**: 8px — Small gaps[cite: 1].
*   **space-3**: 16px — Within component groups[cite: 1].
*   **space-4**: 24px — Card inner padding[cite: 1].
*   **space-5**: 32px — Between components[cite: 1].
*   **space-6**: 48px — Section padding[cite: 1].
*   **space-8**: 64px — Between major page sections[cite: 1].

---

## Border Radius & Elevation

Modern touch executed with restraint, avoiding sharp print edges or heavy skeuomorphism[cite: 1].
*   **radius-sm** (6px): Chips, badges, inline controls[cite: 1].
*   **radius-md** (10px): Buttons, inputs[cite: 1].
*   **radius-lg** (16px): Cards, dropzone, banners[cite: 1].
*   **radius-full** (9999px): Status pills[cite: 1].

### Elevation
*   **shadow-card**: `0 1px 2px rgba(20,23,31,0.04), 0 4px 12px rgba(20,23,31,0.06)` — applied to raised surfaces only[cite: 1]. No shadow on flat UI chrome[cite: 1].
*   **Focus States**: MUST use a 2px `--accent-verified` ring with 2px offset[cite: 1].

---

## Components

### Buttons
*   **Primary**: `--accent-verified` fill, white text, `radius-md`, shadow-card on hover only[cite: 1].
*   **Secondary**: transparent fill, `--ink-primary` text, 1px `--border-default` border, `radius-md`[cite: 1].
*   **Disabled**: 0.4 opacity, no hover transform[cite: 1].

### Cards & Dropzone
*   **Upload Dropzone**: `radius-lg`, `shadow-card`, dashed `--border-default` border when idle, solid `--accent-verified` border on drag-hover[cite: 1].
*   **Chart Card**: `radius-lg`, `shadow-card`, `--surface-raised` fill, `space-4` padding[cite: 1].

### Inputs
*   `radius-md`, 1px `--border-default` border, `--surface-raised` fill[cite: 1]. Focus: `--accent-verified` border + ring[cite: 1]. Error: `--state-error` border, `--state-error-soft` fill[cite: 1].

### Specialty UI Elements
*   **Reading Level Selector**: `shadcn ToggleGroup` with `radius-md` segments, `--accent-verified` fill on selected segment[cite: 1].
*   **Verification Badge**: Small pill, `radius-full`, `--accent-verified-soft` background, `--accent-verified` text/icon[cite: 1].

---

## Do's and Don'ts

1. **Do** prioritize readability over visual decoration — this is a reading tool, not a showcase[cite: 1].
2. **Don't** default to warm cream + terracotta editorial styling — that palette belongs to print/portfolio work[cite: 1].
3. **Do** use `--accent-verified` exclusively for verified/trustworthy signals[cite: 1].
4. **Don't** show a fake or placeholder verification badge before claim-diff actually passes[cite: 1].
5. **Do** maintain a reading column width max limit of 70 characters (`~max-w-prose`)[cite: 1].
6. **Don't** use more than two display typefaces (Inter + Source Serif 4; JetBrains Mono is data-only)[cite: 1].
7. **Do** ensure touch targets are a minimum of 44×44px for accessibility[cite: 1].