# Enable auto indexing

First, [deploy executors](https://docs.sourcegraph.com/admin/deploy_executors) targeting your instance. Then, add the following to your site-config:

```yaml
{
  "codeIntelAutoIndexing.enabled": true
}
```

Env vars (TODO):

- `PRECISE_CODE_INTEL_AUTO_INDEXING_TASK_INTERVAL`, "10m", "The frequency with which to run periodic codeintel auto-indexing tasks.
- `PRECISE_CODE_INTEL_AUTO_INDEXING_REPOSITORY_PROCESS_DELAY`, "24h", "The minimum frequency that the same repository can be considered for auto-index scheduling.
- `PRECISE_CODE_INTEL_AUTO_INDEXING_REPOSITORY_BATCH_SIZE`, "100", "The number of repositories to consider for auto-indexing scheduling at a time.
- `PRECISE_CODE_INTEL_DEPENDENCY_INDEXER_SCHEDULER_POLL_INTERVAL`, "1s", "Interval between queries to the dependency indexing job queue.
- `PRECISE_CODE_INTEL_DEPENDENCY_INDEXER_SCHEDULER_CONCURRENCY`, "1", "The maximum number of dependency graphs that can be processed concurrently.
