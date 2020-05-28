
CREATE TABLE producers (
  address bytea PRIMARY KEY,
  produced_blocks_count INT,
  block_rewards BIGINT
);
