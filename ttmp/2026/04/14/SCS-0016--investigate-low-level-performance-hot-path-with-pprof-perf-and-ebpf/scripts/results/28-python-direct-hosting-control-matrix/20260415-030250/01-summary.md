# 28 python direct hosting control matrix

- parse_launch_run: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/26-python-parse-launch-direct-full-desktop-harness/results/20260415-030250
- manual_run: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/27-python-manual-direct-full-desktop-harness/results/20260415-030259

| scenario | avg_cpu | max_cpu | duration |
|---|---:|---:|---:|
| python-parse-launch | 145.50% | 158.00% | 8.000000 |
| python-manual | 153.75% | 215.00% | 8.000000 |

## Interpretation

- Both Python-hosted controls are far closer to the matched gst-launch control than to the hotter Go-hosted controls.
- The Python manual graph is only slightly hotter than the Python parse_launch control.
- This is evidence against a generic “embedded application is always much hotter than gst-launch” theory and makes the remaining gap look more Go-hosting-specific.
