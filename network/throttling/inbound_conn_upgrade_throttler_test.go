// Copyright (C) 2019-2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package throttling

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/VidarSolutions/avalanchego/utils/ips"
	"github.com/VidarSolutions/avalanchego/utils/logging"
)

var (
	host1      = ips.IPPort{IP: net.IPv4(1, 2, 3, 4), Port: 9696}
	host2      = ips.IPPort{IP: net.IPv4(1, 2, 3, 5), Port: 9697}
	host3      = ips.IPPort{IP: net.IPv4(1, 2, 3, 6), Port: 9698}
	host4      = ips.IPPort{IP: net.IPv4(1, 2, 3, 7), Port: 9699}
	loopbackIP = ips.IPPort{IP: net.IPv4(127, 0, 0, 1), Port: 9699}
)

func TestNoInboundConnUpgradeThrottler(t *testing.T) {
	{
		throttler := NewInboundConnUpgradeThrottler(
			logging.NoLog{},
			InboundConnUpgradeThrottlerConfig{
				UpgradeCooldown:        0,
				MaxRecentConnsUpgraded: 5,
			},
		)
		// throttler should allow all
		for i := 0; i < 10; i++ {
			allow := throttler.ShouldUpgrade(host1)
			require.True(t, allow)
		}
	}
	{
		throttler := NewInboundConnUpgradeThrottler(
			logging.NoLog{},
			InboundConnUpgradeThrottlerConfig{
				UpgradeCooldown:        time.Second,
				MaxRecentConnsUpgraded: 0,
			},
		)
		// throttler should allow all
		for i := 0; i < 10; i++ {
			allow := throttler.ShouldUpgrade(host1)
			require.True(t, allow)
		}
	}
}

func TestInboundConnUpgradeThrottler(t *testing.T) {
	require := require.New(t)

	cooldown := 5 * time.Second
	throttlerIntf := NewInboundConnUpgradeThrottler(
		logging.NoLog{},
		InboundConnUpgradeThrottlerConfig{
			UpgradeCooldown:        cooldown,
			MaxRecentConnsUpgraded: 3,
		},
	)

	// Allow should always return true
	// when called with a given IP for the first time
	require.True(throttlerIntf.ShouldUpgrade(host1))
	require.True(throttlerIntf.ShouldUpgrade(host2))
	require.True(throttlerIntf.ShouldUpgrade(host3))

	// Shouldn't allow this IP because the number of connections
	// within the last [cooldown] is at [MaxRecentConns]
	require.False(throttlerIntf.ShouldUpgrade(host4))

	// Shouldn't allow these IPs again until [cooldown] has passed
	require.False(throttlerIntf.ShouldUpgrade(host1))
	require.False(throttlerIntf.ShouldUpgrade(host2))
	require.False(throttlerIntf.ShouldUpgrade(host3))

	// Local host should never be rate-limited
	require.True(throttlerIntf.ShouldUpgrade(loopbackIP))
	require.True(throttlerIntf.ShouldUpgrade(loopbackIP))
	require.True(throttlerIntf.ShouldUpgrade(loopbackIP))
	require.True(throttlerIntf.ShouldUpgrade(loopbackIP))
	require.True(throttlerIntf.ShouldUpgrade(loopbackIP))

	// Make sure [throttler.done] isn't closed
	throttler := throttlerIntf.(*inboundConnUpgradeThrottler)
	select {
	case <-throttler.done:
		t.Fatal("shouldn't be done")
	default:
	}

	throttler.Stop()

	// Make sure [throttler.done] is closed
	select {
	case _, chanOpen := <-throttler.done:
		require.False(chanOpen)
	default:
		t.Fatal("should be done")
	}
}
