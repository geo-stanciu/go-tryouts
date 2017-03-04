create trigger insert_audit_log_trigger
    before insert on wmeter.audit_log
    for each row execute procedure wmeter.audit_log_partition();
