select * from (
  select
    id             as id,
    topic          as topic,
    send_time      as send_time,
    sender_user_id as sender_id,
    body           as body
  from chat_message
  where
    chat_message.topic = $topic
    -- TODO: add pagination
  order by send_time desc
  limit $rows_limit
)
order by send_time asc;
