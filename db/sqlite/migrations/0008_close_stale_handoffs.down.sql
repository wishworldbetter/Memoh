-- 0008_close_stale_handoffs
-- The up migration is a one-time data cleanup. The original active states
-- of the affected handoffs are unrecoverable from this transformation
-- (we cannot tell pending from dispatched from running after the fact),
-- so the rollback is intentionally a no-op.
SELECT 1;
