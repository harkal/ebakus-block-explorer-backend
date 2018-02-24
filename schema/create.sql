CREATE TABLE blocks (
  number BIGINT PRIMARY KEY, 
  timestamp BIGINT, 
  hash bytea, 
  parent_hash bytea, 
  state_root bytea, 
  transactions_root bytea, 
  receipts_root bytea,
  size INT,
  gas_used BIGINT,
  gas_limit BIGINT
);

CREATE INDEX CONCURRENTLY blockhash_idx ON blocks USING btree (hash);

CREATE TABLE transactions (
  hash bytea PRIMARY KEY,
  nonce BIGINT,
  block_hash bytea,
  block_number BIGINT,
  tx_index BIGINT,
  addr_from bytea,
  addr_to bytea,
  value BIGINT,
  gas_price BIGINT,
  gas BIGINT
);

CREATE INDEX CONCURRENTLY txfrom_idx ON transactions USING btree (addr_from);
CREATE INDEX CONCURRENTLY txto_idx ON transactions USING btree (addr_to);
