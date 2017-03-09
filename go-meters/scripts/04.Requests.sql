DO $$
declare
    _found boolean;
    
    p varchar[];
    arr  varchar[] := array[
        [ 'login', 'Home', 'Login', 'index', 'login' ],
        [ 'logout', 'Home', 'Logout', 'login', 'login' ],
        [ 'register', 'Home', 'Register', 'login', 'register' ]
    ];
begin
    
    FOREACH p SLICE 1 IN ARRAY arr
    LOOP
        select exists(
            select 1
              from wmeter.request
            where request_url = p[1]
        ) into _found;
        
        if _found = FALSE then
            insert into wmeter.request (
                request_url,
                controller,
                action,
                redirect_url,
                redirect_on_error
            )
            values (
                p[1],
                p[2],
                p[3],
                p[4],
                p[5]
            );
        end if;
    
    END LOOP;
    
END$$;
