#!/usr/bin/env python3
import argparse
import json
import os
import time
from pathlib import Path

import gi

gi.require_version('Gst', '1.0')
from gi.repository import Gst  # noqa: E402


def dump_graph(pipeline, name: str) -> None:
    if not name:
        return
    Gst.debug_bin_to_dot_file_with_ts(pipeline, Gst.DebugGraphDetails.ALL, name)


def make(name: str):
    elem = Gst.ElementFactory.make(name)
    if elem is None:
        raise RuntimeError(f'failed to create element {name}')
    return elem


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument('--display-name', default=os.environ.get('DISPLAY', ':0').strip() or ':0')
    parser.add_argument('--fps', type=int, default=24)
    parser.add_argument('--bitrate', type=int, default=6920)
    parser.add_argument('--container', default='mov')
    parser.add_argument('--duration-seconds', type=int, default=8)
    parser.add_argument('--output-path', default=str(Path('/tmp') / 'scs-python-manual-direct.mov'))
    parser.add_argument('--dot-dir', default='')
    args = parser.parse_args()

    started_at = time.time()
    summary = {
        'started_at': time.strftime('%Y-%m-%dT%H:%M:%S%z', time.localtime(started_at)),
        'display_name': args.display_name,
        'fps': args.fps,
        'bitrate': args.bitrate,
        'container': args.container,
        'output_path': args.output_path,
        'duration_seconds': args.duration_seconds,
    }

    try:
        Path(args.output_path).parent.mkdir(parents=True, exist_ok=True)
        if args.dot_dir:
            Path(args.dot_dir).mkdir(parents=True, exist_ok=True)
            os.environ['GST_DEBUG_DUMP_DOT_DIR'] = args.dot_dir

        Gst.init(None)

        pipeline = Gst.Pipeline.new(None)
        if pipeline is None:
            raise RuntimeError('failed to create pipeline')

        ximagesrc = make('ximagesrc')
        ximagesrc.set_property('display-name', args.display_name)
        ximagesrc.set_property('show-pointer', True)
        ximagesrc.set_property('use-damage', False)

        videoconvert = make('videoconvert')
        videorate = make('videorate')
        capsfilter = make('capsfilter')
        capsfilter.set_property('caps', Gst.Caps.from_string(f'video/x-raw,format=I420,framerate={args.fps}/1,pixel-aspect-ratio=1/1'))

        x264enc = make('x264enc')
        x264enc.set_property('bitrate', args.bitrate)
        x264enc.set_property('bframes', 0)
        x264enc.set_property('tune', 4)
        x264enc.set_property('speed-preset', 3)

        h264parse = make('h264parse')
        mux_name = 'qtmux' if args.container.lower() in ('', 'mov', 'qt') else 'mp4mux'
        if args.container.lower() not in ('', 'mov', 'qt', 'mp4'):
            raise RuntimeError(f'unsupported container {args.container!r}')
        mux = make(mux_name)
        filesink = make('filesink')
        filesink.set_property('location', args.output_path)

        elems = [ximagesrc, videoconvert, videorate, capsfilter, x264enc, h264parse, mux, filesink]
        for elem in elems:
            pipeline.add(elem)
        for a, b in zip(elems, elems[1:]):
            if not a.link(b):
                raise RuntimeError(f'failed to link {a.get_name()} -> {b.get_name()}')

        bus = pipeline.get_bus()
        if bus is None:
            raise RuntimeError('pipeline bus is nil')

        dump_graph(pipeline, 'python-manual-direct-pre-play')
        ret = pipeline.set_state(Gst.State.PLAYING)
        if ret == Gst.StateChangeReturn.FAILURE:
            raise RuntimeError('failed to set pipeline to PLAYING')
        time.sleep(1)
        dump_graph(pipeline, 'python-manual-direct-playing')

        if args.duration_seconds > 1:
            time.sleep(args.duration_seconds - 1)
        pipeline.send_event(Gst.Event.new_eos())

        deadline = time.time() + 20
        while time.time() < deadline:
            msg = bus.timed_pop_filtered(500 * Gst.MSECOND, Gst.MessageType.EOS | Gst.MessageType.ERROR)
            if msg is None:
                continue
            if msg.type == Gst.MessageType.EOS:
                dump_graph(pipeline, 'python-manual-direct-eos')
                summary['result'] = 'eos'
                pipeline.set_state(Gst.State.NULL)
                break
            if msg.type == Gst.MessageType.ERROR:
                dump_graph(pipeline, 'python-manual-direct-error')
                err, debug = msg.parse_error()
                summary['error'] = str(err)
                if debug:
                    summary['debug'] = debug
                pipeline.set_state(Gst.State.NULL)
                break
        else:
            summary['error'] = 'timed out waiting for EOS'
            pipeline.set_state(Gst.State.NULL)
    except Exception as e:  # noqa: BLE001
        summary['error'] = str(e)
    finally:
        finished_at = time.time()
        summary['finished_at'] = time.strftime('%Y-%m-%dT%H:%M:%S%z', time.localtime(finished_at))
        print(json.dumps(summary, indent=2))

    return 0 if summary.get('result') == 'eos' else 1


if __name__ == '__main__':
    raise SystemExit(main())
