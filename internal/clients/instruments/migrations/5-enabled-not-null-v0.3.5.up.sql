-- Camera

update instruments_camera
set enabled = false
where enabled is null;

alter table instruments_camera
rename column enabled to enabled_temp;

alter table instruments_camera
add enabled integer not null default false;

update instruments_camera
set enabled = enabled_temp;

alter table instruments_camera
drop column enabled_temp;

create index instruments_camera_idx_enabled
on instruments_camera (enabled);

-- Controller

update instruments_controller
set enabled = false
where enabled is null;

alter table instruments_controller
rename column enabled to enabled_temp;

alter table instruments_controller
add enabled integer not null default false;

update instruments_controller
set enabled = enabled_temp;

alter table instruments_controller
drop column enabled_temp;

create index instruments_controller_idx_enabled
on instruments_controller (enabled);
