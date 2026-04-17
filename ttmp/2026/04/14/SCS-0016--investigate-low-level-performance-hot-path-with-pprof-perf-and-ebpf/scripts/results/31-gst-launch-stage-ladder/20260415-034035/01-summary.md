# 31 gst-launch stage ladder

- display: :0
- root: 2880x1920+0+0
- stage: mux-file
- fps: 24
- bitrate: 6920
- container: mov
- duration_seconds: 6
- avg_cpu: 131.11%
- max_cpu: 146.00%
- output: /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/31-gst-launch-stage-ladder/20260415-034035/output.mov
- pipeline: ximagesrc display-name=:0 use-damage=false show-pointer=true ! videoconvert ! videorate ! video/x-raw,format=I420,framerate=24/1,pixel-aspect-ratio=1/1 ! x264enc bitrate=6920 bframes=0 tune=zerolatency speed-preset=3 ! h264parse ! qtmux ! filesink location=/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/14/SCS-0016--investigate-low-level-performance-hot-path-with-pprof-perf-and-ebpf/scripts/results/31-gst-launch-stage-ladder/20260415-034035/output.mov

## ffprobe

~~~text
codec_name=h264
width=2880
height=1920
avg_frame_rate=24/1
duration=13.958333
size=6014945
~~~
