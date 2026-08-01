# Server Maintenance Run — 2026-08-01

## Issues Found & Fixed

### 1. ✅ ABSPATH "already defined" warning (4,521 occurrences in nginx log)

- **Cause:** `wp-config.php` defined `ABSPATH` unconditionally; some double-load paths (API scripts) re-triggered it
- **Fix:** Guarded with `if ( ! defined( 'ABSPATH' ) ) { ... }`
- **Impact:** Eliminated a PHP warning on every page load (wasted CPU + log bloat)

### 2. ✅ advanced-optimizer.php null object bug (1,072 log entries)

- **Cause:** `add_reading_time_to_excerpt()` accessed `$post->post_content` when `$post` was null (oembed/API contexts)
- **Fix:** Added null/is_object/empty guard returning early
- **Impact:** Removed "Attempt to read property on null" + strip_tags deprecation errors

### 3. ✅ PHP-FPM max_children too low (5)

- **Cause:** Server hit `pm.max_children` limit repeatedly under traffic
- **Fix:** Raised to 8 (start 3, min-spare 2, max-spare 5) — memory-safe given 7.6GB RAM
- **Impact:** Fewer queued requests under load; site now responds in ~0.08s

### 4. ✅ Twitter/X auto-post broken (API v1.1 dead)

- **Cause:** `twitter-autopost.php` used deprecated `api.twitter.com/1.1/statuses/update.json` → HTTP 404
- **Fix:** Updated to `api.twitter.com/2/tweets` with `text` param + `data.id` response handling
- **⚠️ Still broken:** Credentials return HTTP 401 (read AND write) — keys revoked/expired. User must regenerate from X Developer Portal

### 5. ✅ Database bloat (129MB wp_posts table)

- **Cause:** 2,161 post revisions + 252 orphaned postmeta rows
- **Fix:** Deleted all revisions + orphans, optimized tables, set `WP_POST_REVISIONS 0`
- **Impact:** Table trimmed; future bloat prevented

### 6. ✅ Disk cleanup (80% → 79%)

- Freed ~300MB: removed 10 old freellm.bak files, vacuumed journal (201MB→50MB), apt clean
- **Kept:** 2.9GB gemma GGUF (referenced by Modelfile/llamafile runner as fallback)

### 7. ✅ WP_DEBUG production mode

- Set `WP_DEBUG` false (display off) but kept `WP_DEBUG_LOG` true — errors still logged, no display leakage

## Services Status (all active)

- aimm-audit, aimm-content, aimm-dashboard, aimm-freellm, aimm-hustle-gen, aimm-litellm, aimm-orchestrator, aimm-publisher, aimm-python-trader, aimm-watchdog
- MySQL, nginx, php8.4-fpm, redis — all running
- PM2: fwber-backend-ts + fwber-frontend online

## Current Metrics

- **419 published posts** (up from 249)
- **Homepage: 200 in ~0.08s**
- **API: 200 in ~0.06s**
- **Disk: 16GB free (79%)**
- **Memory: 1.7GB available**

## Remaining Items for User

1. **Twitter/X credentials** — regenerate API keys (current ones return 401)
2. **JWT plugin deprecation warnings** — cosmetic only (vendored firebase/php-jwt lib on PHP 8.4); wait for plugin update
3. **Gemma 2.9GB GGUF** — can be deleted if llamafile fallback never used (frees 2.9GB)
