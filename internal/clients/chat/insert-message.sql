insert into chat_message (topic, send_time, sender_user_id, body)
values ($topic, $send_time, $sender_id, $body);

select last_insert_rowid() as id;
