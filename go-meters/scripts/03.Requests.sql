DO $$
declare
    lContor int;
    
    p varchar[];
    arr  varchar[] := array[
        [ 'login', 'Home', 'Login', 'index', 'login' ]
    ];
begin
    
    FOREACH p SLICE 1 IN ARRAY arr
    LOOP
        select count(*)
          into lContor
          from wmeter.request
         where request_url = p[1];
        
        if lContor = 0 then
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
