---
Title: 'Spike: Align Glazed Serve SPA Architecture with Smailnail and Mento Dagger Pipeline'
Ticket: SPIKE-GLAZED-SERVE-SMAILNAIL-ALIGN
Status: active
Topics:
    - architecture
    - spike
    - web-frontend
    - go-embedding
    - dagger
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../mento/go-go-mento/docs/infra/dagger-build-pipeline.md
      Note: Canonical Mento Dagger pipeline documentation
    - Path: and SPA.md
      Note: Glazed Serve project specification and history
    - Path: glazed/cmd/build-web/main.go
      Note: Glazed Dagger builder - main reference for SPA build
    - Path: glazed/pkg/web/static.go
      Note: Glazed SPA handler - //go:embed and SPA fallback
    - Path: smailnail/cmd/build-web/main.go
      Note: Smailnail Dagger builder - alternative implementation with CacheVolume
    - Path: smailnail/pkg/smailnaild/web/embed.go
      Note: Smailnail embed - //go:embed embed/public pattern
    - Path: smailnail/pkg/smailnaild/web/spa.go
      Note: Smailnail RegisterSPA - API prefix handling pattern
ExternalSources: []
Summary: ""
LastUpdated: 2026-04-09T20:37:11.004285143-04:00
WhatFor: ""
WhenToUse: ""
---


# Spike: Align Glazed Serve SPA Architecture with Smailnail and Mento Dagger Pipeline

Document workspace for SPIKE-GLAZED-SERVE-SMAILNAIL-ALIGN.
