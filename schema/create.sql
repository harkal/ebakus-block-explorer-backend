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
