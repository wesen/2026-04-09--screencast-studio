#!/usr/bin/env python3
import ctypes
import os
import sys

PR_SET_THP_DISABLE = 41
libc = ctypes.CDLL(None, use_errno=True)
libc.prctl.argtypes = [ctypes.c_int, ctypes.c_ulong, ctypes.c_ulong, ctypes.c_ulong, ctypes.c_ulong]
libc.prctl.restype = ctypes.c_int


def main() -> int:
    if len(sys.argv) < 2:
        print("usage: 22-prctl-disable-thp-and-exec.py <program> [args...]", file=sys.stderr)
        return 2
    rc = libc.prctl(PR_SET_THP_DISABLE, 1, 0, 0, 0)
    if rc != 0:
        err = ctypes.get_errno()
        print(f"prctl(PR_SET_THP_DISABLE) failed: errno={err} ({os.strerror(err)})", file=sys.stderr)
        return 1
    os.execvpe(sys.argv[1], sys.argv[1:], os.environ.copy())
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
