-- Note: we can safely relax durability on power-loss in exchange for reducing fsync() calls,
-- in order to increase write performance. Durability on application crash is not affected.
pragma synchronous = normal;
