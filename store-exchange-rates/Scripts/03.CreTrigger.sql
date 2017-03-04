create trigger insert_exchange_rate_trigger
    before insert on exchange_rate
    for each row execute procedure exchange_rate_partition();

create trigger insert_audit_log_trigger
    before insert on audit_log
    for each row execute procedure audit_log_partition();
