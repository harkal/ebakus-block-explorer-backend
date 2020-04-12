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
CREATE INDEX producer_idx ON blocks USING btree (producer);
CREATE INDEX timestamp_idx ON blocks USING btree (timestamp);

CREATE TABLE transactions (
  hash bytea PRIMARY KEY,
  nonce BIGINT,
  block_hash bytea,
  block_number BIGINT,
  tx_index BIGINT,
  addr_from bytea,
  addr_to bytea,
  value BIGINT,
  gas_limit BIGINT,
  gas_used BIGINT,
  cumulative_gas_used BIGINT,
  gas_price BIGINT,
  contract_address bytea,
  input bytea,
  status BIGINT,
  work_nonce BIGINT,
  timestamp BIGINT
);

CREATE INDEX txfrom_idx ON transactions USING btree (addr_from);
CREATE INDEX txto_idx ON transactions USING btree (addr_to);
CREATE INDEX txblock_number_idx ON transactions USING btree (block_number);
CREATE INDEX txtimestamp_idx ON transactions USING btree (timestamp);

CREATE TABLE balances (
  address bytea PRIMARY KEY,
  amount BIGINT
);

CREATE INDEX bl_amount_idx ON balances USING btree (amount);
