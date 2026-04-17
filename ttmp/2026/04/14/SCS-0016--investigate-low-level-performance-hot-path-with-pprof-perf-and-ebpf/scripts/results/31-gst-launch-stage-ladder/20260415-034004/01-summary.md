# 31 gst-launch stage ladder

- display: :0
- root: 2880x1920+0+0
- stage: parse
- fps: 24
- bitrate: 6920
- container: mov
- duration_seconds: 6
- avg_cpu: 128.44%
- max_cpu: 134.00%
- output: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/31-gst-launch-stage-ladder/20260415-034004/output.mov
- pipeline: ximagesrc display-name=:0 use-damage=false show-pointer=true ! videoconvert ! videorate ! video/x-raw,format=I420,framerate=24/1,pixel-aspect-ratio=1/1 ! x264enc bitrate=6920 bframes=0 tune=zerolatency speed-preset=3 ! h264parse ! fakesink sync=false
