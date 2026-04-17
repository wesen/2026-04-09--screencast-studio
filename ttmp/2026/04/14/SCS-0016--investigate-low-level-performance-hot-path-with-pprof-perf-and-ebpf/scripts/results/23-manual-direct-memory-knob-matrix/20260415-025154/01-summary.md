# 23 manual direct memory knob matrix

| scenario | avg_cpu | max_cpu | page-faults | minor-faults | major-faults | attached_tids |
|---|---:|---:|---:|---:|---:|---:|
| baseline | 161.38% | 278.00% | 296940 | 296939 | 0 | 24 |
| thp_disable_prctl | 151.50% | 216.00% | 317743 | 317743 | 0 | 23 |
| malloc_arena_max_1 | 212.25% | 391.00% | 290263 | 290246 | 0 | 24 |

## Notes

- Global THP mode on this machine was already `madvise` during the run.
- `thp_disable_prctl` used a per-process `prctl(PR_SET_THP_DISABLE)` helper as a non-root proxy for testing a stronger THP-off posture.
- All observed faults in these three runs were effectively minor faults; major faults stayed at zero.
