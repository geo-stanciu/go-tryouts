DO $$
declare
	lContor int;
	
	p varchar[];
	arr  varchar[] := array[
		[ 'Index', 'index.html', 'Home', 'Index', 'index' ],
		[ 'About', 'about.html', 'Home', 'About', 'about' ]
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
