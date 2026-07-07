-- No-op: we cannot safely distinguish backfilled rows from legitimate
-- membership records that were subsequently modified, so rolling back
-- this migration would risk removing real project members.
-- If a rollback is truly required, restore from a database backup taken
-- prior to applying 000007.
SELECT 1;
