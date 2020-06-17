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
CREATE INDEX txcontract_address_idx ON transactions USING btree (contract_address);
CREATE INDEX txblock_number_idx ON transactions USING btree (block_number);
CREATE INDEX txtimestamp_idx ON transactions USING btree (timestamp);

CREATE TABLE balances (
  address bytea PRIMARY KEY,
  liquid_amount BIGINT,
  staked_amount BIGINT,
  block_number BIGINT
);

CREATE INDEX bl_liquid_amount_idx ON balances USING btree (liquid_amount);
CREATE INDEX bl_staked_amount_idx ON balances USING btree (staked_amount);

CREATE TABLE ens (
  hash bytea PRIMARY KEY,
  address bytea,
  name VARCHAR(64)
);

CREATE INDEX ens_address_idx ON ens USING btree (address);

CREATE TABLE producers (
  address bytea PRIMARY KEY,
  produced_blocks_count INT,
  block_rewards BIGINT
);

CREATE TABLE globals (
  var_name CHAR(64) PRIMARY KEY,
  value_int BIGINT,
  value_str VARCHAR(64)
)
