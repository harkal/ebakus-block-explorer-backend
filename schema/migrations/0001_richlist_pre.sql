
CREATE INDEX txblock_number_idx ON transactions USING btree (block_number);

CREATE TABLE balances (
  address bytea PRIMARY KEY,
  amount BIGINT,
  block_number BIGINT
);

CREATE INDEX bl_amount_idx ON balances USING btree (amount);

CREATE TABLE globals (
  var_name CHAR(64) PRIMARY KEY,
  value_int BIGINT,
  value_str VARCHAR(64)
)
