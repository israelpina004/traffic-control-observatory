import type { BackendStats } from "../types/types";
import "./BackendPanel.css"

export default function BackendPanel({ stats }: { stats: BackendStats | null }) {
    if (!stats) {
        return (
            <div className="panel-loading">
                Loading stats...
            </div>
        )
    }

    const backends = Object.entries(stats)

    return (
        <div className="backend-panel">
            <h3>Server Metrics</h3>
            <div className="backend-grid">
                {backends.map(([backendId, info]) => {
                    const latencyMs = (info.Latency / 1e6).toFixed(2);

                    return (
                        <div key={backendId} className="backend-card">
                            <div className="backend-header">
                                Server {backendId}
                            </div>
                            <div className="metric-row">
                                <span>Routed:</span>
                                <span>{info.Routed}</span>
                            </div>
                            <div className="metric-row" id="completed">
                                <span>Completed:</span>
                                <span>{info.Completed}</span>
                            </div>
                            <div className="metric-row" id="errored">
                                <span>Errored:</span>
                                <span>{info.Errored}</span>
                            </div>
                            <div className="metric-row">
                                <span>Latency:</span>
                                <span>{latencyMs}ms</span>
                            </div>
                        </div>
                    )
                })}
            </div>
        </div>
    )
}
    