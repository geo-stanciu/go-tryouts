create or replace function exchange_rate_partition() returns trigger as
$$
declare
    _schema    varchar(64) := 'wmeter';
    _table     varchar(64) := 'exchange_rate';
    _year      varchar(4);
    _tablePartition varchar(64);
begin
    _year := to_char(NEW.exchange_date, 'yyyy');
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
            CHECK (exchange_date >= DATE %L AND exchange_date <= DATE %L),
            CONSTRAINT %I PRIMARY KEY (currency_id, exchange_date),
            CONSTRAINT %I foreign key (currency_id)
                references %s.currency (currency_id)
            ) INHERITS (%I)',
            _schema,
            _tablePartition,
            _year || '-01-01',
            _year || '-12-31',
            _tablePartition || '_pk',
            _tablePartition || '_currency_fk',
            _schema,
            _table);
            
        RAISE NOTICE 'A partition has been created %', _tablePartition;
    end if;
    
    -- Insert the current record into the correct partition, which we are sure will now exist.
    EXECUTE format('INSERT INTO %s.%I VALUES ($1.*)', _schema, _tablePartition)
      USING NEW;
      
    RETURN NULL;
end;
$$
language plpgsql;
