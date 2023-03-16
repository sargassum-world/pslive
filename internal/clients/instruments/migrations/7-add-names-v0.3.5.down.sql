-- Camera

alter table instruments_camera
drop column name;

alter table instruments_camera
drop column description;

drop index instruments_camera_idx_name;

-- Controller

alter table instruments_controller
drop column name;

alter table instruments_controller
drop column description;

drop index instruments_controller_idx_name;

-- Controller

alter table instruments_automation_job
drop column name;

alter table instruments_automation_job
drop column description;

drop index instruments_automation_job_idx_name;
