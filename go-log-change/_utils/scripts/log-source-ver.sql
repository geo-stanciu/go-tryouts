with msg AS (
    select audit_log_id as id,
           substr(source, 1, position('/' in source) - 1) source,
           substr(source, position('/' in source) + 1) source_version
      from audit_log
     where position('/' in source) > 0
)
update audit_log a
   set source = m.source,
       source_version = m.source_version
  from msg m
 where a.audit_log_id = m.id;
