
CREATE TABLE ens (
  address bytea PRIMARY KEY,
  hash bytea,
  name VARCHAR(64)
);

CREATE INDEX ens_name_idx ON ens USING btree (name);
