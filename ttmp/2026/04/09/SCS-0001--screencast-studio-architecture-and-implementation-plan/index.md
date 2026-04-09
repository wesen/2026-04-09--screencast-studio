---
Title: Screencast Studio Architecture and Implementation Plan
Ticket: SCS-0001
Status: active
Topics:
    - backend
    - frontend
    - video
    - audio
    - dsl
    - cli
DocType: index
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources:
    - local:screencast-studio-v2.jsx.jsx
Summary: Ticket landing page for the screencast studio architecture, implementation guide, and delivery artifacts.
LastUpdated: 2026-04-09T13:12:49.209385706-04:00
WhatFor: Detailed architecture, implementation guide, and ticket landing page for a Go-based screencast studio with a streaming control frontend.
WhenToUse: Read this first when starting implementation, reviewing scope, or locating the primary design and diary documents.
---


# Screencast Studio Architecture and Implementation Plan

## Overview

This ticket defines a new screencast studio application with a Go backend, Glazed-powered CLI flags, an intermediate setup DSL, and a React control frontend that streams live source previews. The goal is to replace the current jank prototype with a structured system that can discover capture sources, preview them in the browser, and record separate output files according to user-defined templates.

This workspace is written as onboarding material for a new intern. It explains the current prototype, the desired target architecture, the recommended package layout, the HTTP and WebSocket contracts, the DSL and compiled-plan layers, and the phased implementation plan.

## Key Links

- Primary design doc: `design-doc/01-screencast-studio-system-design.md`
- Diary: `reference/01-diary.md`
- Tasks: `tasks.md`
- Imported UI reference: `sources/local/screencast-studio-v2.jsx.jsx`

## Status

Current status: **active**

## Topics

- backend
- frontend
- video
- audio
- dsl
- cli

## Tasks

See `tasks.md` for the current implementation checklist.

## Changelog

See `changelog.md` for recent changes and decisions.

## Structure

- design-doc/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- sources/ - Imported local artifacts and research inputs
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
