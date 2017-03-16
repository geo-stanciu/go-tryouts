DO $$
declare
    _found boolean;
    
    p varchar[];
    arr  varchar[] := array[
        [ 'password-fail-interval', '10' ],
        [ 'max-allowed-failed-atmpts', '3' ]
    ];
begin
    
    FOREACH p SLICE 1 IN ARRAY arr
    LOOP
        select exists(
            select 1
              from wmeter.system_params
             where param = p[1]
        ) into _found;
        
        if _found = FALSE then
            insert into wmeter.system_params (
                param,
                val
            )
            values (
                p[1],
                p[2]
            );
        end if;
    
    END LOOP;
    
END$$;
