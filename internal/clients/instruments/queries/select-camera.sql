select
  id            as id,
  instrument_id as instrument_id,
  url           as url,
  protocol      as protocol,
  enabled       as enabled
from instruments_camera as c
where
  c.id = $id
