Yes.

On **X11**, you can do both:

1. **Capture one specific window by XID**
2. **Enumerate top-level windows and compute their screen rectangles**

For capture, both of the tools you were already considering support per-window input:

* **FFmpeg `x11grab`** has a `window_id` option. Its docs say this grabs that window instead of the root window, and the window ID can be found with `xwininfo`. It also notes two important caveats: if the window later grows, the new area is not recorded, and capture ends if the window is closed, unmapped/iconified, or shrunk beyond the capture size. ([FFmpeg][1])
* **GStreamer `ximagesrc`** exposes an `xid` property for the window to capture, and also an `xname` property. The GStreamer docs snippet explicitly describes `xid` as “The XID of the window to capture. 0 for the root window.” ([gstreamer.freedesktop.org][2])

A minimal FFmpeg example is:

```bash
ffmpeg -f x11grab -window_id 0x3a00007 -framerate 60 \
  -c:v libx264 -preset veryfast -crf 20 out.mkv
```

And the GStreamer equivalent is conceptually:

```text
ximagesrc xid=0x3a00007 ! videoconvert ! x264enc ! matroskamux ! filesink location=out.mkv
```

For **listing windows and their areas**, the standard X11/EWMH route is:

* Read the root window’s **`_NET_CLIENT_LIST`** or **`_NET_CLIENT_LIST_STACKING`** property. The EWMH spec defines those as the arrays of all windows managed by the window manager. ([specifications.freedesktop.org][3])
* For each window, read **`_NET_WM_NAME`** for a UTF-8 title if present. ([specifications.freedesktop.org][3])
* Get geometry via **`XGetWindowAttributes`**. Its `x`/`y` are relative to the **parent** window, while `width`/`height` are the inside size and do **not** include the border. ([X.Org][4])
* Convert coordinates to root/screen space with **`XTranslateCoordinates`**, which returns coordinates in the destination window’s coordinate system. ([X.Org][5])
* If you care about decorations, read **`_NET_FRAME_EXTENTS`**. The spec defines it as left/right/top/bottom widths added by the window manager. ([specifications.freedesktop.org][3])

So the effective algorithm is:

```text
root
 -> _NET_CLIENT_LIST
 -> for each XID:
      title = _NET_WM_NAME or WM_NAME
      attrs = XGetWindowAttributes(xid)
      rootX, rootY = XTranslateCoordinates(parent-relative x/y -> root)
      frame = _NET_FRAME_EXTENTS (optional)
      rect = {x: rootX, y: rootY, w: attrs.width, h: attrs.height}
      frameRect = rect expanded by extents
```

That gives you a usable table like:

```text
XID        Title              Client Rect             Frame Rect
0x3a00007  "Firefox"          1920,24 1280x1016       1918,0 1284x1040
0x4200012  "Alacritty"        0,24 960x516            0,24 960x516
```

Two implementation details matter:

* If your recorder should follow moves/resizes live, watch **`ConfigureNotify`** and re-read geometry when it changes.
* Decide whether you want the **client area** or the **framed window**. For “record exactly the app contents,” use the client rect. For “record what the user visually sees including titlebar/borders,” add `_NET_FRAME_EXTENTS`. ([specifications.freedesktop.org][3])

For your “minimal OBS in Go” case, the clean split is:

* **Enumeration/selection:** use X11/XCB from Go to build the window list and rectangles.
* **Capture:** pass the chosen XID to FFmpeg `window_id`, or to GStreamer `ximagesrc xid=...`. ([FFmpeg][1])

If you want, next step should be a small **Go/X11 window enumerator** that prints JSON like:

```json
[
  {"id":"0x3a00007","title":"Firefox","x":1920,"y":24,"w":1280,"h":1016},
  {"id":"0x4200012","title":"Alacritty","x":0,"y":24,"w":960,"h":516}
]
```

and then feeds the selected `id` into FFmpeg or GStreamer.

[1]: https://ffmpeg.org/ffmpeg-devices.html "      FFmpeg Devices Documentation
"
[2]: https://gstreamer.freedesktop.org/documentation/ximagesrc/index.html?utm_source=chatgpt.com "ximagesrc"
[3]: https://specifications.freedesktop.org/wm/latest-single "Extended Window Manager Hints"
[4]: https://www.x.org/releases/current/doc/man/man3/XGetWindowAttributes.3.xhtml?utm_source=chatgpt.com "XGetWindowAttributes"
[5]: https://www.x.org/releases/current/doc/man/man3/XTranslateCoordinates.3.xhtml?utm_source=chatgpt.com "XTranslateCoordinates"


---
