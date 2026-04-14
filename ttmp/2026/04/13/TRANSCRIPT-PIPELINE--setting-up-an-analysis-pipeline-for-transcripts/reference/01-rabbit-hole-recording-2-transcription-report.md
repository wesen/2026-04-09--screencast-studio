---
Title: ""
Ticket: ""
Status: ""
Topics: []
DocType: ""
Intent: ""
Owners: []
RelatedFiles:
    - Path: ttmp/2026/04/13/TRANSCRIPT-PIPELINE--setting-up-an-analysis-pipeline-for-transcripts/scripts/query_transcript.py
      Note: Query and custom export tool
    - Path: ttmp/2026/04/13/TRANSCRIPT-PIPELINE--setting-up-an-analysis-pipeline-for-transcripts/scripts/transcribe_to_srt.py
      Note: Basic SRT transcription script
    - Path: ttmp/2026/04/13/TRANSCRIPT-PIPELINE--setting-up-an-analysis-pipeline-for-transcripts/scripts/transcript_db.py
      Note: Main transcription pipeline script
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# Rabbit Hole Recording 2 Transcription Report

## Summary

This report documents the transcription of the second rabbit hole recording using the NVIDIA Nemotron Speech Streaming 0.6B ASR pipeline. The transcription was successfully completed on 2026-04-13.

### Transcription Results

| Metric | Value |
|--------|-------|
| **Audio Duration** | 27.7 minutes (1664.8 seconds) |
| **Total Words** | 4,248 |
| **Filler Words** | 173 (4.1%) |
| **Segments** | 192-303 depending on filtering |
| **Processing Time** | ~5 minutes on CPU |
| **Model** | NVIDIA Nemotron Speech Streaming 0.6B |

### Output Files Generated

- `no_fillers.srt` - Clean transcript (192 segments, fillers removed) → **Copy in ticket sources**
- `full_with_fillers.srt` - Complete transcript (198 segments)  
- `short_segments.srt` - Short segments (303 segments, max 5s)
- `audio_transcript.db` - SQLite database with word-level timestamps → **Copy in ticket sources**

---

## Transcript Sample (First 60 Seconds)

```
[00:00:02] Welcome back to the Go Go Golems lab spider welcome back to the 
[00:00:16] welcome back to the Go Go Golems lab today a special episode we're 
[00:00:25] gonna be rabbit holeing with LLMs I'm sure you've gone down deep the 
[00:00:35] rabbit hole talking about one thing or the other. I'll show you what my 
[00:00:42] workflow looks and how I use it to basically learn new things, learn 
[00:00:47] new skills both in terms of how LLMs work, what you can do with LLMs, 
[00:00:57] what can tools do that are out there, but also just pure fundamental 
[00:01:00] knowledge that will help me create better software at the end of the day.
[00:01:08] So let's imagine that I wanna do something random I have a certain task 
[00:01:19] I'm gonna go to Chat GPT and I'm gonna use that as a starting point for 
[00:01:27] a rabbit hole. So today's rabbit hole it's a pretty stupid one. It's 
[00:01:36] I want to do the square root of something so square root of and I'm 
[00:01:43] gonna see you know I'm using this to show how LLMs can be wrong as well...
```

**Full transcript location:** `sources/rabbit-hole-2-transcript.srt` (in this ticket)

---

## Recording Details

| Property | Value |
|----------|-------|
| **File** | `recordings/rabbit-hole-2026-04-10--2/audio-mix.wav` |
| **Size** | ~305 MB |
| **Format** | WAV (stereo, needs 16kHz mono conversion) |
| **Pipeline** | NVIDIA Nemotron Speech Streaming 0.6B (CPU-only) |

---

## How the Transcription Pipeline Works

### 1. Audio Preprocessing

```bash
ffmpeg -y -i audio-mix.wav -ar 16000 -ac 1 -c:a pcm_s16le audio-mix_16k_mono.wav
```

Converts to 16kHz mono WAV format required by the ASR model.

### 2. Word-Level Extraction (transcript_db.py)

The pipeline uses NVIDIA's NeMo ASR toolkit with the Nemotron Speech Streaming 0.6B model:

```python
import nemo.collections.asr as nemo_asr

asr_model = nemo_asr.models.ASRModel.from_pretrained(
    "nvidia/nemotron-speech-streaming-en-0.6b"
)

# Configure for word-level timestamps
decoding_cfg = asr_model.cfg.decoding
with open_dict(decoding_cfg):
    decoding_cfg.preserve_alignments = True
    decoding_cfg.compute_timestamps = True
    decoding_cfg.segment_separators = []
    decoding_cfg.word_separator = " "
    asr_model.change_decoding_strategy(decoding_cfg)
```

### 3. Database Schema

Words are stored in SQLite with this schema:

```sql
-- Words table: atomic word-level timing
CREATE TABLE words (
    id INTEGER PRIMARY KEY,
    word TEXT NOT NULL,
    start_time REAL NOT NULL,
    end_time REAL NOT NULL,
    duration REAL GENERATED ALWAYS AS (end_time - start_time) STORED,
    chunk_id INTEGER,
    is_filler BOOLEAN DEFAULT 0,
    is_removed BOOLEAN DEFAULT 0,
    confidence REAL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Chunks table: grouping of words into display chunks
CREATE TABLE chunks (
    id INTEGER PRIMARY KEY,
    start_time REAL NOT NULL,
    end_time REAL NOT NULL,
    duration REAL GENERATED ALWAYS AS (end_time - start_time) STORED,
    text TEXT NOT NULL,
    word_count INTEGER,
    source_type TEXT DEFAULT 'auto',
    metadata TEXT
);

-- Chunk-word mapping (many-to-many)
CREATE TABLE chunk_words (
    chunk_id INTEGER,
    word_id INTEGER,
    position INTEGER,
    PRIMARY KEY (chunk_id, word_id)
);
```

### 4. SRT Export Options

The pipeline supports multiple export configurations:

| Export | Description | Command |
|--------|-------------|---------|
| `full_with_fillers.srt` | Complete transcript including fillers | `db.export_srt("full_with_fillers.srt", include_removed=True)` |
| `clean_no_fillers.srt` | No filler words (um, uh, like, etc.) | `db.export_srt("clean_no_fillers.srt", include_removed=False)` |
| `short_segments.srt` | Max 5 second segments | `db.export_srt("short_segments.srt", max_duration=5.0)` |
| `custom_no_fillers.srt` | Exclude specific words | `q.export_custom_srt("custom.srt", exclude_words=["like", "um"])` |

---

## Step-by-Step Process for Rabbit Hole Recording 2

### Step 1: Navigate to the recording directory

```bash
cd /home/manuel/code/wesen/2026-04-09--screencast-studio/recordings/rabbit-hole-2026-04-10--2/
```

### Step 2: Copy the transcription scripts

```bash
cp /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/TRANSCRIPT-PIPELINE--setting-up-an-analysis-pipeline-for-transcripts/scripts/*.py .
```

### Step 3: Convert audio to required format

```bash
ffmpeg -y -i audio-mix.wav -ar 16000 -ac 1 -c:a pcm_s16le audio-mix_16k_mono.wav
```

### Step 4: Run the full pipeline

```bash
# Use the working venv from the sqrt project (Python 3.13.2)
# The venv has all compatible dependencies
/home/manuel/code/wesen/patreon/videos/002-sqrt/venv/bin/python \
  /home/manuel/code/wesen/patreon/videos/002-sqrt/transcript_db.py
```

**Note:** We initially tried pyenv 3.11.3 but encountered dependency conflicts with torch/torchvision. The sqrt project's venv has working dependencies.

This will:
1. Load the Nemotron 0.6B model (~1.2GB download on first run)
2. Process audio in 60-second chunks (avoids OOM on long audio)
3. Extract all words with timestamps
4. Create default chunks with sentence boundaries
5. Generate multiple SRT variants

**Actual runtime:** ~5 minutes for 27.7-minute audio on CPU (faster than expected due to chunking)

### Step 5: Query the transcript

```bash
# Run query tool using the same venv
/home/manuel/code/wesen/patreon/videos/002-sqrt/venv/bin/python \
  /home/manuel/code/wesen/patreon/videos/002-sqrt/query_transcript.py --cli
```

Interactive commands:
- `search <word>` - Find all instances of a word
- `context <id>` - Show context around word ID  
- `fillers` - List filler words with counts (found 173 fillers in this transcript)
- `freq` - Show most common words
- `export <filename>` - Generate custom SRT with custom filtering

**Sample search results for "spider":**
```
Found: 'spider' at 8.0s (chunk: 1)
  In: 'back to the Go Go Golems lab fuck spider welcome back to the...'
```

---

## Pipeline Scripts Reference

| Script | Purpose | Key Functions |
|--------|---------|---------------|
| `transcribe_to_srt.py` | Basic SRT generation | `transcribe_to_srt()`, format_time() |
| `transcript_db.py` | Full database pipeline | `TranscriptDatabase.ingest_audio()`, `export_srt()`, `remove_filler_words()` |
| `query_transcript.py` | Query and custom export | `TranscriptQuery.search()`, `export_custom_srt()`, interactive mode |

---

## Output Files

After running the pipeline, these files are generated:

```
rabbit-hole-2026-04-10--2/
├── audio-mix_16k_mono.wav          # Converted audio
├── audio_transcript.db              # SQLite database with word-level data
├── full_with_fillers.srt            # Complete transcript
├── clean_no_fillers.srt             # No filler words
├── short_segments.srt               # 5-second max segments
├── clean_no_fuck.srt                # Specific word removed example
└── no_fillers.srt                   # All fillers removed
```

---

## Key Capabilities

### 1. Filler Word Detection

Default filler words: `{'um', 'uh', 'er', 'ah', 'like', 'you know', 'sort of', 'kind of'}`

```python
db.remove_filler_words()  # Marks fillers as removed, regenerates chunks
```

### 2. Custom Word Removal

```python
db.remove_word("fuck")  # Remove all instances of specific word
```

### 3. Search with Context

```python
results = db.search_word("spider", context_words=8)
# Returns: word, start_time, end_time, surrounding context
```

### 4. Segment Duration Control

```python
db.export_srt("output.srt", 
              min_duration=0.5,   # Minimum segment length
              max_duration=15.0,  # Maximum segment length
              include_removed=False)  # Whether to include removed words
```

---

## Dependencies

Required packages (auto-installed by scripts):
- `nemo_toolkit[asr]` - NVIDIA NeMo ASR toolkit
- `torch`, `torchaudio` - PyTorch
- `soundfile` - Audio file handling
- `omegaconf` - Configuration management

---

## Transcription Completed ✓

The transcription of Rabbit Hole Recording 2 has been completed successfully. The transcript is available in the ticket's `sources/` folder:

- **Transcript (SRT):** `sources/rabbit-hole-2-transcript.srt`  
- **Database:** `sources/audio_transcript.db` (SQLite with word-level timestamps)

### Next Steps

1. Use query tools to analyze the transcript content
2. Search for key terms ("spider", "LLM", "rabbit hole", etc.)
3. Create summaries or extract key segments
4. Apply the same pipeline to other recordings

---

## References

| Resource | Location |
|----------|----------|
| **Transcript (SRT)** | `sources/rabbit-hole-2-transcript.srt` (this ticket) |
| **Transcript (DB)** | `sources/audio_transcript.db` (this ticket) |
| **Pipeline Scripts** | `scripts/` (this ticket) |
| **Original Pipeline** | `/home/manuel/code/wesen/patreon/videos/002-sqrt/` |
| **Source Recording** | `recordings/rabbit-hole-2026-04-10--2/audio-mix.wav` |
| **NVIDIA Nemotron** | https://huggingface.co/nvidia/nemotron-speech-streaming-en-0.6b |

---

**Generated:** 2026-04-13  
**Ticket:** TRANSCRIPT-PIPELINE  
**Status:** ✓ Transcription Complete
