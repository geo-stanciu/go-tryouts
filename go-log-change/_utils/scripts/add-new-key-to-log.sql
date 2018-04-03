DO $$
DECLARE

c_audit cursor for
select * from audit_log
where log_msg->>'msg_type' = 'get rss'
  and (log_msg->>'new_rss_items') is null
  for  update;
  
BEGIN

  for rec in c_audit loop
    update audit_log set log_msg = log_msg || jsonb '{"new_rss_items":-1}'
    where current of c_audit;
  end loop;

END$$;
