# HANDOFFS

Incoming patches from the frontend builder live here, named `TASK-###.patch`.

Header format (see HANDOFF.md):

```
---
task: TASK-001
owner: frontend
files: web/templates/..., web/static/...
summary: what changed
---
<unified diff>
```

Patches are never applied by AI agents. The human applies them after a PASS review in `REVIEWS/`.
