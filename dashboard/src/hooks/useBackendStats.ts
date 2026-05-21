// src/hooks/useBackendStats.ts

import { useState, useEffect, useRef } from 'react'
import type { BackendStats } from '../types/types'


const API_URL = "" // TODO: Edit after deployment of backend to Render
const NUM_SERVERS = 10

export function useBackendStats() {
    const [stats, setStats] = useState<BackendStats | null>(null)
    const [history, setHistory] = useState<BackendStats[] | []>([])
    const baselineRequests = useRef<BackendStats | null>(null)

    useEffect(() => {
        // Fetch backend stats from the API
        const fetchStats = async () => {
            try {
                const response = await fetch(API_URL)
                const data = await response.json() as BackendStats

                if (!baselineRequests.current) {
                    console.log("Setting baseline")
                    baselineRequests.current = JSON.parse(JSON.stringify(data))
                }

                const adjustedData: BackendStats = {}

                for (let i = 0; i < NUM_SERVERS; i++) {
                    const server = data[i]
                    const baselineServer = baselineRequests.current![i]

                    adjustedData[i] = {
                        ...server,
                        Completed: server.Completed - baselineServer.Completed,
                        Routed: server.Routed - baselineServer.Routed,
                        Errored: server.Errored - baselineServer.Errored
                    }
                }

                setStats(adjustedData)

                setHistory(prev => {
                    if (prev.length >= 10) {
                        return [...prev.slice(1), adjustedData]
                    } else {
                        return [...prev, adjustedData]
                    }
                })

            } catch (err) {
                console.error('Failed to fetch backend stats:', err)
            }
        }

        fetchStats()

        const interval = setInterval(fetchStats, 1000)

        return () => clearInterval(interval)
    }, [])

    return { stats, history }
}
