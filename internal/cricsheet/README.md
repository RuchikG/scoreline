# Cricsheet Archive Notes

Scoreline caches Cricsheet JSON under the Scoreline cache directory and indexes
the extracted match files locally. Archive browsing does not call CricketData.org
and therefore does not spend live API quota.

The refresh command downloads:

```text
https://cricsheet.org/downloads/all_json.zip
```

This archive is currently about 100 MB, so the TUI only refreshes it when the
user explicitly presses `r` inside Cricket Archives. Normal archive entry only
scans the local cache and applies the user's format/team/competition/recent-day
filters.
