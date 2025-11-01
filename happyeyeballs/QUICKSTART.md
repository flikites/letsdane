# Happy Eyeballs v2 - Quick Start Guide

## For Users

### Enable Happy Eyeballs

```bash
export LETSDANE_HAPPY_EYEBALLS=true
letsdane -r 1.1.1.1
```

That's it! Happy Eyeballs is now active with sensible defaults.

### Common Configuration Examples

#### IPv6 Aggressive Mode (Prefer IPv6 more strongly)
```bash
export LETSDANE_HAPPY_EYEBALLS=true
export LETSDANE_HE_RESOLUTION_DELAY=100  # Wait longer for IPv6
letsdane -r 1.1.1.1
```

#### Fast Fallback Mode (Quick IPv4 fallback)
```bash
export LETSDANE_HAPPY_EYEBALLS=true
export LETSDANE_HE_RESOLUTION_DELAY=25   # Start IPv4 sooner
export LETSDANE_HE_CONNECTION_DELAY=150  # Faster connection attempts
letsdane -r 1.1.1.1
```

#### Silent Mode (No metrics logging)
```bash
export LETSDANE_HAPPY_EYEBALLS=true
export LETSDANE_HE_METRICS=false
letsdane -r 1.1.1.1
```

#### Debug Mode (Verbose output)
```bash
export LETSDANE_HAPPY_EYEBALLS=true
export LETSDANE_HE_VERBOSE=true
letsdane -r 1.1.1.1
```

#### With Database Metrics
```bash
export LETSDANE_HAPPY_EYEBALLS=true
export LETSDANE_HE_METRICS_DB=true
export SUPABASE_URL=https://your-project.supabase.co
export SUPABASE_ANON_KEY=your-anon-key
letsdane -r 1.1.1.1
```

## Understanding the Metrics

When Happy Eyeballs is enabled with metrics, you'll see output like:

```
[INFO] Happy Eyeballs: DNS IPv6 lookup for example.com completed in 15ms (1 addresses)
[INFO] Happy Eyeballs: DNS IPv4 lookup for example.com completed in 18ms (1 addresses)
[INFO] Happy Eyeballs: Connection to example.com (IPv6 2001:db8::1) succeeded in 45ms [WINNER]
[INFO] Happy Eyeballs: Connection to example.com (IPv4 192.0.2.1) failed after 250ms: connection timeout
```

This shows:
- IPv6 DNS lookup completed first
- IPv4 DNS lookup completed shortly after
- IPv6 connection succeeded and won the race
- IPv4 connection attempt timed out

## Performance Tuning

### Resolution Delay

**Default: 50ms** - Good balance for most networks

- **Increase (100-200ms)**: If you have fast IPv6 and want to avoid IPv4 almost entirely
- **Decrease (10-25ms)**: If IPv6 is often slow or broken in your environment

### Connection Delay

**Default: 250ms** - RFC 8305 recommendation

- **Increase (500-1000ms)**: To reduce network load, give more time to first attempt
- **Decrease (100-150ms)**: For faster fallback, more aggressive connection attempts

### When to Disable

Disable Happy Eyeballs if:
- You're on an IPv4-only or IPv6-only network (no benefit)
- You're experiencing connection issues (troubleshooting)
- You want to minimize CPU/memory usage (small overhead)

## Troubleshooting

### Happy Eyeballs Not Working?

1. **Check it's enabled:**
   ```bash
   echo $LETSDANE_HAPPY_EYEBALLS
   ```
   Should output: `true`

2. **Look for startup logs:**
   - With verbose mode: `export LETSDANE_HE_VERBOSE=true`
   - Check for "Happy Eyeballs" in log output

3. **Verify dual-stack environment:**
   - Run `curl -6 https://ipv6.google.com` (should work)
   - Run `curl -4 https://google.com` (should work)

### Metrics Not Showing?

1. Check metrics are enabled:
   ```bash
   echo $LETSDANE_HE_METRICS
   ```

2. Enable explicitly if needed:
   ```bash
   export LETSDANE_HE_METRICS=true
   ```

### Database Storage Not Working?

1. Verify environment variables:
   ```bash
   echo $SUPABASE_URL
   echo $SUPABASE_ANON_KEY
   ```

2. Check database tables exist:
   - Log into Supabase dashboard
   - Verify `he_dns_resolutions` and `he_connection_attempts` tables exist

3. Check logs for database errors:
   - Look for "[WARN] Failed to save..." messages

## For Developers

### Running Tests

```bash
# Test Happy Eyeballs package
go test ./happyeyeballs/... -v

# Test entire project
go test ./... -v

# Run with race detector
go test ./happyeyeballs/... -race

# Run benchmarks
go test ./happyeyeballs/... -bench=.
```

### Code Structure

```
happyeyeballs/
├── config.go          # Environment variable configuration
├── sorter.go          # Address sorting and interleaving
├── resolver.go        # Concurrent DNS lookups
├── dialer.go          # Connection racing algorithm
├── metrics.go         # Metrics collection
├── supabase.go        # Database storage
└── *_test.go          # Test files
```

### Integration Points

The Happy Eyeballs implementation integrates at these points:

1. **Dialer initialization** (`dialer.go:newDialer()`)
   - Loads config from environment
   - Creates metrics collector
   - Initializes Happy Eyeballs dialer

2. **DNS resolution** (`dialer.go:resolveAddr()`, `resolveDANE()`)
   - Uses concurrent DNS lookups when enabled
   - Falls back to sequential lookups when disabled

3. **Connection attempts** (`dialer.go:dialAddrList()`, `dialTLSContext()`)
   - Uses Happy Eyeballs racing when enabled
   - Falls back to sequential dialing when disabled

### Adding Features

To extend Happy Eyeballs:

1. **New configuration option:**
   - Add to `config.go:Config` struct
   - Parse in `LoadConfigFromEnv()`
   - Document in README

2. **New metric:**
   - Add to `metrics.go` structures
   - Update `RecordConnectionAttempt()` or `RecordDNSResolution()`
   - Update database schema if persisting

3. **Algorithm modification:**
   - Update `dialer.go` for connection logic
   - Update `resolver.go` for DNS logic
   - Update `sorter.go` for address ordering
   - Add tests in corresponding `*_test.go`

## FAQ

### Q: Does Happy Eyeballs work with DANE?
**A:** Yes! DANE validation is fully preserved. TLSA records are checked before connections, and certificate validation occurs on the winning connection.

### Q: Does it work with DNSSEC?
**A:** Yes! Both IPv4 and IPv6 DNS lookups maintain DNSSEC validation, and security flags are properly propagated.

### Q: What's the performance impact?
**A:** Minimal when disabled (single boolean check). When enabled, there's small memory overhead for metrics, but connection attempts are efficient with proper cleanup.

### Q: Can I use it in production?
**A:** Yes! The implementation is complete, tested, and production-ready. Start with default settings and tune as needed.

### Q: Will it slow down single-stack networks?
**A:** No! Happy Eyeballs detects single-stack scenarios and avoids unnecessary delays.

### Q: Can I disable metrics?
**A:** Yes! Set `LETSDANE_HE_METRICS=false` to disable metrics logging.

### Q: How do I analyze the database metrics?
**A:** Query the Supabase tables:

```sql
-- See IPv6 vs IPv4 win rates
SELECT
  family,
  COUNT(*) as attempts,
  SUM(CASE WHEN winner THEN 1 ELSE 0 END) as wins
FROM he_connection_attempts
GROUP BY family;

-- Average connection times by family
SELECT
  family,
  AVG(duration_ms) as avg_duration_ms
FROM he_connection_attempts
WHERE success = true
GROUP BY family;
```

## Support

For issues or questions:
1. Check the main README: `../README.md`
2. Check package documentation: `README.md`
3. Review implementation details: `../HAPPY_EYEBALLS_IMPLEMENTATION.md`
4. Check test files for usage examples: `*_test.go`
