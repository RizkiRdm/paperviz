# Facts — Chunk 6.1 Usage Measurement

1. `documents` table gets new column `processing_time_ms INTEGER` — set on pipeline completion
2. New `analytics_events` table: `(id, event_type TEXT, entity_id TEXT, metadata TEXT, created_at INTEGER)` — tracks comparison and share-referral events
3. `GET /api/analytics` endpoint returns: total_papers, total_figures, avg_processing_time_ms, returning_users, papers_per_user, total_shares, total_share_visits, total_share_conversions, total_comparisons, success_rate
4. Returning users = count of distinct non-null `user_id` from documents table
5. Papers per user = total_papers / returning_users (or 0 if no users)
6. Share events aggregated from `documents.share_visits` + `documents.share_conversions` + `charts.share_visits` + `charts.share_conversions`
7. Comparison events counted from `analytics_events` where `event_type = 'comparison'`
8. Success rate = count(status='complete') / count(*) from documents
9. Endpoint requires auth (admin/user only)
10. No personal data collected beyond existing user_id
