
ALTER TABLE balances RENAME TO balances_old;

CREATE TABLE balances (
  address bytea PRIMARY KEY,
  liquid_amount BIGINT,
  staked_amount BIGINT,
  block_number BIGINT
);

CREATE INDEX bl_liquid_amount_idx ON balances USING btree (liquid_amount);
CREATE INDEX bl_staked_amount_idx ON balances USING btree (staked_amount);

INSERT INTO balances (address, block_number) SELECT address, block_number FROM balances_old;

DROP TABLE balances_old;
