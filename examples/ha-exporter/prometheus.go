package main

import (
	"fmt"
	"net/http"
)

// ServeMetrics writes Prometheus text-format metrics from the collector's latest snapshot.
func ServeMetrics(col *Collector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap := col.Snapshot()
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")

		fmt.Fprintf(w, "# HELP enphase_current_power_watts Current production in watts.\n")
		fmt.Fprintf(w, "# TYPE enphase_current_power_watts gauge\n")
		fmt.Fprintf(w, "enphase_current_power_watts %.2f\n", snap.CurrentPowerW)

		fmt.Fprintf(w, "# HELP enphase_energy_today_wh Energy produced today in watt-hours.\n")
		fmt.Fprintf(w, "# TYPE enphase_energy_today_wh gauge\n")
		fmt.Fprintf(w, "enphase_energy_today_wh %.2f\n", snap.EnergyTodayWh)

		fmt.Fprintf(w, "# HELP enphase_energy_lifetime_wh Lifetime energy production in watt-hours.\n")
		fmt.Fprintf(w, "# TYPE enphase_energy_lifetime_wh counter\n")
		fmt.Fprintf(w, "enphase_energy_lifetime_wh %.2f\n", snap.EnergyLifetimeWh)

		fmt.Fprintf(w, "# HELP enphase_net_power_watts Net power (production minus consumption) in watts.\n")
		fmt.Fprintf(w, "# TYPE enphase_net_power_watts gauge\n")
		fmt.Fprintf(w, "enphase_net_power_watts %.2f\n", snap.NetPowerW)

		if len(snap.InverterWatts) > 0 {
			fmt.Fprintf(w, "# HELP enphase_inverter_watts Per-inverter current production in watts.\n")
			fmt.Fprintf(w, "# TYPE enphase_inverter_watts gauge\n")
			for serial, watts := range snap.InverterWatts {
				fmt.Fprintf(w, "enphase_inverter_watts{serial=%q} %.2f\n", serial, watts)
			}
		}
	}
}
