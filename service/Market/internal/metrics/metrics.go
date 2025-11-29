package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Business metrics
	OrdersCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "market_orders_created_total",
			Help: "Total number of orders created",
		},
	)

	ProductsViewedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "market_products_viewed_total",
			Help: "Total number of product views",
		},
	)

	CartItemsAddedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "market_cart_items_added_total",
			Help: "Total number of items added to cart",
		},
	)

	ActiveUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "market_active_users",
			Help: "Number of currently active users",
		},
	)

	// Redis metrics
	RedisHitsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "market_redis_hits_total",
			Help: "Total number of Redis cache hits",
		},
	)

	RedisMissesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "market_redis_misses_total",
			Help: "Total number of Redis cache misses",
		},
	)
)
