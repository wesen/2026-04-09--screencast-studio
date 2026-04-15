---
Title: 02 capture pprof cpu profile summary
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
Summary: Saved CPU profile capture from the opt-in pprof server.
LastUpdated: 2026-04-14T20:48:26-04:00
WhatFor: Preserve one pprof CPU profile and its top summaries for later comparison.
WhenToUse: Read when comparing repeated pprof captures across the same high-signal repro.
---

# 02 capture pprof cpu profile summary

- pprof_url: http://127.0.0.1:6060/debug/pprof/profile
- seconds: 41
- run_dir: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/02-capture-pprof-cpu-profile/20260414-204800

## Files

- cpu.pprof
- pprof-top.txt
- pprof-top-cum.txt
