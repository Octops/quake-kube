package metrics

import (
	"github.com/criticalstack/quake-kube/internal/quake/net"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	activePlayers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "quake_active_players",
		Help: "The current number of active players",
	})

	scores = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "quake_player_scores",
		Help: "Current scores by player, by map",
	}, []string{"player", "map"})

	pings = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "quake_player_pings",
		Help: "Current ping by player",
	}, []string{"player"})

	configReloads = promauto.NewCounter(prometheus.CounterOpts{
		Name: "quake_config_reloads",
		Help: "Config file reload count",
	})
)

type StatusMetrics struct {
	MapName string
	Players []net.Player
}

func ConfigReloads() prometheus.Counter {
	return configReloads
}

func Pings() *prometheus.GaugeVec {
	return pings
}

func Scores() *prometheus.GaugeVec {
	return scores
}

func ActivePlayers() prometheus.Gauge {
	return activePlayers
}

func Report(metrics *StatusMetrics) {
	ActivePlayers().Set(float64(len(metrics.Players)))
	for _, p := range metrics.Players {
		if len(metrics.MapName) > 0 {
			Scores().WithLabelValues(p.Name, metrics.MapName).Set(float64(p.Score))
		}
		Pings().WithLabelValues(p.Name).Set(float64(p.Ping))
	}
}
