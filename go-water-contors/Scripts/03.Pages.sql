DO $$
declare
	lContor int;
begin
	select count(*) into lContor from wmeter.page;
	
	if lContor = 0 then

		insert into wmeter.page (
			page_title,
			page_template,
			controller,
			action,
			page_url
		)
		values (
			'Index',
			'index.html',
			'Home',
			'Index',
			'index'
		);
		
		insert into wmeter.page (
			page_title,
			page_template,
			controller,
			action,
			page_url
		)
		values (
			'About',
			'about.html',
			'Home',
			'About',
			'about'
		);
	
	end if;
END$$;
