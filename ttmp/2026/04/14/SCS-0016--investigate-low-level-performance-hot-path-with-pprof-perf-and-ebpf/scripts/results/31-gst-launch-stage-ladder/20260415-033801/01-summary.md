# 31 gst-launch stage ladder

- display: :0
- root: 2880x1920+0+0
- stage: capture
- fps: 24
- bitrate: 6920
- container: mov
- duration_seconds: 6
- avg_cpu: 1.00%
- max_cpu: 2.00%
- output: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/31-gst-launch-stage-ladder/20260415-033801/output.mov
- pipeline: ximagesrc display-name=:0 use-damage=false show-pointer=true ! fakesink sync=false
