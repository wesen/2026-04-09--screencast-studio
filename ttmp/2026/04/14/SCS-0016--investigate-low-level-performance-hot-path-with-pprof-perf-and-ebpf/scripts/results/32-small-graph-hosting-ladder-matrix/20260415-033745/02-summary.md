# 32 small graph hosting ladder matrix

| stage | go avg_cpu | python avg_cpu | gst-launch avg_cpu | go page-faults | python page-faults | gst-launch page-faults |
|---|---:|---:|---:|---:|---:|---:|
| capture | 1.00 | 2.16 | 1.00 | 28 | 61 | 0 |
| convert | 1.00 | 2.17 | 1.00 | 28 | 59 | 0 |
| rate-caps | 37.83 | 36.83 | 36.67 | 34 | 51 | 0 |
| encode | 199.97 | 125.17 | 129.00 | 312438 | 62 | 29 |
| parse | 170.50 | 126.50 | 128.44 | 275291 | 163 | 4 |
| mux-file | 197.17 | 131.17 | 131.11 | 288738 | 209 | 233 |
