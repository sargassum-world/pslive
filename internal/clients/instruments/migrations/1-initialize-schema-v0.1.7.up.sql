-- Instrument

create table instruments_instrument (
  id            integer primary key,
  name          text    not null,
  description   text    not null,
  admin_user_id text    not null
) strict;

create index instruments_instrument_idx_id
on instruments_instrument (id);

create index instruments_instrument_idx_admin_user_id
on instruments_instrument (admin_user_id);

-- Camera

create table instruments_camera (
  id            integer primary key,
  instrument_id integer not null,
  url           text    not null,
  protocol      text    not null,
  constraint instruments_camera_fk_instrument_id
    foreign key(instrument_id)
      references instruments_instrument(id)
      on delete cascade
) strict;

create index instruments_camera_idx_instrument_id
on instruments_camera (instrument_id);

-- Controller

create table instruments_controller (
  id            integer primary key,
  instrument_id integer not null,
  url           text    not null,
  protocol      text    not null,
  constraint instruments_controller_fk_instrument_id
    foreign key(instrument_id)
      references instruments_instrument(id)
      on delete cascade
) strict;

create index instruments_controller_idx_instrument_id
on instruments_controller (instrument_id);

create index instruments_controller_idx_protocol
on instruments_controller (protocol);
