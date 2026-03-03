-- finalized blocks table
CREATE TABLE IF NOT EXISTS blocks (
  chain_id      TEXT    NOT NULL,
  number        BIGINT  NOT NULL,
  hash          TEXT    NOT NULL,
  parent_hash   TEXT    NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (chain_id, number)
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_blocks_chain_hash
  ON blocks(chain_id, hash);

-- listener state table
CREATE TABLE IF NOT EXISTS listener_state (
  chain_id             TEXT   PRIMARY KEY,
  last_finalized_block BIGINT NOT NULL DEFAULT 0,
  last_finalized_hash  TEXT   NOT NULL DEFAULT '',
  updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
