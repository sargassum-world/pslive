-- Automation Job

create table instruments_automation_job (
  id            integer primary key,
  instrument_id integer not null,
  enabled       integer not null default false,
  type          text    not null,
  specification text    not null,
  constraint instruments_automation_job_fk_instrument_id
    foreign key(instrument_id)
      references instruments_instrument(id)
      on delete cascade
) strict;

create index instruments_automation_job_idx_id
on instruments_automation_job (id);

create index instruments_automation_job_idx_instrument_id
on instruments_automation_job (instrument_id);
