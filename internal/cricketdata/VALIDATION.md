# CricketData API Validation

Validated documentation targets:

- Current matches/live scores: `https://api.cricapi.com/v1/currentMatches?apikey=[YOUR_API_KEY]&offset=0`
- Simple score feed: `https://api.cricapi.com/v1/cricScore?apikey=[YOUR_API_KEY]`

Implementation notes:

- The free live score guide states current matches are part of the free API and commonly limited to 100 hits per day.
- The current implementation starts with `currentMatches` only.
- Match scorecard/details endpoints should be validated with a real Scoreline API key before wiring selected-match polling.
- Captured fixtures must be sanitized and committed under `internal/cricketdata/testdata`.
- Never commit API keys.
