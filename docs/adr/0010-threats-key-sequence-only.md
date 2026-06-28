# `threats:` is a threat list only when it is a YAML sequence

The loader treats a `threats:` key as a list of threats only if its value is a YAML *sequence*. A *mapping* under the same key is ignored — it belongs to company-vocabulary files that reuse the word. This lets one key serve both a threat model and an unrelated vocabulary file without a separate key, but it is genuinely surprising: a `threats:` accidentally written as a mapping silently contributes zero threats.

Status: accepted wart — the alternative (two different keys) was judged worse than the disambiguation rule.
