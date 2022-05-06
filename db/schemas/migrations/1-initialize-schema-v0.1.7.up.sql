create table chat_messages (
  message_id        integer primary key,
  topic             text    not null,
  send_time         integer not null,
  sender_id         text    not null,
  sender_identifier text    not null,
  body              text    not null
) strict;

create index chat_messages_idx_topics_time
on chat_messages (topic, send_time);
