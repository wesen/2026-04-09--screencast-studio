---
Title: 04 capture perf cpu profile summary
Ticket: SCS-0016
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - backend
    - analysis
    - perf
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Saved mixed-stack perf profile capture from the live server PID while the browser high-signal repro is active.
LastUpdated: 2026-04-14T22:50:25-04:00
WhatFor: Preserve one perf capture and its text reports for comparison against pprof and later reruns.
WhenToUse: Read when comparing mixed-stack perf captures across the same high-signal repro.
---

# 04 capture perf cpu profile summary

- label: desktop-preview-recording-one-browser-tab
- server_url: http://127.0.0.1:7777
- server_pid: 716961
- perf_seconds: 20
- perf_freq: 99
- call_graph: dwarf,16384
- perf_data: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/04-capture-perf-cpu-profile/20260414-224952/perf.data

## Files

- context.txt
- healthz.json
- perf-version.txt
- perf.data
- perf-record.stdout.log
- perf-record.stderr.log
- perf-report.txt
- perf-report-dso-symbol.txt
- perf-script.txt
- go-top-addresses.txt
- go-addr2line.txt
