create trigger insert_exchange_rate_trigger
    before insert on exchange_rate
    for each row execute procedure exchange_rate_partition();
