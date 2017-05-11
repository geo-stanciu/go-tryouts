create or replace function wmeter.audit_log_partition() returns trigger as
$$
declare
    _schema    varchar(64) := 'wmeter';
    _table     varchar(64) := 'audit_log';
    _year      varchar(4);
    _tablePartition varchar(64);
begin
    _year := to_char(NEW.log_time, 'yyyy');
    _tablePartition := _table || '_' || _year;
    
    -- Check if the partition needed for the current record exists
    PERFORM 1
       FROM pg_catalog.pg_class c
       JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
      WHERE c.relkind = 'r'
        AND c.relname = _tablePartition
        AND n.nspname = _schema;
        
    -- If the partition needed does not yet exist, then we create it:
    if NOT FOUND then
        EXECUTE format('CREATE TABLE %s.%I (
            PRIMARY KEY (audit_log_id),
            CHECK (log_time >= TIMESTAMP %L AND log_time <= TIMESTAMP %L)
            ) INHERITS (%s.%I)',
            _schema,
            _tablePartition,
            _year || '-01-01 00:00:00',
            _year || '-12-31 23:59:59',
            _schema,
            _table);
            
        EXECUTE format('CREATE INDEX %I ON %s.%I (log_time)',
            _tablePartition || '_LOG_TIME_IDX',
            _schema,
            _tablePartition);
            
        RAISE NOTICE 'A partition has been created %', _schema || '.' || _tablePartition;
    end if;
    
    -- Insert the current record into the correct partition, which we are sure will now exist.
    EXECUTE format('INSERT INTO %s.%I VALUES ($1.*)', _schema, _tablePartition)
      USING NEW;
      
    RETURN NULL;
end;
$$
language plpgsql;
