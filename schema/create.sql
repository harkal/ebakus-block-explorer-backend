CREATE TABLE blocks (
  number BIGINT PRIMARY KEY,
  timestamp BIGINT,
  hash bytea,
  parent_hash bytea,
  transactions_root bytea,
  receipts_root bytea,
  size INT,
  transaction_count INT,
  gas_used BIGINT,
  gas_limit BIGINT,
  delegates bytea,
  producer bytea,
  signature bytea
);

CREATE INDEX blockhash_idx ON blocks USING btree (hash);

CREATE TABLE transactions (
  hash bytea PRIMARY KEY,
  nonce BIGINT,
  block_hash bytea,
  block_number BIGINT,
  tx_index BIGINT,
  addr_from bytea,
  addr_to bytea,
  value BIGINT,
  gasLimit BIGINT,
  gasUsed BIGINT,
  cumulativeGasUsed BIGINT,
  gasPrice BIGINT,
  input bytea,
  status BIGINT,
  workNonce BIGINT,
  timestamp BIGINT
);

CREATE INDEX txfrom_idx ON transactions USING btree (addr_from);
CREATE INDEX txto_idx ON transactions USING btree (addr_to);
CREATE INDEX txblock_hash_idx ON transactions USING btree (block_hash);
CREATE INDEX timestamp_idx ON transactions USING btree (timestamp);
