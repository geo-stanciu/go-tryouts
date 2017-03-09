create or replace function wmeter.user_role_partition() returns trigger as
$$
declare
    _schema         varchar(64)  := 'wmeter';
    _table          varchar(64)  := 'user_role';
    _offset         int;
    _low            varchar(32);
    _high           varchar(32);
    _max            int          := 100000;
    _tablePartition varchar(64);
begin
    _offset := NEW.user_role_id - MOD(NEW.user_role_id, _max);
    _low    := ltrim(to_char(_offset + 1, '999999999990'));
    _high   := ltrim(to_char(_offset + _max, '999999999990'));
    _tablePartition := _table || '_' || _low || '_' || _high;
    
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
            PRIMARY KEY (user_role_id),
            CHECK (user_role_id >= %s AND user_role_id <= %s),
            constraint  %I foreign key (role_id)
                references %s.role (role_id)
            ) INHERITS (%s.%I)',
            _schema,
            _tablePartition,
            _low,
            _high,
            _tablePartition || '_fk',
            _schema,
            _schema,
            _table);
            
        EXECUTE format('CREATE UNIQUE INDEX %I ON %s.%I (user_id, role_id)',
            _tablePartition || '_UK',
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
