DO $$
declare
    _found boolean;
    
    p varchar[];
    arr  varchar[] := array[
        -- pages
        -- request_title     request_template             request_url                controller   action              redirect_url       redirect_on_error
        [ 'Index',           'home/index.html',           'index',                   'Home',      'Index',            '-',               '-' ],
        [ 'About',           'home/about.html',           'about',                   'Home',      '-',                '-',               '-' ],
        [ 'Login',           'home/login.html',           'login',                   'Home',      '-',                '-',               '-' ],
        [ 'Register',        'home/register.html',        'register',                'Home',      '-',                '-',               '-' ],
        [ 'Change Password', 'home/change-password.html', 'change-password',         'Home',      '-',                '-',               '-' ],
        
        -- gets
        -- request_title     request_template             request_url                controller   action              redirect_url       redirect_on_error
        [ 'Logout',          '-',                         'logout',                  'Home',      'Logout',           '/',               '-' ],
        [ 'Exchange Rates',  '-',                         'exchange-rates',          'Home',      'GetExchangeRates', '-',               '-' ],
        
        -- posts
        -- request_title     request_template             request_url                controller   action              redirect_url       redirect_on_error
        [ 'Login',           '-',                         'perform-login',           'Home',      'Login',            'index',           'login' ],
        [ 'Logout',          '-',                         'perform-logout',          'Home',      'Logout',           'login',           'login' ],
        [ 'Register',        '-',                         'perform-register',        'Home',      'Register',         'login',           'register' ],
        [ 'Change Password', '-',                         'perform-change-password', 'Home',      'ChangePassword',   'change-password', 'change-password' ]
    ];
begin
    
    FOREACH p SLICE 1 IN ARRAY arr
    LOOP
        select exists(
            select 1
              from wmeter.request
            where request_url = p[3]
        ) into _found;
        
        if _found = FALSE then
            insert into wmeter.request (
                request_title,
                request_template,
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
                p[5],
                p[6],
                p[7]
            );
        end if;
    
    END LOOP;
    
END$$;
