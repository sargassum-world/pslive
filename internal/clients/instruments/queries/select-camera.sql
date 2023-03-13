select
  id            as id,
  instrument_id as instrument_id,
  enabled       as enabled,
  name          as name,
  description   as description,
  protocol      as protocol,
  url           as url
from instruments_camera as c
where
  c.id = $id
