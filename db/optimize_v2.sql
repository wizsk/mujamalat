-- ============================
-- READ-ONLY PERFORMANCE PRAGMAS
-- ============================

PRAGMA journal_mode = OFF;
PRAGMA synchronous = OFF;
PRAGMA temp_store = MEMORY;
PRAGMA cache_size = -20000;       -- ~20MB cache
PRAGMA mmap_size = 300000000;     -- enable memory mapping if supported
PRAGMA locking_mode = EXCLUSIVE;


-- ============================
-- INDEXES FOR WORD LOOKUPS
-- ============================

CREATE INDEX IF NOT EXISTS idx_ghoni_word
ON mujamul_ghoni(word);

CREATE INDEX IF NOT EXISTS idx_ghoni_no_harakat
ON mujamul_ghoni(no_harakat);

CREATE INDEX IF NOT EXISTS idx_muashiroh_word
ON mujamul_muashiroh(word);

CREATE INDEX IF NOT EXISTS idx_wasith_word
ON mujamul_wasith(word);

CREATE INDEX IF NOT EXISTS idx_muhith_word
ON mujamul_muhith(word);

CREATE INDEX IF NOT EXISTS idx_shihah_word
ON mujamul_shihah(word);

CREATE INDEX IF NOT EXISTS idx_lisan_word
ON lisanularab(word);

CREATE INDEX IF NOT EXISTS idx_mufradat_word
ON mufradat_alfajul_quran(word);

CREATE INDEX IF NOT EXISTS idx_maqayeesul_word
ON maqayeesul_luga(word);


-- ============================
-- HIERARCHY INDEXES
-- ============================

CREATE INDEX IF NOT EXISTS idx_hanswehr_parent
ON hanswehr(parent_id);

CREATE INDEX IF NOT EXISTS idx_lane_parent
ON lanelexcon(parent_id);


-- ============================
-- HANSWEHR FULL TEXT SEARCH
-- ============================

CREATE VIRTUAL TABLE IF NOT EXISTS hanswehr_fts
USING fts5(
    word,
    meanings,
    content='hanswehr',
    content_rowid='id'
);

-- Run once after data import:
-- INSERT INTO hanswehr_fts(hanswehr_fts) VALUES('rebuild');


-- ============================
-- FINAL OPTIMIZATION PASS
-- ============================

ANALYZE;
PRAGMA optimize;
