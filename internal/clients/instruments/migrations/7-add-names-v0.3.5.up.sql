-- Camera

alter table instruments_camera
add name text not null not null default "";

alter table instruments_camera
add description text not null not null default "";

create index instruments_camera_idx_name
on instruments_camera (name);

-- Controller

alter table instruments_controller
add name text not null default "";

alter table instruments_controller
add description text not null default "";

create index instruments_controller_idx_name
on instruments_controller (name);

-- Automation Job

alter table instruments_automation_job
add name text not null default "";

alter table instruments_automation_job
add description text not null default "";

create index instruments_automation_job_idx_name
on instruments_automation_job (name);
