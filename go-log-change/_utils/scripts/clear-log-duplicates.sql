with msg as (
	select audit_msg->'old' as old_m, audit_msg->'new' as new_m, count(*) c1
	from audit_log
	where log_source like 'GoLogChange%'
	group by audit_msg->'old', audit_msg->'new'
	having count(*) > 1
),
msg2keep as (
	select m.old_m, m.new_m, max(audit_log_id) max_id
	from audit_log a, msg m
	where a.audit_msg->'old' = m.old_m
	and a.audit_msg->'new' = m.new_m
	and log_source like 'GoLogChange%'
	group by m.old_m, m.new_m
)
delete from audit_log a1
where a1.log_source like 'GoLogChange%'
and (a1.audit_msg->'old', a1.audit_msg->'new') in (select old_m, new_m from msg)
and a1.audit_log_id not in (select max_id from msg2keep mk);
