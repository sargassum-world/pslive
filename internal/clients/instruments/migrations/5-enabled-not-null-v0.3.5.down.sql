-- Camera

alter table instruments_camera
rename column enabled to enabled_temp;

alter table instruments_camera
add enabled integer;

update instruments_camera
set enabled = enabled_temp;

alter table instruments_camera
drop column enabled_temp;

drop index instruments_camera_enabled;

-- Controller

alter table instruments_controller
rename column enabled to enabled_temp;

alter table instruments_controller
add enabled integer;

update instruments_controller
set enabled = enabled_temp;

alter table instruments_controller
drop column enabled_temp;

drop index instruments_controller_enabled;
