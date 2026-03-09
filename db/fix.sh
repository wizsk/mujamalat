#!/bin/bash
set -ex

DB="mujamalat.db"
TABLE="mujamul_ghoni"
COLUMN="no_harakat"

sqlite3 "$DB" <<EOF
UPDATE $TABLE
SET $COLUMN =
    REPLACE(
      REPLACE(
        REPLACE(
          REPLACE(
            REPLACE($COLUMN, '،', '_'),  -- replace comma
            '   ', '_'),                 -- triple spaces
          '  ', '_'),                    -- double spaces
        '____', '_'),                    -- collapse 4 underscores
      '___', '_'),                       -- collapse 3 underscores
    '__', '_');                           -- collapse 2 underscores
EOF
