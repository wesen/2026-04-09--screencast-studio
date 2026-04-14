#!/usr/bin/env python3
"""
Query tool for the transcript database.
Search, filter, and regenerate SRTs with custom queries.
"""

import sqlite3
import sys
from dataclasses import dataclass

@dataclass
class WordInfo:
    id: int
    word: str
    start: float
    end: float
    is_filler: bool
    is_removed: bool

class TranscriptQuery:
    def __init__(self, db_path: str = "audio_transcript.db"):
        self.conn = sqlite3.connect(db_path)
        self.conn.row_factory = sqlite3.Row
    
    def get_word_context(self, word_id: int, context_size: int = 5):
        """Get words around a specific word"""
        cursor = self.conn.execute('''
            SELECT * FROM words 
            WHERE id >= ? - ? AND id <= ? + ? AND is_removed = 0
            ORDER BY id
        ''', (word_id, context_size, word_id, context_size))
        return cursor.fetchall()
    
    def search(self, query: str, limit: int = 20):
        """Search for words"""
        cursor = self.conn.execute('''
            SELECT w.*, c.text as chunk_text
            FROM words w
            LEFT JOIN chunks c ON w.chunk_id = c.id
            WHERE w.word LIKE ? AND w.is_removed = 0
            ORDER BY w.start_time
            LIMIT ?
        ''', (f'%{query}%', limit))
        return cursor.fetchall()
    
    def get_fillers(self):
        """List all filler words with counts"""
        cursor = self.conn.execute('''
            SELECT word, COUNT(*) as count, 
                   GROUP_CONCAT(printf('%.1f', start_time), ', ') as times
            FROM words
            WHERE is_filler = 1
            GROUP BY LOWER(word)
            ORDER BY count DESC
        ''')
        return cursor.fetchall()
    
    def get_removed_words(self):
        """List removed words"""
        cursor = self.conn.execute('''
            SELECT word, start_time, end_time
            FROM words
            WHERE is_removed = 1
            ORDER BY start_time
        ''')
        return cursor.fetchall()
    
    def get_word_frequency(self):
        """Get word frequency analysis"""
        cursor = self.conn.execute('''
            SELECT LOWER(word) as word_lower, COUNT(*) as count
            FROM words
            WHERE is_removed = 0 AND is_filler = 0
            GROUP BY LOWER(word)
            HAVING count > 5
            ORDER BY count DESC
            LIMIT 20
        ''')
        return cursor.fetchall()
    
    def export_custom_srt(self, output_file: str, 
                          exclude_words: list = None,
                          max_segment_duration: float = 15.0,
                          min_segment_duration: float = 0.5):
        """Export SRT with custom filtering"""
        
        if exclude_words is None:
            exclude_words = []
        
        # Build query
        placeholders = ','.join(['?' for _ in exclude_words])
        where_clause = "is_removed = 0"
        if exclude_words:
            where_clause += f" AND LOWER(word) NOT IN ({placeholders})"
        
        cursor = self.conn.execute(f'''
            SELECT word, start_time, end_time
            FROM words
            WHERE {where_clause}
            ORDER BY start_time
        ''', [w.lower() for w in exclude_words])
        
        words = cursor.fetchall()
        
        # Build segments
        def fmt_time(s):
            h = int(s // 3600)
            m = int((s % 3600) // 60)
            sec = int(s % 60)
            ms = int((s % 1) * 1000)
            return f'{h:02d}:{m:02d}:{sec:02d},{ms:03d}'
        
        segments = []
        current = []
        chunk_start = None
        
        for word, start, end in words:
            if not current:
                chunk_start = start
            
            current.append(word)
            
            duration = end - chunk_start
            text = ' '.join(current)
            
            should_split = False
            if any(word.endswith(p) for p in ['.', '?', '!']) and duration > min_segment_duration:
                should_split = True
            if duration > max_segment_duration or len(text) > 120:
                should_split = True
            
            if should_split and current:
                segments.append({
                    'start': chunk_start,
                    'end': end,
                    'text': ' '.join(current)
                })
                current = []
                chunk_start = None
        
        if current:
            segments.append({
                'start': chunk_start,
                'end': end,
                'text': ' '.join(current)
            })
        
        # Write
        with open(output_file, 'w', encoding='utf-8') as f:
            for i, seg in enumerate(segments, 1):
                f.write(f"{i}\n")
                f.write(f"{fmt_time(seg['start'])} --> {fmt_time(seg['end'])}\n")
                f.write(f"{seg['text']}\n\n")
        
        print(f"✓ Exported {output_file} ({len(segments)} segments, {len(words)} words)")
        return output_file
    
    def close(self):
        self.conn.close()


def interactive_query():
    """Interactive query mode"""
    q = TranscriptQuery()
    
    print("\n=== TRANSCRIPT QUERY TOOL ===")
    print("Commands:")
    print("  search <word>    - Search for word")
    print("  context <id>     - Show context around word ID")
    print("  fillers          - Show filler words")
    print("  freq             - Show word frequency")
    print("  removed          - Show removed words")
    print("  export <file>    - Export custom SRT")
    print("  exit             - Quit")
    
    while True:
        try:
            cmd = input("\n> ").strip().split()
            if not cmd:
                continue
            
            if cmd[0] == "exit":
                break
            
            elif cmd[0] == "search" and len(cmd) > 1:
                results = q.search(cmd[1])
                print(f"\nFound {len(results)} matches:")
                for r in results[:10]:
                    print(f"  ID {r['id']}: {r['word']} @ {r['start_time']:.1f}s "
                          f"(chunk: {r['chunk_id']})")
                    if r['chunk_text']:
                        preview = r['chunk_text'][:60] + "..."
                        print(f"    In: '{preview}'")
            
            elif cmd[0] == "context" and len(cmd) > 1:
                word_id = int(cmd[1])
                ctx = q.get_word_context(word_id, 5)
                print(f"\nContext around word ID {word_id}:")
                for c in ctx:
                    marker = " >>> " if c['id'] == word_id else "     "
                    print(f"{marker} {c['word']:12} {c['start_time']:6.1f}s - {c['end_time']:6.1f}s")
            
            elif cmd[0] == "fillers":
                fillers = q.get_fillers()
                print(f"\nFiller words:")
                for f in fillers:
                    times = f['times'][:50] + "..." if len(f['times']) > 50 else f['times']
                    print(f"  '{f['word']}': {f['count']}x @ [{times}]")
            
            elif cmd[0] == "freq":
                freqs = q.get_word_frequency()
                print(f"\nTop words (excluding fillers/removed):")
                for f in freqs:
                    print(f"  {f['word_lower']:15} : {f['count']}x")
            
            elif cmd[0] == "removed":
                removed = q.get_removed_words()
                print(f"\nRemoved words ({len(removed)} total):")
                for r in removed:
                    print(f"  '{r['word']}' @ {r['start_time']:.1f}s")
            
            elif cmd[0] == "export" and len(cmd) > 1:
                filename = cmd[1]
                exclude = cmd[2:] if len(cmd) > 2 else []
                q.export_custom_srt(filename, exclude_words=exclude)
            
            else:
                print("Unknown command. Type 'exit' to quit.")
        
        except KeyboardInterrupt:
            print("\nUse 'exit' to quit.")
        except Exception as e:
            print(f"Error: {e}")
    
    q.close()


if __name__ == "__main__":
    if len(sys.argv) > 1 and sys.argv[1] == "--cli":
        interactive_query()
    else:
        # Run example queries
        q = TranscriptQuery()
        
        print("=== TRANSCRIPT DATABASE QUERIES ===\n")
        
        # 1. Search for spider
        print("1. SEARCH: 'spider'")
        results = q.search("spider")
        for r in results[:3]:
            print(f"   Found: '{r['word']}' at {r['start_time']:.2f}s")
            ctx = q.get_word_context(r['id'], 5)
            ctx_text = ' '.join([c['word'] for c in ctx])
            print(f"   Context: '...{ctx_text}...'")
        
        # 2. Top fillers
        print("\n2. FILLER WORDS:")
        fillers = q.get_fillers()
        for f in fillers[:5]:
            print(f"   '{f['word']}': {f['count']} times")
        
        # 3. Most common words
        print("\n3. MOST COMMON WORDS (non-filler):")
        freqs = q.get_word_frequency()
        for f in freqs[:10]:
            print(f"   '{f['word_lower']}': {f['count']}x")
        
        # 4. Custom export example
        print("\n4. CUSTOM EXPORT:")
        print("   Exporting without 'like', 'um', 'uh'")
        q.export_custom_srt("custom_no_fillers.srt", 
                           exclude_words=["like", "um", "uh"],
                           max_segment_duration=10.0)
        
        q.close()
        
        print("\n✓ Done! Run with --cli for interactive mode:")
        print("  python3 query_transcript.py --cli")
