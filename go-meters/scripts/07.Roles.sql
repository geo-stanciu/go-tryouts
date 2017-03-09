DO $$
declare
    _found boolean;
    
    _r varchar;
    arr  varchar[] := array[
        'Administrator'
    ];
begin
    
    FOREACH _r IN ARRAY arr
    LOOP
        select exists(
            select 1
              from wmeter.role
             where lower(role) = lower(_r)
        ) into _found;
        
        if _found = FALSE then
            insert into wmeter.role (
                role
            )
            values (
                _r
            );
        end if;
    
    END LOOP;
    
END$$;
