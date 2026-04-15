---
Title: 03 check profiler prereqs summary
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
Summary: Checks local availability and permissions for pprof-adjacent lower-level profiling tools.
LastUpdated: 2026-04-14T21:13:02-04:00
WhatFor: Preserve profiler availability and permission checks before attempting perf or eBPF captures.
WhenToUse: Read before running perf or bpftrace scripts on this machine.
---

# 03 check profiler prereqs summary

- listen_pid: 716961
- perf_path: /usr/bin/perf
- bpftrace_path: /usr/bin/bpftrace
- bpftool_path: /usr/sbin/bpftool
- perf_event_paranoid: 1

## Files

- context.txt
- perf-check.txt
- bpftrace-check.txt
- bpftool-check.txt
