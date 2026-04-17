---
Title: 19 perf compare go manual vs gst-launch
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
Summary: Ticket-local result summary artifact for SCS-0016.
LastUpdated: 2026-04-15T01:45:00-04:00
WhatFor: Preserve this local result summary so the SCS-0016 evidence trail stays reproducible and reviewable.
WhenToUse: Read when reviewing or comparing this specific saved result inside SCS-0016.
---

# 19 perf compare go manual vs gst-launch

| scenario | codec | width | height | fps | duration |
|---|---|---:|---:|---|---:|
| go-manual | h264 | 2880 | 1920 | 24/1 | 8.041667 |
| gst-launch | h264 | 2880 | 1920 | 24/1 | 7.958333 |

## go-manual top mixed-stack entries

~~~text
42.48%    38.23%  libx264.so.164                [.] x264_8_trellis_coefn
25.35%     0.00%  [unknown]                     [k] 0xffffffffffffffff
24.31%     0.00%  libc.so.6                     [.] clone3
13.20%     0.00%  libglib-2.0.so.0.8000.0       [.] 0x000072c32c6cae61
13.20%     0.00%  libglib-2.0.so.0.8000.0       [.] 0x000072c32c6d0421
12.91%     0.00%  libgstreamer-1.0.so.0.2402.0  [.] gst_pad_push
11.70%     0.00%  [unknown]                     [.] 0x000072c2de848584
11.70%     0.00%  [unknown]                     [.] 0x000072c2de8516f3
9.77%      0.00%  [unknown]                     [.] 0x000072c2de84977c
8.26%      0.00%  [unknown]                     [.] 0x000072c2de8474a9
8.24%      0.00%  [unknown]                     [.] 0x000072c2de8474d9
7.18%      0.00%  [unknown]                     [.] 0x000072c2de8487d6
~~~

## gst-launch top mixed-stack entries

~~~text
42.09%    41.44%  libx264.so.164                [.] x264_8_trellis_coefn
34.93%     0.00%  libc.so.6                     [.] clone3
19.49%     0.00%  libglib-2.0.so.0.8000.0       [.] 0x000070424c7aee61
19.49%     0.00%  libglib-2.0.so.0.8000.0       [.] 0x000070424c7b4421
19.29%     0.00%  libgstreamer-1.0.so.0.2402.0  [.] gst_pad_push
10.75%     0.00%  [unknown]                     [.] 0x000070424b7c0f5a
10.25%     0.00%  [unknown]                     [.] 0x000070424b7bc995
9.97%      0.00%  [unknown]                     [k] 0xffffffffffffffff
9.90%      0.00%  [unknown]                     [.] 0x000070424b7bd539
9.86%      0.00%  [unknown]                     [.] 0xffffffffffffffff
8.74%      0.00%  [unknown]                     [.] 0x000070424c931bb3
8.31%      0.00%  [unknown]                     [.] 0x000070424b700ec6
~~~

## Interpretation

- Both profiles are dominated by native encoder and GStreamer push-path work, not by ordinary Go userland logic.
- In the Go-hosted manual harness, directly visible Go/cgo frames are tiny in the dso-sorted report (for example `runtime.asmcgocall.abi0` around `0.37%`, `_cgo_*gst_element_set_state` around `0.24%`).
- This supports the model that, once the direct pipeline is running, video frames are not flowing through Go code on every frame.
- The remaining Go-vs-`gst-launch` gap therefore looks more like a process/runtime-hosting difference than a graph-construction or per-frame-Go-copy issue.
