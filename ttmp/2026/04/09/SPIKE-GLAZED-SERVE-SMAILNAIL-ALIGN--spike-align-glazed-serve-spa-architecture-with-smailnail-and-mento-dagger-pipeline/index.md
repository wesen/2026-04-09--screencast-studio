---
Title: 'Spike: Align Glazed Serve SPA Architecture with Smailnail and Mento Dagger Pipeline'
Ticket: SPIKE-GLAZED-SERVE-SMAILNAIL-ALIGN
Status: active
Topics:
    - architecture
    - frontend
    - backend
    - ui
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/code/mento/go-go-mento/docs/infra/dagger-build-pipeline.md
      Note: Canonical Mento Dagger pipeline documentation
    - Path: /home/manuel/code/wesen/corporate-headquarters/glazed/cmd/build-web/main.go
      Note: Glazed Dagger builder - main reference for SPA build
    - Path: /home/manuel/code/wesen/corporate-headquarters/glazed/pkg/web/static.go
      Note: Glazed SPA handler - //go:embed and SPA fallback
    - Path: /home/manuel/code/wesen/corporate-headquarters/smailnail/cmd/build-web/main.go
      Note: Smailnail Dagger builder - alternative implementation with CacheVolume
    - Path: /home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/smailnaild/web/embed.go
      Note: Smailnail embed - //go:embed embed/public pattern
    - Path: /home/manuel/code/wesen/corporate-headquarters/smailnail/pkg/smailnaild/web/spa.go
      Note: Smailnail RegisterSPA - API prefix handling pattern
ExternalSources: []
Summary: Spike workspace for aligning the screencast-studio web build and serve architecture with the Dagger-backed SPA pipeline used in Smailnail and discussed in Mento.
LastUpdated: 2026-04-09T22:31:00-04:00
WhatFor: Track the alignment work needed to add a reproducible frontend build pipeline and a disk-or-embed SPA serving path to this repo.
WhenToUse: Read when working on build-web, go:generate frontend builds, embedded SPA serving, or related Dagger pipeline work.
---


# Spike: Align Glazed Serve SPA Architecture with Smailnail and Mento Dagger Pipeline

Document workspace for SPIKE-GLAZED-SERVE-SMAILNAIL-ALIGN.
