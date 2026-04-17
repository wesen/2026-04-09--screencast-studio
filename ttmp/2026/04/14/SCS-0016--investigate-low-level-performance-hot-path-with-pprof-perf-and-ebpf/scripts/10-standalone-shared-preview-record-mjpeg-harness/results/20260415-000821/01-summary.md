---
Title: 10 standalone shared preview record mjpeg harness summary
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
Summary: Standalone harness run for shared source + preview to MJPEG HTTP + recording without reusing the web server.
LastUpdated: 2026-04-15T00:08:34.542128946-04:00
WhatFor: Preserve standalone near-web harness evidence for comparing native/shared path cost against the full web server path.
WhenToUse: Read when comparing this standalone harness against server/browser-backed captures.
---

# 10 standalone shared preview record mjpeg harness summary

- exit_status: 0
- source_type: display
- http_addr: 127.0.0.1:7791
- mjpeg_url: http://127.0.0.1:7791/mjpeg
- output_path: /tmp/scs-standalone-shared-preview-record.mp4
- client_enabled: True
- warmup_seconds: 2
- record_seconds: 8
- avg_cpu: 187.36%
- max_cpu: 508.00%
- preview_frames_seen: 62
- preview_bytes_seen: 12542731
- mjpeg_streams: 1
- mjpeg_frames_served: 61
- mjpeg_bytes_served: 12244271
- recording_state: finished
- recording_reason: recording stop requested
- error: None

## Files

- harness.stdout.json
- harness.stderr.log
- harness.pidstat.log
- output.mp4 (or configured output path)
