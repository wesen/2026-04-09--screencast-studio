#!/usr/bin/env python3
"""
Transcribe audio to SRT using NVIDIA Nemotron Speech Streaming 0.6B ASR model.
CPU-only version - no GPU required.
"""

import sys
import os
import subprocess
import json

def ensure_dependencies():
    """Install required packages if not present."""
    try:
        import nemo.collections.asr
        import torch
        import soundfile
    except ImportError:
        print("Installing required dependencies...")
        subprocess.check_call([sys.executable, "-m", "pip", "install", "-q", "Cython", "packaging"])
        subprocess.check_call([sys.executable, "-m", "pip", "install", "-q",
            "git+https://github.com/NVIDIA/NeMo.git@main#egg=nemo_toolkit[asr]",
            "soundfile", "torch", "torchaudio"
        ])

def convert_audio(input_path, output_path):
    """Convert audio to 16kHz mono WAV format required by the model."""
    print(f"Converting {input_path} to 16kHz mono...")
    cmd = [
        "ffmpeg", "-y", "-i", input_path,
        "-ar", "16000", "-ac", "1", "-c:a", "pcm_s16le",
        output_path
    ]
    subprocess.run(cmd, capture_output=True, check=True)
    print(f"Converted: {output_path}")

def format_time(seconds):
    """Convert seconds to SRT timestamp format: HH:MM:SS,mmm"""
    hours = int(seconds // 3600)
    minutes = int((seconds % 3600) // 60)
    secs = int(seconds % 60)
    millis = int((seconds % 1) * 1000)
    return f"{hours:02d}:{minutes:02d}:{secs:02d},{millis:03d}"

def transcribe_to_srt(audio_path, srt_path, model_name="nvidia/nemotron-speech-streaming-en-0.6b"):
    """Transcribe audio and save as SRT file with timestamps."""

    import nemo.collections.asr as nemo_asr
    import torch

    print(f"Loading model: {model_name}")
    print("This may take a few minutes on first run (downloading ~1.2GB model)...")
    asr_model = nemo_asr.models.ASRModel.from_pretrained(model_name)

    # CPU-only mode
    device = torch.device("cpu")
    asr_model = asr_model.to(device)
    asr_model.eval()
    print("Using CPU for inference (this will be slower than GPU)")

    # Configure decoding for timestamps
    from omegaconf import OmegaConf, open_dict
    decoding_cfg = asr_model.cfg.decoding
    with open_dict(decoding_cfg):
        decoding_cfg.preserve_alignments = True
        decoding_cfg.compute_timestamps = True
        decoding_cfg.segment_separators = [".", "?", "!"]
        decoding_cfg.word_separator = " "
        asr_model.change_decoding_strategy(decoding_cfg)

    print(f"Transcribing: {audio_path}")
    print("Processing... (this may take several minutes for long audio)")

    # For CPU, process in smaller batches to avoid memory issues
    hypotheses = asr_model.transcribe([audio_path], return_hypotheses=True, batch_size=1)

    # Extract segment timestamps
    time_stride = 8 * asr_model.cfg.preprocessor.window_stride

    segments = []
    if hypotheses and len(hypotheses) > 0:
        hyp = hypotheses[0]

        # Try to get segment timestamps
        if hasattr(hyp, 'timestamp') and hyp.timestamp:
            timestamp_dict = hyp.timestamp

            if 'segment' in timestamp_dict:
                segment_timestamps = timestamp_dict['segment']
                for i, stamp in enumerate(segment_timestamps):
                    start = stamp.get('start_offset', 0) * time_stride
                    end = stamp.get('end_offset', 0) * time_stride
                    segment = stamp.get('segment', '')
                    if segment.strip():
                        segments.append({
                            'index': i + 1,
                            'start': start,
                            'end': end,
                            'text': segment.strip()
                        })

        # Fallback: if no segments, use the full text
        if not segments:
            text = hyp.text if hasattr(hyp, 'text') else str(hyp)
            # Get audio duration
            import soundfile as sf
            info = sf.info(audio_path)
            segments.append({
                'index': 1,
                'start': 0.0,
                'end': info.duration,
                'text': text.strip()
            })

    # Write SRT file
    print(f"Writing SRT: {srt_path}")
    with open(srt_path, 'w', encoding='utf-8') as f:
        for seg in segments:
            f.write(f"{seg['index']}\n")
            f.write(f"{format_time(seg['start'])} --> {format_time(seg['end'])}\n")
            f.write(f"{seg['text']}\n\n")

    print(f"✓ Transcription complete: {srt_path}")
    print(f"  Segments: {len(segments)}")
    if segments:
        total_duration = segments[-1]['end']
        print(f"  Duration: {format_time(total_duration)}")

def main():
    input_audio = "audio-mix.wav"
    converted_audio = "audio-mix_16k_mono.wav"
    output_srt = "audio-mix.srt"

    # Check input file
    if not os.path.exists(input_audio):
        print(f"Error: {input_audio} not found")
        sys.exit(1)

    # Install dependencies
    ensure_dependencies()

    # Convert audio to required format
    if not os.path.exists(converted_audio):
        convert_audio(input_audio, converted_audio)

    # Transcribe
    transcribe_to_srt(converted_audio, output_srt)

    print("\nDone!")

if __name__ == "__main__":
    main()
