DO $$
declare
    lContor int;
    
    p varchar[];
    arr  varchar[] := array[
        [ 'Index', 'home/index.html', 'Home', 'Index', 'index' ],
        [ 'About', 'home/about.html', 'Home', 'About', 'about' ],
        [ 'Login', 'home/login.html', 'Home', 'Login', 'login' ]
    ];
begin
    
    FOREACH p SLICE 1 IN ARRAY arr
    LOOP
        select count(*)
          into lContor
          from wmeter.page
         where page_url = p[5];
        
        if lContor = 0 then
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
