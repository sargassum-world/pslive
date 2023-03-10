-- Camera

alter table instruments_camera
add enabled integer; -- used as boolean

update instruments_camera
set enabled = true;

-- Controller

alter table instruments_controller
add enabled integer; -- used as boolean

update instruments_controller
set enabled = true;
