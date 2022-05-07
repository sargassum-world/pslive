create table chat_message (
  id             integer primary key,
  topic          text    not null,
  send_time      integer not null,
  sender_user_id text    not null,
  body           text    not null
) strict;

create index chat_message_idx_topic_time
on chat_message (topic, send_time);
