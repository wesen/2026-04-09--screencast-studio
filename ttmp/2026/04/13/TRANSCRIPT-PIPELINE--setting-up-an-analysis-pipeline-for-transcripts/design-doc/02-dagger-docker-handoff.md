---
ticket: TRANSCRIPT-PIPELINE
doc_type: design-doc
title: Dagger + Docker Handoff - Transcription Pipeline Containerization
status: active
intent: long-term
topics: transcripts, audio, dagger, docker, go
---

# Dagger + Docker Handoff

For: Colleague building Go + Dagger + Docker version of the transcription pipeline

---

## Quick Overview

Transcribe audio using NVIDIA Nemotron 0.6B ASR → Extract word-level timestamps → Store in SQLite → Export SRT files with optional filtering.

**Runtime:** ~5 min for 30-min audio on CPU  
**Model size:** ~1.2GB (downloaded once)  
**Output:** SQLite DB + SRT files

---

## Essential Files to Copy

From: `/home/manuel/code/wesen/patreon/videos/002-sqrt/`

```
transcript_db.py      ← Main pipeline (processes in chunks, SQLite storage)
query_transcript.py   ← Query tool for searching/filtering
requirements.txt      ← Create this (dependencies below)
```

**Location in ticket:** `ttmp/2026/04/13/TRANSCRIPT-PIPELINE--setting-up-an-analysis-pipeline-for-transcripts/scripts/`

---

## Key Dependencies (requirements.txt)

```
--extra-index-url https://download.pytorch.org/whl/cpu
torch==2.11.0
torchaudio
soundfile==0.13.1
omegaconf==2.3.0
pytorch-lightning==2.6.1

# Install from git (NeMo toolkit)
git+https://github.com/NVIDIA/NeMo.git@e99060017bafe5e6a7443c486ee259f81b9049e8#egg=nemo_toolkit[asr]
```

**Working reference:** `/home/manuel/code/wesen/patreon/videos/002-sqrt/venv/` (Python 3.13.2 venv with working deps)

---

## Pipeline Flow

```
Input: audio-mix.wav
    ↓
ffmpeg -ar 16000 -ac 1 → audio-mix_16k_mono.wav
    ↓
NVIDIA Nemotron 0.6B ASR (60s chunks, CPU)
    ↓
SQLite: words(id, word, start_time, end_time, is_filler, is_removed)
       + chunks(id, start_time, end_time, text)
       + chunk_words(mapping)
    ↓
Export: no_fillers.srt (or other variants with filtering)
```

---

## Critical Implementation Details

### 1. Audio Preprocessing (ffmpeg)

```bash
ffmpeg -y -i audio-mix.wav -ar 16000 -ac 1 -c:a pcm_s16le audio-mix_16k_mono.wav
```
- Must be 16kHz mono WAV for Nemotron
- Do this before transcription

### 2. ASR Model Setup

```python
import nemo.collections.asr as nemo_asr
from omegaconf import OmegaConf, open_dict

asr_model = nemo_asr.models.ASRModel.from_pretrained(
    "nvidia/nemotron-speech-streaming-en-0.6b"
)
asr_model = asr_model.cpu().eval()

# Enable word-level timestamps
decoding_cfg = asr_model.cfg.decoding
with open_dict(decoding_cfg):
    decoding_cfg.preserve_alignments = True
    decoding_cfg.compute_timestamps = True
    decoding_cfg.segment_separators = []
    decoding_cfg.word_separator = " "
    asr_model.change_decoding_strategy(decoding_cfg)
```

### 3. Chunked Processing (Prevents OOM)

**Critical:** Process audio in 60-second chunks. Full file at once causes OOM on long audio.

```python
import soundfile as sf
info = sf.info(audio_path)
duration = info.duration

chunk_size = 60  # seconds
for start in range(0, int(duration), chunk_size):
    end = min(start + chunk_size + 2, duration)
    # Extract chunk with ffmpeg
    subprocess.run([
        'ffmpeg', '-y', '-i', audio_path,
        '-ss', str(start), '-t', str(end - start),
        '-c', 'copy', 'temp_chunk.wav'
    ])
    # Transcribe chunk
    hypotheses = asr_model.transcribe(['temp_chunk.wav'], return_hypotheses=True, batch_size=1)
    # Extract word timestamps, adjust by chunk start time
    # ...
    os.remove('temp_chunk.wav')
```

### 4. Word Timestamp Calculation

```python
time_stride = 8 * asr_model.cfg.preprocessor.window_stride

# From hypothesis timestamp
for stamp in hyp.timestamp['word']:
    word = stamp.get('word', '')
    word_start = stamp.get('start_offset', 0) * time_stride + chunk_start
    word_end = stamp.get('end_offset', 0) * time_stride + chunk_start
```

### 5. Filler Word Detection

```python
FILLER_WORDS = {'um', 'uh', 'er', 'ah', 'like', 'you know', 'sort of', 'kind of'}
is_filler = word.lower().strip('.,!?') in FILLER_WORDS
```

### 6. SQLite Schema (Minimal)

```sql
CREATE TABLE words (
    id INTEGER PRIMARY KEY,
    word TEXT NOT NULL,
    start_time REAL NOT NULL,
    end_time REAL NOT NULL,
    is_filler BOOLEAN DEFAULT 0,
    is_removed BOOLEAN DEFAULT 0
);

CREATE TABLE chunks (
    id INTEGER PRIMARY KEY,
    start_time REAL NOT NULL,
    end_time REAL NOT NULL,
    text TEXT NOT NULL
);
```

### 7. SRT Export

```python
def fmt_time(s):
    h = int(s // 3600)
    m = int((s % 3600) // 60)
    sec = int(s % 60)
    ms = int((s % 1) * 1000)
    return f'{h:02d}:{m:02d}:{sec:02d},{ms:03d}'

# Build segments with duration constraints
segments = []
current_words = []
chunk_start = None

for word, start, end in words:
    if not current_words:
        chunk_start = start
    current_words.append(word)
    
    duration = end - chunk_start
    text = ' '.join(current_words)
    
    # Split on sentence end or max duration
    should_split = False
    if any(word.endswith(p) for p in ['.', '?', '!']) and duration > 0.5:
        should_split = True
    if duration > 15.0 or len(text) > 120:
        should_split = True
    
    if should_split:
        segments.append({'start': chunk_start, 'end': end, 'text': text})
        current_words = []
        chunk_start = None
```

---

## Docker/Dagger Considerations

### Model Caching

The Nemotron 0.6B model (~1.2GB) is downloaded to:
```
~/.cache/huggingface/hub/models--nvidia--nemotron-speech-streaming-en-0.6b/
```

**Recommendation:** Mount/cache this directory to avoid re-downloading.

### CPU-Only Runtime

```python
device = torch.device("cpu")
asr_model = asr_model.to(device)
asr_model.eval()
```

No GPU needed. CPU inference is slower but sufficient.

### File I/O

Pipeline reads/writes:
- Input: `*.wav` (any audio ffmpeg can convert)
- Working: `*_16k_mono.wav`, `temp_chunk.wav`
- Output: `audio_transcript.db`, `*.srt`

---

## Go + Dagger Integration Points

1. **Audio conversion:** Dagger container with ffmpeg
2. **Transcription:** Dagger container with Python + NeMo
3. **Model cache:** Dagger cache volume for `~/.cache/huggingface`
4. **Output:** SQLite DB + SRT files as Dagger artifacts

**CLI interface sketch:**
```bash
transcribe --input audio.wav --output-dir ./out/
# → out/audio_transcript.db
# → out/audio.srt
```

---

## Testing

Test audio available at:
- `/home/manuel/code/wesen/2026-04-09--screencast-studio/recordings/rabbit-hole-2026-04-10--2/audio-mix.wav`
- Expected output: 4,248 words, ~5 min processing on CPU

---

## Reference Documents

| Doc | Location | Purpose |
|-----|----------|---------|
| Full transcription report | `reference/01-rabbit-hole-recording-2-transcription-report.md` | Step-by-step pipeline guide |
| Investigation diary | `reference/02-investigation-diary.md` | What worked, what didn't |
| Python scripts | `scripts/transcript_db.py` | Working implementation |
| Sample output | `sources/rabbit-hole-2-transcript.srt` | Example SRT format |
| SQLite DB | `sources/audio_transcript.db` | Example database |

---

## Contact / Questions

Ticket: `TRANSCRIPT-PIPELINE` in docmgr  
Original working pipeline: `/home/manuel/code/wesen/patreon/videos/002-sqrt/`
