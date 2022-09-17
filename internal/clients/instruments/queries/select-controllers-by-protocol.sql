select
  id            as id,
  instrument_id as instrument_id,
  url           as url,
  protocol      as protocol
from instruments_controller as c
where
  c.protocol = $protocol
