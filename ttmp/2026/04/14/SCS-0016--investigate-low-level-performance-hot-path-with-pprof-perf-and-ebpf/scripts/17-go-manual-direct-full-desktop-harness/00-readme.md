---
Title: 17 go manual direct full desktop harness
Ticket: SCS-0016
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - backend
    - analysis
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Ticket-local script/readme artifact for SCS-0016.
LastUpdated: 2026-04-15T01:45:00-04:00
WhatFor: Preserve this local script/readme artifact as part of the reproducible SCS-0016 investigation surface.
WhenToUse: Read when rerunning or reviewing this specific script directory inside SCS-0016.
---

# 17 go manual direct full desktop harness

Manual Go-constructed full-desktop direct recording graph with no DSL structs and no app helper reuse.

Graph:

`ximagesrc -> videoconvert -> videorate -> capsfilter(I420,framerate) -> x264enc -> h264parse -> qtmux/mp4mux -> filesink`
