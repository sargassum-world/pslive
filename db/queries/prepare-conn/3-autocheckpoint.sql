-- Note: we should only disable autocheckpointing if we're running Litestream and we have tens of
-- thousands of write transactions per second.
-- pragma wal_autocheckpoint = 0;
pragma wal_autocheckpoint = 1000;
