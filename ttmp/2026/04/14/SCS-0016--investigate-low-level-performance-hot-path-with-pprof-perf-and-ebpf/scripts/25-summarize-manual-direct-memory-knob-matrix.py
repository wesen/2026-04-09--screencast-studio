#!/usr/bin/env python3
"""Summarize a saved 23-manual-direct-memory-knob-matrix result directory.

This reproduces the quick analysis used in chat:
- corrected avg/max CPU from pidstat CPU rows only
- page/minor/major fault totals from perf-stat.csv
- last-sample smaps_rollup summary including AnonHugePages
- global THP context snapshot

Usage:
  25-summarize-manual-direct-memory-knob-matrix.py <result-dir>
"""

from __future__ import annotations

import sys
from pathlib import Path

SCENARIOS = ["baseline", "thp_disable_prctl", "malloc_arena_max_1"]


def parse_pidstat_cpu(pidstat_path: Path, duration_seconds: int = 8) -> tuple[str, str]:
    vals: list[float] = []
    for line in pidstat_path.read_text(errors="replace").splitlines():
        parts = line.split()
        if len(parts) == 11 and parts[2].isdigit() and parts[10].startswith("manual-direct"):
            vals.append(float(parts[8]))
    if not vals:
        return "?", "?"
    use = vals[:duration_seconds] if len(vals) >= duration_seconds else vals
    avg_cpu = sum(use) / len(use)
    max_cpu = max(use)
    return f"{avg_cpu:.2f}%", f"{max_cpu:.2f}%"


def parse_perf_stat(perf_path: Path) -> dict[str, str]:
    out: dict[str, str] = {}
    for line in perf_path.read_text(errors="replace").splitlines():
        if not line or line.startswith("#"):
            continue
        parts = line.split(",")
        if len(parts) < 3:
            continue
        value = parts[0].strip()
        event = parts[2].strip()
        if event and event not in out:
            out[event] = value
    return out


def last_smaps_summary(proc_samples_dir: Path) -> dict[str, str]:
    samples = sorted([p for p in proc_samples_dir.iterdir() if p.is_dir()])
    if not samples:
        return {}
    txt = (samples[-1] / "smaps_rollup.txt").read_text(errors="replace").splitlines()
    keys = ["Rss:", "Pss_Anon:", "Private_Dirty:", "Anonymous:", "AnonHugePages:", "Swap:"]
    out: dict[str, str] = {"sample": samples[-1].name}
    for line in txt:
        for key in keys:
            if line.startswith(key):
                out[key.rstrip(":")] = line.split(":", 1)[1].strip()
    return out


def main(argv: list[str]) -> int:
    if len(argv) != 2:
        print(__doc__.strip(), file=sys.stderr)
        return 2
    run_dir = Path(argv[1])
    print("# manual direct memory knob matrix summary")
    print()
    for scenario in SCENARIOS:
        sdir = run_dir / scenario
        print(f"## {scenario}")
        avg_cpu, max_cpu = parse_pidstat_cpu(sdir / "pidstat.log")
        perf = parse_perf_stat(sdir / "perf-stat.csv")
        smaps = last_smaps_summary(sdir / "proc-samples")
        print(f"- avg_cpu: {avg_cpu}")
        print(f"- max_cpu: {max_cpu}")
        print(f"- page-faults: {perf.get('page-faults', '?')}")
        print(f"- minor-faults: {perf.get('minor-faults', '?')}")
        print(f"- major-faults: {perf.get('major-faults', '?')}")
        if smaps:
            print(f"- last_proc_sample: {smaps.get('sample', '?')}")
            print(f"- AnonHugePages: {smaps.get('AnonHugePages', '?')}")
            print(f"- Anonymous: {smaps.get('Anonymous', '?')}")
            print(f"- Private_Dirty: {smaps.get('Private_Dirty', '?')}")
        print()
    print("## global THP context")
    thp = (run_dir / "baseline" / "thp-context-before.txt").read_text(errors="replace")
    print(thp.strip())
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
