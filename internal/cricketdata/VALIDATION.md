# CricketData API Validation

Validated documentation targets:

- Current matches/live scores: `https://api.cricapi.com/v1/currentMatches?apikey=[YOUR_API_KEY]&offset=0`
- Selected match details: `https://api.cricapi.com/v1/match_info?apikey=[YOUR_API_KEY]&id=[MATCH_ID]`
- Simple score feed: `https://api.cricapi.com/v1/cricScore?apikey=[YOUR_API_KEY]`

Implementation notes:

- The free live score guide states current matches are part of the free API and commonly limited to 100 hits per day.
- The current implementation uses `currentMatches` for the dashboard and fetches `match_info` only for the selected match.
- List and detail responses are cached briefly and calls are rate-limited to avoid aggressive API usage.
- Captured fixtures must be sanitized and committed under `internal/cricketdata/testdata`.
- Never commit API keys.

Live verification:

- Verified on 2026-05-10 with a user-provided CricketData API key.
- `currentMatches` returned `status: success`, 25 rows at `offset=0`, and quota metadata showing a 100 hit daily limit.
- `match_info` returned `status: success` for the first current match ID.
- The sampled `match_info` response included match metadata, teams, venue, status, and toss fields. It had no score rows yet because the sampled match had not started scoring.
- The API key was not committed or written into source-controlled files.
