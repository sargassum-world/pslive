alter table instruments_instrument
rename column admin_user_id to admin_identity_id;

drop index instruments_instrument_idx_admin_user_id;

create index instruments_instrument_idx_admin_identity_id
on instruments_instrument (admin_identity_id);
