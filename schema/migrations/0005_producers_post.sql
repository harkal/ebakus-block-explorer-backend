
-- Calculate block rewards for existing blocks in DB
--
-- IMPORTANT: while running this, keep the crawler stopped

INSERT INTO producers(address, produced_blocks_count, block_rewards)
SELECT producer, count(*) as produced_blocks_count, count(*) * 3171 as block_rewards
FROM blocks
WHERE producer IN (
  SELECT DISTINCT producer FROM blocks WHERE producer IS NOT NULL
)
GROUP BY producer;
