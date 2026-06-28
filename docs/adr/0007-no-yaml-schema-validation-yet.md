# No YAML schema validation yet

The loader and `validate` enforce **referential** integrity (cross-references resolve, ids are unique within a kind) but there is no structural YAML *schema*. A schema is deferred until the file structures stabilise — adding one now would mean churning it alongside an evolving model for little gain, since referential checks already catch the errors that produce wrong reports.

Status: deferred.
