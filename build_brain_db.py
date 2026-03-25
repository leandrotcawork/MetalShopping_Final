#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
import sys
import sqlite3
import yaml
import re
from pathlib import Path
from datetime import datetime

sys.stdout.reconfigure(encoding='utf-8')

def log(color, text):
    colors = {
        'cyan': '\033[36m',
        'green': '\033[32m',
        'yellow': '\033[33m',
        'blue': '\033[34m',
        'reset': '\033[0m',
    }
    print(f"{colors.get(color, '')}{text}{colors['reset']}")

def find_markdown_files(root_dir):
    md_files = []
    for dirpath, dirnames, filenames in os.walk(root_dir):
        for filename in filenames:
            if filename.endswith('.md'):
                md_files.append(os.path.join(dirpath, filename))
    return md_files

def parse_frontmatter(content):
    match = re.match(r'^---\n([\s\S]*?)\n---', content)
    if not match:
        return None
    try:
        return yaml.safe_load(match.group(1))
    except:
        return None

def extract_wikilinks(content):
    wikilinks = []
    for match in re.finditer(r'\[\[([\w\-\/\.]+)\]\]', content):
        wikilinks.append(match.group(1))
    return wikilinks

def initialize_database(db_path):
    conn = sqlite3.connect(db_path)
    c = conn.cursor()

    c.execute('''CREATE TABLE IF NOT EXISTS sinapses (
        id TEXT PRIMARY KEY,
        title TEXT NOT NULL,
        region TEXT NOT NULL,
        file_path TEXT NOT NULL,
        weight REAL DEFAULT 0.5,
        updated_at TEXT NOT NULL,
        last_accessed TEXT,
        severity TEXT,
        occurrence_count INTEGER DEFAULT 1
    )''')

    c.execute('''CREATE TABLE IF NOT EXISTS sinapse_tags (
        sinapse_id TEXT NOT NULL,
        tag TEXT NOT NULL,
        PRIMARY KEY (sinapse_id, tag),
        FOREIGN KEY (sinapse_id) REFERENCES sinapses(id) ON DELETE CASCADE
    )''')

    c.execute('''CREATE TABLE IF NOT EXISTS sinapse_links (
        source_id TEXT NOT NULL,
        target_id TEXT NOT NULL,
        PRIMARY KEY (source_id, target_id),
        FOREIGN KEY (source_id) REFERENCES sinapses(id) ON DELETE CASCADE
    )''')

    c.execute('CREATE INDEX IF NOT EXISTS idx_sinapses_region ON sinapses(region)')
    c.execute('CREATE INDEX IF NOT EXISTS idx_sinapses_weight ON sinapses(weight DESC)')

    conn.commit()
    return conn

def index_sinapses(brain_root, db_path):
    if os.path.exists(db_path):
        os.remove(db_path)
        log('yellow', '  ⟳ Removed existing brain.db')

    conn = initialize_database(db_path)
    c = conn.cursor()

    log('cyan', '📍 Scanning sinapses...')

    md_files = find_markdown_files(brain_root)
    indexed = 0
    errors = 0

    for file_path in sorted(md_files):
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()

            frontmatter = parse_frontmatter(content)
            if not frontmatter or 'id' not in frontmatter:
                rel_path = os.path.relpath(file_path, brain_root)
                log('yellow', f'  ⚠ Skipped {rel_path} (no id)')
                errors += 1
                continue

            sinapse_id = frontmatter['id']
            title = frontmatter.get('title', '')
            region = frontmatter.get('region', '')
            tags = frontmatter.get('tags', [])
            weight = frontmatter.get('weight', 0.5)
            updated_at = frontmatter.get('updated_at', datetime.now().isoformat())
            severity = frontmatter.get('severity')
            occurrence_count = frontmatter.get('occurrence_count', 1)

            rel_path = os.path.relpath(file_path, brain_root)

            c.execute('''INSERT OR REPLACE INTO sinapses
                (id, title, region, file_path, weight, updated_at, severity, occurrence_count)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?)''',
                (sinapse_id, title, region, rel_path, weight, updated_at, severity, occurrence_count))

            for tag in tags:
                c.execute('INSERT OR IGNORE INTO sinapse_tags (sinapse_id, tag) VALUES (?, ?)',
                    (sinapse_id, tag))

            wikilinks = extract_wikilinks(content)
            for link in wikilinks:
                c.execute('INSERT OR IGNORE INTO sinapse_links (source_id, target_id) VALUES (?, ?)',
                    (sinapse_id, link))

            log('green', f'  ✓ Indexed {sinapse_id}')
            indexed += 1

        except Exception as e:
            log('yellow', f'  ✗ Failed: {os.path.basename(file_path)} ({str(e)})')
            errors += 1

    conn.commit()

    c.execute('SELECT COUNT(*) as count FROM sinapses')
    total_count = c.fetchone()[0]

    conn.close()

    log('green', f'\n✅ Brain indexed!')
    log('blue', f'   Sinapses: {total_count}')
    log('blue', f'   DB: {db_path}')
    if errors > 0:
        log('yellow', f'   Errors: {errors}')

def main():
    log('cyan', '\n🧠 ForgeFlow Mini — Brain Indexer\n')

    project_path = os.getcwd()
    brain_root = os.path.join(project_path, '.brain')

    if not os.path.exists(brain_root):
        log('yellow', f'❌ Brain not found: {brain_root}')
        sys.exit(1)

    db_path = os.path.join(brain_root, 'brain.db')

    try:
        index_sinapses(brain_root, db_path)
    except Exception as e:
        log('yellow', f'\n❌ Error: {str(e)}')
        sys.exit(1)

if __name__ == '__main__':
    main()
