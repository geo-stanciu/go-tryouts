create trigger insert_audit_log_trigger
    before insert on wmeter.audit_log
    for each row execute procedure wmeter.audit_log_partition();

create trigger user_password_trigger
    before insert on wmeter.user_password
    for each row execute procedure wmeter.user_password_partition();

create trigger user_trigger
    before insert on wmeter.user
    for each row execute procedure wmeter.user_partition();

create trigger user_role_trigger
    before insert on wmeter.user_role
    for each row execute procedure wmeter.user_role_partition();

create trigger user_role_history_trigger
    before insert on wmeter.user_role_history
    for each row execute procedure wmeter.user_role_history_partition();
