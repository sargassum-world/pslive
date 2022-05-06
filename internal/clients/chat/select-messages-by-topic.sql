select * from (
  select
    topic as topic,
    send_time as send_time,
    sender_id as sender_id,
    sender_identifier as sender_identifier,
    body as body
  from chat_messages as c
  where
    c.topic = $topic
    -- TODO: add pagination
  order by send_time desc
  limit $rows_limit
)
order by send_time asc;
