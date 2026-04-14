#!/usr/bin/env python3
"""
Multi-level transcript extraction with SQLite storage.
Store: words (with precise timing) → chunks → regenerated SRTs
Allows filtering (remove um/uh) and rebuilding.
"""

import sqlite3
import json
from dataclasses import dataclass
from typing import List, Optional, Tuple
import nemo.collections.asr as nemo_asr
from omegaconf import OmegaConf, open_dict
import subprocess
import os

@dataclass
class Word:
    word: str
    start: float
    end: float
    chunk_id: Optional[int] = None
    is_filler: bool = False
    
@dataclass  
class Chunk:
    id: int
    start: float
    end: float
    text: str
    words: List[Word]

class TranscriptDatabase:
    FILLER_WORDS = {'um', 'uh', 'er', 'ah', 'like', 'you know', 'sort of', 'kind of'}
    
    def __init__(self, db_path: str = "transcript.db"):
        self.db_path = db_path
        self.conn = sqlite3.connect(db_path)
        self.cursor = self.conn.cursor()
        self._init_schema()
    
    def _init_schema(self):
        """Initialize SQLite schema with words, chunks, and configuration tables"""
        
        # Words table: atomic word-level timing
        self.cursor.execute('''
            CREATE TABLE IF NOT EXISTS words (
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
            )
        ''')
        
        # Chunks table: grouping of words into display chunks
        self.cursor.execute('''
            CREATE TABLE IF NOT EXISTS chunks (
                id INTEGER PRIMARY KEY,
                start_time REAL NOT NULL,
                end_time REAL NOT NULL,
                duration REAL GENERATED ALWAYS AS (end_time - start_time) STORED,
                text TEXT NOT NULL,
                word_count INTEGER,
                source_type TEXT DEFAULT 'auto',  -- 'auto', 'manual', 'split_on_word'
                metadata TEXT  -- JSON for extra config
            )
        ''')
        
        # Chunk-word mapping (many-to-many)
        self.cursor.execute('''
            CREATE TABLE IF NOT EXISTS chunk_words (
                chunk_id INTEGER,
                word_id INTEGER,
                position INTEGER,
                PRIMARY KEY (chunk_id, word_id),
                FOREIGN KEY (chunk_id) REFERENCES chunks(id),
                FOREIGN KEY (word_id) REFERENCES words(id)
            )
        ''')
        
        # SRT exports tracking
        self.cursor.execute('''
            CREATE TABLE IF NOT EXISTS srt_exports (
                id INTEGER PRIMARY KEY,
                filename TEXT,
                config TEXT,  -- JSON config used
                segment_count INTEGER,
                word_count INTEGER,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            )
        ''')
        
        # SRT segments
        self.cursor.execute('''
            CREATE TABLE IF NOT EXISTS srt_segments (
                id INTEGER PRIMARY KEY,
                export_id INTEGER,
                sequence_num INTEGER,
                start_time REAL,
                end_time REAL,
                text TEXT,
                FOREIGN KEY (export_id) REFERENCES srt_exports(id)
            )
        ''')
        
        # Indexes for performance
        self.cursor.execute('CREATE INDEX IF NOT EXISTS idx_words_time ON words(start_time, end_time)')
        self.cursor.execute('CREATE INDEX IF NOT EXISTS idx_words_filler ON words(is_filler) WHERE is_filler = 1')
        self.cursor.execute('CREATE INDEX IF NOT EXISTS idx_chunks_time ON chunks(start_time, end_time)')
        self.cursor.execute('CREATE INDEX IF NOT EXISTS idx_words_chunk ON words(chunk_id)')
        
        self.conn.commit()
        print(f"✓ Initialized database: {self.db_path}")
    
    def ingest_audio(self, audio_path: str, model_name: str = "nvidia/nemotron-speech-streaming-en-0.6b"):
        """Extract all words from audio and store in database"""
        
        print(f"\n=== INGESTING AUDIO: {audio_path} ===")
        
        # Load model
        print("Loading ASR model...")
        asr_model = nemo_asr.models.ASRModel.from_pretrained(model_name)
        asr_model = asr_model.cpu().eval()
        
        # Configure for word-level timestamps
        decoding_cfg = asr_model.cfg.decoding
        with open_dict(decoding_cfg):
            decoding_cfg.preserve_alignments = True
            decoding_cfg.compute_timestamps = True
            decoding_cfg.segment_separators = []
            decoding_cfg.word_separator = " "
            asr_model.change_decoding_strategy(decoding_cfg)
        
        time_stride = 8 * asr_model.cfg.preprocessor.window_stride
        
        # Process in chunks
        import soundfile as sf
        info = sf.info(audio_path)
        duration = info.duration
        
        print(f"Audio duration: {duration:.1f}s ({duration/60:.1f} min)")
        print("Extracting words in chunks...")
        
        all_words = []
        chunk_size = 60  # 60 second chunks
        
        for start in range(0, int(duration), chunk_size):
            end = min(start + chunk_size + 2, duration)
            chunk_file = f"temp_db_chunk.wav"
            
            # Extract chunk
            subprocess.run([
                'ffmpeg', '-y', '-i', audio_path,
                '-ss', str(start), '-t', str(end - start),
                '-c', 'copy', chunk_file
            ], capture_output=True, check=True)
            
            # Transcribe
            hypotheses = asr_model.transcribe([chunk_file], return_hypotheses=True, batch_size=1)
            
            # Extract words
            if hypotheses and hypotheses[0] and hasattr(hypotheses[0], 'timestamp'):
                hyp = hypotheses[0]
                if hyp.timestamp:
                    if 'word' in hyp.timestamp:
                        for stamp in hyp.timestamp['word']:
                            word = stamp.get('word', '') or stamp.get('char', '')
                            word_start = stamp.get('start_offset', 0) * time_stride + start
                            word_end = stamp.get('end_offset', 0) * time_stride + start
                            if word.strip():
                                is_filler = word.lower().strip('.,!?') in self.FILLER_WORDS
                                all_words.append((word.strip(), word_start, word_end, is_filler))
                    elif 'char' in hyp.timestamp:
                        # Group chars into words
                        chars = []
                        for stamp in hyp.timestamp['char']:
                            char = stamp.get('char', '')
                            c_start = stamp.get('start_offset', 0) * time_stride + start
                            c_end = stamp.get('end_offset', 0) * time_stride + start
                            chars.append((char, c_start, c_end))
                        
                        current_word = []
                        word_start = None
                        for char, c_start, c_end in chars:
                            if char == ' ':
                                if current_word:
                                    word = ''.join([c[0] for c in current_word])
                                    w_start = current_word[0][1]
                                    w_end = current_word[-1][2]
                                    is_filler = word.lower().strip('.,!?') in self.FILLER_WORDS
                                    all_words.append((word, w_start, w_end, is_filler))
                                    current_word = []
                                    word_start = None
                            else:
                                current_word.append((char, c_start, c_end))
                        
                        if current_word:
                            word = ''.join([c[0] for c in current_word])
                            w_start = current_word[0][1]
                            w_end = current_word[-1][2]
                            is_filler = word.lower().strip('.,!?') in self.FILLER_WORDS
                            all_words.append((word, w_start, w_end, is_filler))
            
            os.remove(chunk_file)
            
            if (start // chunk_size) % 10 == 0:
                print(f"  Processed {start}s... ({len(all_words)} words so far)")
        
        # Insert into database
        print(f"\nInserting {len(all_words)} words into database...")
        
        # Clear existing
        self.cursor.execute("DELETE FROM chunk_words")
        self.cursor.execute("DELETE FROM chunks")
        self.cursor.execute("DELETE FROM words")
        self.conn.commit()
        
        # Insert words
        for word, start, end, is_filler in all_words:
            self.cursor.execute('''
                INSERT INTO words (word, start_time, end_time, is_filler)
                VALUES (?, ?, ?, ?)
            ''', (word, start, end, is_filler))
        
        self.conn.commit()
        
        # Generate initial chunks
        self._create_default_chunks()
        
        print(f"✓ Ingested {len(all_words)} words ({sum(1 for w in all_words if w[3])} fillers)")
    
    def _create_default_chunks(self, max_duration: float = 15.0, max_chars: int = 120, 
                               split_on: List[str] = None):
        """Create chunks from words with default configuration"""
        
        if split_on is None:
            split_on = ['.', '?', '!']
        
        # Get all non-removed words
        self.cursor.execute('''
            SELECT id, word, start_time, end_time 
            FROM words 
            WHERE is_removed = 0
            ORDER BY start_time
        ''')
        words = self.cursor.fetchall()
        
        # Clear existing chunks
        self.cursor.execute("DELETE FROM chunk_words")
        self.cursor.execute("DELETE FROM chunks")
        self.conn.commit()
        
        # Build chunks
        chunks = []
        current_chunk_words = []
        chunk_start = None
        
        for word_id, word, start, end in words:
            if not current_chunk_words:
                chunk_start = start
            
            current_chunk_words.append((word_id, word, start, end))
            
            # Check split conditions
            text = ' '.join([w[1] for w in current_chunk_words])
            duration = end - chunk_start
            
            should_split = False
            if any(word.endswith(p) for p in split_on):
                should_split = True
            if duration > max_duration:
                should_split = True
            if len(text) > max_chars:
                should_split = True
            
            if should_split and current_chunk_words:
                text = ' '.join([w[1] for w in current_chunk_words])
                chunks.append({
                    'start': chunk_start,
                    'end': end,
                    'text': text,
                    'words': [w[0] for w in current_chunk_words]
                })
                current_chunk_words = []
                chunk_start = None
        
        # Final chunk
        if current_chunk_words:
            text = ' '.join([w[1] for w in current_chunk_words])
            end = current_chunk_words[-1][3]
            chunks.append({
                'start': chunk_start,
                'end': end,
                'text': text,
                'words': [w[0] for w in current_chunk_words]
            })
        
        # Insert chunks
        for chunk in chunks:
            self.cursor.execute('''
                INSERT INTO chunks (start_time, end_time, text, word_count)
                VALUES (?, ?, ?, ?)
            ''', (chunk['start'], chunk['end'], chunk['text'], len(chunk['words'])))
            
            chunk_id = self.cursor.lastrowid
            
            for pos, word_id in enumerate(chunk['words']):
                self.cursor.execute('''
                    INSERT INTO chunk_words (chunk_id, word_id, position)
                    VALUES (?, ?, ?)
                ''', (chunk_id, word_id, pos))
                
                # Update word's chunk_id
                self.cursor.execute('''
                    UPDATE words SET chunk_id = ? WHERE id = ?
                ''', (chunk_id, word_id))
        
        self.conn.commit()
        print(f"✓ Created {len(chunks)} chunks")
    
    def remove_filler_words(self) -> int:
        """Mark all filler words as removed and regenerate chunks"""
        
        self.cursor.execute('''
            UPDATE words SET is_removed = 1 WHERE is_filler = 1
        ''')
        removed = self.cursor.rowcount
        self.conn.commit()
        
        print(f"✓ Marked {removed} filler words as removed")
        
        # Regenerate chunks
        self._create_default_chunks()
        
        return removed
    
    def remove_word(self, word: str) -> int:
        """Remove all instances of a specific word"""
        
        self.cursor.execute('''
            UPDATE words SET is_removed = 1 WHERE LOWER(word) = LOWER(?)
        ''', (word,))
        removed = self.cursor.rowcount
        self.conn.commit()
        
        print(f"✓ Removed {removed} instances of '{word}'")
        
        # Regenerate chunks
        self._create_default_chunks()
        
        return removed
    
    def search_word(self, query: str, context_words: int = 5) -> List[dict]:
        """Search for words with context"""
        
        self.cursor.execute('''
            SELECT id, word, start_time, end_time
            FROM words
            WHERE LOWER(word) LIKE LOWER(?) AND is_removed = 0
            ORDER BY start_time
        ''', (f'%{query}%',))
        
        results = []
        for row in self.cursor.fetchall():
            word_id, word, start, end = row
            
            # Get context
            self.cursor.execute('''
                SELECT word FROM words
                WHERE id >= ? - ? AND id <= ? + ? AND is_removed = 0
                ORDER BY id
            ''', (word_id, context_words, word_id, context_words))
            
            context = [r[0] for r in self.cursor.fetchall()]
            
            results.append({
                'word': word,
                'start': start,
                'end': end,
                'context': ' '.join(context)
            })
        
        return results
    
    def export_srt(self, filename: str, 
                   min_duration: float = 0.5,
                   max_duration: float = 30.0,
                   include_removed: bool = False,
                   custom_filter: str = None) -> str:
        """Export chunks to SRT file with filtering options"""
        
        # Get words based on filter
        if include_removed:
            where_clause = "1=1"
            params = ()
        else:
            where_clause = "is_removed = 0"
            params = ()
        
        if custom_filter:
            where_clause += f" AND ({custom_filter})"
        
        self.cursor.execute(f'''
            SELECT word, start_time, end_time, is_filler
            FROM words
            WHERE {where_clause}
            ORDER BY start_time
        ''', params)
        
        words = self.cursor.fetchall()
        
        # Build segments with duration constraints
        segments = []
        current_words = []
        chunk_start = None
        
        for word, start, end, is_filler in words:
            if not current_words:
                chunk_start = start
            
            current_words.append(word)
            
            duration = end - chunk_start
            text = ' '.join(current_words)
            
            should_split = False
            if duration > max_duration or len(text) > 120:
                should_split = True
            if any(word.endswith(p) for p in ['.', '?', '!']) and duration > min_duration:
                should_split = True
            
            if should_split and current_words:
                segments.append({
                    'start': chunk_start,
                    'end': end,
                    'text': ' '.join(current_words)
                })
                current_words = []
                chunk_start = None
        
        if current_words:
            segments.append({
                'start': chunk_start,
                'end': end,
                'text': ' '.join(current_words)
            })
        
        # Write SRT
        def fmt_time(s):
            h = int(s // 3600)
            m = int((s % 3600) // 60)
            sec = int(s % 60)
            ms = int((s % 1) * 1000)
            return f'{h:02d}:{m:02d}:{sec:02d},{ms:03d}'
        
        with open(filename, 'w', encoding='utf-8') as f:
            for i, seg in enumerate(segments, 1):
                f.write(f"{i}\n")
                f.write(f"{fmt_time(seg['start'])} --> {fmt_time(seg['end'])}\n")
                f.write(f"{seg['text']}\n\n")
        
        # Log export
        self.cursor.execute('''
            INSERT INTO srt_exports (filename, config, segment_count, word_count)
            VALUES (?, ?, ?, ?)
        ''', (filename, json.dumps({'min_duration': min_duration, 'max_duration': max_duration, 
                                     'include_removed': include_removed}), 
              len(segments), len(words)))
        export_id = self.cursor.lastrowid
        
        # Store segments
        for i, seg in enumerate(segments, 1):
            self.cursor.execute('''
                INSERT INTO srt_segments (export_id, sequence_num, start_time, end_time, text)
                VALUES (?, ?, ?, ?, ?)
            ''', (export_id, i, seg['start'], seg['end'], seg['text']))
        
        self.conn.commit()
        
        print(f"✓ Exported: {filename}")
        print(f"  Segments: {len(segments)}")
        print(f"  Words: {len(words)}")
        
        return filename
    
    def get_stats(self) -> dict:
        """Get database statistics"""
        
        stats = {}
        
        self.cursor.execute('SELECT COUNT(*) FROM words')
        stats['total_words'] = self.cursor.fetchone()[0]
        
        self.cursor.execute('SELECT COUNT(*) FROM words WHERE is_filler = 1')
        stats['filler_words'] = self.cursor.fetchone()[0]
        
        self.cursor.execute('SELECT COUNT(*) FROM words WHERE is_removed = 1')
        stats['removed_words'] = self.cursor.fetchone()[0]
        
        self.cursor.execute('SELECT COUNT(*) FROM chunks')
        stats['chunks'] = self.cursor.fetchone()[0]
        
        self.cursor.execute('SELECT COUNT(*) FROM srt_exports')
        stats['srt_exports'] = self.cursor.fetchone()[0]
        
        return stats
    
    def close(self):
        self.conn.close()


def main():
    # Initialize database
    db = TranscriptDatabase("audio_transcript.db")
    
    # Check if already ingested
    stats = db.get_stats()
    
    if stats['total_words'] == 0:
        # Ingest audio
        if not os.path.exists("audio-mix_16k_mono.wav"):
            print("Converting audio...")
            subprocess.run([
                'ffmpeg', '-y', '-i', 'audio-mix.wav',
                '-ar', '16000', '-ac', '1', '-c:a', 'pcm_s16le',
                'audio-mix_16k_mono.wav'
            ], capture_output=True, check=True)
        
        db.ingest_audio("audio-mix_16k_mono.wav")
    else:
        print(f"Database already has {stats['total_words']} words")
    
    # Show stats
    stats = db.get_stats()
    print(f"\n=== DATABASE STATS ===")
    print(f"Total words: {stats['total_words']}")
    print(f"Filler words: {stats['filler_words']}")
    print(f"Active chunks: {stats['chunks']}")
    print(f"SRT exports: {stats['srt_exports']}")
    
    # Search example
    print(f"\n=== SEARCH: 'spider' ===")
    results = db.search_word("spider", context_words=8)
    for r in results[:3]:
        print(f"  {r['start']:.1f}s - '{r['word']}' in: '...{r['context']}...'")
    
    # Export options
    print(f"\n=== EXPORTING SRT VARIANTS ===")
    
    # 1. Full transcript (with fillers)
    db.export_srt("full_with_fillers.srt", include_removed=True)
    
    # 2. Clean transcript (no fillers)
    db.export_srt("clean_no_fillers.srt", include_removed=False)
    
    # 3. Very short segments (max 5s)
    db.export_srt("short_segments.srt", max_duration=5.0, include_removed=False)
    
    # 4. Remove specific word and regenerate
    print(f"\n=== REMOVING 'fuck' AND REGENERATING ===")
    db.remove_word("fuck")
    db.export_srt("clean_no_fuck.srt", include_removed=False)
    
    # 5. Remove all fillers and show difference
    print(f"\n=== REMOVING ALL FILLERS ===")
    removed = db.remove_filler_words()
    db.export_srt("no_fillers.srt", include_removed=False)
    
    # Final stats
    stats = db.get_stats()
    print(f"\n=== FINAL STATS ===")
    print(f"Words removed: {stats['removed_words']}")
    print(f"Active chunks: {stats['chunks']}")
    
    db.close()
    
    print(f"\n✓ Done! Files created:")
    for f in ["full_with_fillers.srt", "clean_no_fillers.srt", "short_segments.srt", 
              "clean_no_fuck.srt", "no_fillers.srt"]:
        if os.path.exists(f):
            size = os.path.getsize(f)
            segs = sum(1 for _ in open(f) if _.strip().isdigit())
            print(f"  - {f} ({size} bytes, {segs} segments)")


if __name__ == "__main__":
    main()
