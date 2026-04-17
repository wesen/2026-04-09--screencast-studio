#!/usr/bin/env python3
"""Compare two GStreamer .dot graph dumps after normalizing unstable addresses/paths.

Usage:
  24-compare-dot-graphs-normalized.py <dot-a> <dot-b>

Prints either IDENTICAL_AFTER_NORMALIZATION or DIFFERS_AFTER_NORMALIZATION,
and emits a small unified diff when they differ.
"""

from __future__ import annotations

import difflib
import re
import sys
from pathlib import Path


def norm(text: str) -> str:
    text = re.sub(r"0x[0-9a-f]+", "0xADDR", text)
    text = re.sub(r"/tmp/qtmux[^\"\\]+", "/tmp/qtmuxTMP", text)
    text = re.sub(r'location="[^"]+"', 'location="OUTPUT"', text)
    text = re.sub(r"0\.00\.[0-9.]+-", "TS-", text)
    return text


def main(argv: list[str]) -> int:
    if len(argv) != 3:
        print(__doc__.strip(), file=sys.stderr)
        return 2
    a = Path(argv[1])
    b = Path(argv[2])
    ta = norm(a.read_text())
    tb = norm(b.read_text())
    if ta == tb:
        print("IDENTICAL_AFTER_NORMALIZATION")
        return 0
    print("DIFFERS_AFTER_NORMALIZATION")
    diff = difflib.unified_diff(
        ta.splitlines(),
        tb.splitlines(),
        fromfile=str(a),
        tofile=str(b),
        n=2,
    )
    for i, line in enumerate(diff):
        if i >= 200:
            break
        print(line)
    return 1


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
