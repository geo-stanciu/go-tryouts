DO $$
declare
    _found boolean;
    
    p varchar[];
    arr  varchar[] := array[
        [ 'Index', 'home/index.html', 'Home', 'Index', 'index' ],
        [ 'About', 'home/about.html', 'Home', 'About', 'about' ],
        [ 'Login', 'home/login.html', 'Home', 'Login', 'login' ],
        [ 'Logout', '-', 'Home', 'Logout', 'logout' ]
    ];
begin
    
    FOREACH p SLICE 1 IN ARRAY arr
    LOOP
        select exists(
            select *
              from wmeter.page
             where page_url = p[5]
        ) into _found;
        
        if _found = FALSE then
            insert into wmeter.page (
                page_title,
                page_template,
                controller,
                action,
                page_url
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
