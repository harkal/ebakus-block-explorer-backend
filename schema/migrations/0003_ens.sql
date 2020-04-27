
CREATE TABLE ens (
  hash bytea PRIMARY KEY,
  address bytea,
  name VARCHAR(64)
);

CREATE INDEX ens_address_idx ON ens USING btree (address);
