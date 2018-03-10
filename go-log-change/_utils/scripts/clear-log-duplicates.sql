with msg as (
	select log_msg->'old' as old_m, log_msg->'new' as new_m, count(*) c1
	from audit_log
	where source like 'GoLogChange%'
	group by log_msg->'old', log_msg->'new'
	having count(*) > 1
),
msg2keep as (
	select m.old_m, m.new_m, max(audit_log_id) max_id
	from audit_log a, msg m
	where a.log_msg->'old' = m.old_m
	and a.log_msg->'new' = m.new_m
	and source like 'GoLogChange%'
	group by m.old_m, m.new_m
)
delete from audit_log a1
where a1.source like 'GoLogChange%'
and (a1.log_msg->'old', a1.log_msg->'new') in (select old_m, new_m from msg)
and a1.audit_log_id not in (select max_id from msg2keep mk);
