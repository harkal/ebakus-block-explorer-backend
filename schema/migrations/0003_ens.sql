
CREATE TABLE ens (
  hash bytea PRIMARY KEY,
  address bytea,
  name VARCHAR(64)
);

CREATE INDEX ens_address_idx ON ens USING btree (address);

CREATE INDEX txcontract_address_idx ON transactions USING btree (contract_address);
