drop view if exists v_rss;

create view v_rss as
select
	rs.source_name as source,
	r.title,
	convert_tz(r.rss_date, 'UTC', 'Europe/Bucharest') local_time,
	r.link,
	r.description
from
	rss r
join rss_source rs on
	(
		r.rss_source_id = rs.rss_source_id
	)
order by
	rss_date desc;
