# Happy Eyeballs v2 Implementation

This package implements Happy Eyeballs v2 ([RFC 8305](https://tools.ietf.org/html/rfc8305)) for the letsdane project. Happy Eyeballs v2 reduces connection establishment delays in dual-stack environments by intelligently attempting connections to both IPv4 and IPv6 addresses.

## Architecture

The implementation consists of several components:

### Configuration (`config.go`)

Loads configuration from environment variables with sensible defaults following RFC 8305 recommendations:

- **Resolution Delay**: 50ms (time to wait before starting IPv4 DNS lookup after IPv6)
- **Connection Delay**: 250ms (time between connection attempts)
- **Metrics**: Enabled by default when Happy Eyeballs is enabled
- **Verbose Logging**: Disabled by default

### Address Sorting (`sorter.go`)

Implements address family sorting and interleaving:

- Separates addresses into IPv6 and IPv4 groups
- Sorts addresses to prefer IPv6 (RFC 8305 recommendation)
- Interleaves addresses to alternate between address families
- Ensures IPv6 gets the first connection attempt when available

### DNS Resolution (`resolver.go`)

Implements concurrent DNS lookups with configurable delay:

- Starts IPv6 (AAAA) lookup immediately
- Delays IPv4 (A) lookup by the configured resolution delay
- Collects results from both lookups
- Returns combined address list while maintaining security flags
- Integrates with metrics collection

### Connection Dialing (`dialer.go`)

Implements the connection racing algorithm:

- Attempts connections to addresses in interleaved order
- Staggers connection attempts with configurable delays
- Cancels remaining attempts once first connection succeeds
- Supports both plain TCP and TLS connections
- Handles context cancellation properly
- Cleans up failed connections to prevent resource leaks

### Metrics Collection (`metrics.go`)

Tracks connection performance:

- DNS resolution timing for IPv4 and IPv6
- Connection attempt success/failure rates
- Winning connection by address family
- Performance statistics for analysis
- In-memory metrics with log output

## Integration with letsdane

The Happy Eyeballs implementation integrates with letsdane's existing dialer:

1. **Initialization**: Config is loaded from environment variables in `newDialer()`
2. **DNS Resolution**: `resolveAddr()` and `resolveDANE()` use concurrent DNS lookups when enabled
3. **Connection Attempts**: `dialAddrList()` and `dialTLSContext()` use Happy Eyeballs algorithm when enabled
4. **Fallback**: If Happy Eyeballs is disabled or fails, falls back to sequential connection attempts
5. **Security**: DANE and DNSSEC validation are preserved throughout

## Security Considerations

The Happy Eyeballs implementation maintains letsdane's security guarantees:

- DNSSEC validation occurs for both IPv4 and IPv6 lookups
- DANE TLSA records are validated before connections
- Only the winning connection is used; others are closed immediately
- No security downgrade attacks are possible
- Certificate validation happens on the established connection

## Testing

The package includes comprehensive tests:

- **config_test.go**: Tests environment variable parsing and validation
- **sorter_test.go**: Tests address sorting and interleaving logic
- **resolver_test.go**: Tests concurrent DNS resolution with delays
- **dialer_test.go**: Tests connection racing with various scenarios

Tests cover:
- Single and multiple address scenarios
- IPv4-only, IPv6-only, and dual-stack environments
- Connection failures and timeouts
- Context cancellation
- Metrics collection

## Environment Variables

### Core Settings

- `LETSDANE_HAPPY_EYEBALLS` (bool, default: false)
  - Master switch to enable/disable Happy Eyeballs

### Timing Configuration

- `LETSDANE_HE_RESOLUTION_DELAY` (milliseconds, default: 50)
  - Time to wait before starting IPv4 DNS lookup after IPv6
  - Lower values reduce IPv6 preference
  - Higher values increase IPv6 preference

- `LETSDANE_HE_CONNECTION_DELAY` (milliseconds, default: 250, min: 100, max: 2000)
  - Time between connection attempts
  - Lower values are more aggressive
  - Higher values reduce network load

### Metrics and Logging

- `LETSDANE_HE_METRICS` (bool, default: true when Happy Eyeballs enabled)
  - Enable/disable metrics collection and logging

- `LETSDANE_HE_VERBOSE` (bool, default: false)
  - Enable verbose debugging output for troubleshooting

## Performance Impact

Happy Eyeballs provides several benefits:

1. **Reduced Latency**: Faster connections in dual-stack environments
2. **IPv6 Preference**: Encourages IPv6 adoption while ensuring connectivity
3. **Fault Tolerance**: Automatically falls back to working address family
4. **Efficient**: Minimal overhead when disabled or in single-stack environments

## References

- [RFC 8305: Happy Eyeballs Version 2](https://tools.ietf.org/html/rfc8305)
- [RFC 6555: Happy Eyeballs (original)](https://tools.ietf.org/html/rfc6555)
- [RFC 6698: DANE](https://tools.ietf.org/html/rfc6698)
- [RFC 7671: DANE Operations](https://tools.ietf.org/html/rfc7671)
