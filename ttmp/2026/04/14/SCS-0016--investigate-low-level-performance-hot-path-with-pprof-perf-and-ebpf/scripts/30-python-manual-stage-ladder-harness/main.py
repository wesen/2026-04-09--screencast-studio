#!/usr/bin/env python3
import argparse
import json
import os
import time
from pathlib import Path

import gi

gi.require_version('Gst', '1.0')
from gi.repository import Gst


def dump_graph(pipeline, name: str) -> None:
    Gst.debug_bin_to_dot_file_with_ts(pipeline, Gst.DebugGraphDetails.ALL, name)


def make(name: str):
    elem = Gst.ElementFactory.make(name)
    if elem is None:
        raise RuntimeError(f'failed to create element {name}')
    return elem


def apply_encoder_settings(elem, encoder_name: str, fps: int, bitrate: int, x264_speed_preset: int, x264_tune: int, x264_bframes: int, x264_trellis: bool) -> None:
    encoder_name = (encoder_name or 'x264enc').strip()
    if encoder_name == 'x264enc':
        elem.set_property('bitrate', bitrate)
        elem.set_property('bframes', x264_bframes)
        elem.set_property('tune', x264_tune)
        elem.set_property('speed-preset', x264_speed_preset)
        elem.set_property('trellis', x264_trellis)
        return
    if encoder_name == 'openh264enc':
        elem.set_property('bitrate', bitrate * 1000)
        elem.set_property('rate-control', 1)
        elem.set_property('usage-type', 1)
        elem.set_property('complexity', 0)
        elem.set_property('gop-size', fps)
        return
    if encoder_name == 'vaapih264enc':
        elem.set_property('bitrate', bitrate)
        elem.set_property('rate-control', 2)
        elem.set_property('keyframe-period', fps)
        elem.set_property('max-bframes', 0)
        elem.set_property('quality-level', 7)
        return
    raise RuntimeError(f'unsupported encoder {encoder_name!r}')


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument('--display-name', default=os.environ.get('DISPLAY', ':0').strip() or ':0')
    ap.add_argument('--stage', default='capture')
    ap.add_argument('--fps', type=int, default=24)
    ap.add_argument('--bitrate', type=int, default=6920)
    ap.add_argument('--encoder', default='x264enc')
    ap.add_argument('--x264-speed-preset', type=int, default=3)
    ap.add_argument('--x264-tune', type=int, default=4)
    ap.add_argument('--x264-bframes', type=int, default=0)
    ap.add_argument('--x264-trellis', default='true')
    ap.add_argument('--container', default='mov')
    ap.add_argument('--duration-seconds', type=int, default=8)
    ap.add_argument('--output-path', default=str(Path('/tmp') / 'scs-python-stage-ladder.mov'))
    ap.add_argument('--dot-dir', default='')
    args = ap.parse_args()

    summary = {
        'started_at': time.strftime('%Y-%m-%dT%H:%M:%S%z'),
        'display_name': args.display_name,
        'stage': args.stage,
        'fps': args.fps,
        'bitrate': args.bitrate,
        'encoder': args.encoder,
        'x264_speed_preset': args.x264_speed_preset,
        'x264_tune': args.x264_tune,
        'x264_bframes': args.x264_bframes,
        'x264_trellis': str(args.x264_trellis).lower() == 'true',
        'container': args.container,
        'output_path': args.output_path,
        'duration_seconds': args.duration_seconds,
    }

    try:
        if args.dot_dir:
            Path(args.dot_dir).mkdir(parents=True, exist_ok=True)
            os.environ['GST_DEBUG_DUMP_DOT_DIR'] = args.dot_dir
        if args.stage == 'mux-file':
            Path(args.output_path).parent.mkdir(parents=True, exist_ok=True)
        Gst.init(None)
        pipeline = Gst.Pipeline.new(None)
        if pipeline is None:
            raise RuntimeError('failed to create pipeline')

        elems = []
        ximagesrc = make('ximagesrc')
        ximagesrc.set_property('display-name', args.display_name)
        ximagesrc.set_property('show-pointer', True)
        ximagesrc.set_property('use-damage', False)
        elems.append(ximagesrc)

        stage = args.stage.strip()
        if stage in ('convert', 'rate-caps', 'encode', 'parse', 'mux-file'):
            elems.append(make('videoconvert'))
        if stage in ('rate-caps', 'encode', 'parse', 'mux-file'):
            elems.append(make('videorate'))
            capsfilter = make('capsfilter')
            capsfilter.set_property('caps', Gst.Caps.from_string(f'video/x-raw,format=I420,framerate={args.fps}/1,pixel-aspect-ratio=1/1'))
            elems.append(capsfilter)
        if stage in ('encode', 'parse', 'mux-file'):
            encoder_name = (args.encoder or 'x264enc').strip()
            enc = make(encoder_name)
            apply_encoder_settings(enc, encoder_name, args.fps, args.bitrate, args.x264_speed_preset, args.x264_tune, args.x264_bframes, str(args.x264_trellis).lower() == 'true')
            elems.append(enc)
        if stage in ('parse', 'mux-file'):
            elems.append(make('h264parse'))

        if stage == 'mux-file':
            mux_name = 'qtmux' if args.container.lower() in ('', 'mov', 'qt') else 'mp4mux'
            if args.container.lower() not in ('', 'mov', 'qt', 'mp4'):
                raise RuntimeError(f'unsupported container {args.container!r}')
            elems.append(make(mux_name))
            fs = make('filesink')
            fs.set_property('location', args.output_path)
            elems.append(fs)
        elif stage in ('capture', 'convert', 'rate-caps', 'encode', 'parse'):
            fs = make('fakesink')
            fs.set_property('sync', False)
            elems.append(fs)
        else:
            raise RuntimeError(f'unsupported stage {stage!r}')

        for elem in elems:
            pipeline.add(elem)
        for a, b in zip(elems, elems[1:]):
            if not a.link(b):
                raise RuntimeError(f'failed to link {a.get_name()} -> {b.get_name()}')

        bus = pipeline.get_bus()
        dump_graph(pipeline, f'python-stage-ladder-{stage}-pre-play')
        if pipeline.set_state(Gst.State.PLAYING) == Gst.StateChangeReturn.FAILURE:
            raise RuntimeError('failed to set pipeline to PLAYING')
        time.sleep(1)
        dump_graph(pipeline, f'python-stage-ladder-{stage}-playing')
        if args.duration_seconds > 1:
            time.sleep(args.duration_seconds - 1)
        pipeline.send_event(Gst.Event.new_eos())
        deadline = time.time() + 20
        while time.time() < deadline:
            msg = bus.timed_pop_filtered(500 * Gst.MSECOND, Gst.MessageType.EOS | Gst.MessageType.ERROR)
            if msg is None:
                continue
            if msg.type == Gst.MessageType.EOS:
                summary['result'] = 'eos'
                dump_graph(pipeline, f'python-stage-ladder-{stage}-eos')
                pipeline.set_state(Gst.State.NULL)
                break
            if msg.type == Gst.MessageType.ERROR:
                err, debug = msg.parse_error()
                summary['error'] = str(err)
                if debug:
                    summary['debug'] = debug
                dump_graph(pipeline, f'python-stage-ladder-{stage}-error')
                pipeline.set_state(Gst.State.NULL)
                break
        else:
            summary['error'] = 'timed out waiting for EOS'
            pipeline.set_state(Gst.State.NULL)
    except Exception as e:
        summary['error'] = str(e)
    finally:
        summary['finished_at'] = time.strftime('%Y-%m-%dT%H:%M:%S%z')
        print(json.dumps(summary, indent=2))
    return 0 if summary.get('result') == 'eos' else 1


if __name__ == '__main__':
    raise SystemExit(main())
