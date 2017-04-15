DO $$
declare
    _found boolean;
    
    p varchar[];
    arr  varchar[] := array[
        [ 'password-rules', 'change-interval', '30' ],
        [ 'password-rules', 'password-fail-interval', '10' ],
        [ 'password-rules', 'max-allowed-failed-atmpts', '3' ],
        [ 'password-rules', 'not-repeat-last-x-passwords', '5' ],
        [ 'password-rules', 'min-characters', '8' ],
        [ 'password-rules', 'min-letters', '2' ],
        [ 'password-rules', 'min-capitals', '1' ],
        [ 'password-rules', 'min-digits', '1' ],
        [ 'password-rules', 'min-non-alpha-numerics', '1' ],
        [ 'password-rules', 'allow-repetitive-characters', '0' ],
        [ 'password-rules', 'can-contain-username', '0' ]
    ];
begin
    
    FOREACH p SLICE 1 IN ARRAY arr
    LOOP
        select exists(
            select 1
              from wmeter.system_params
             where param_group = p[1]
               and param       = p[2]
        ) into _found;
        
        if _found = FALSE then
            insert into wmeter.system_params (
                param_group,
                param,
                val
            )
            values (
                p[1],
                p[2],
                p[3]
            );
        end if;
    
    END LOOP;
    
END$$;
