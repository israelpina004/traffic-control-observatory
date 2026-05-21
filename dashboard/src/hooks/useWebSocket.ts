// src/hooks/useWebSocket.ts

import { useState, useEffect, useRef, useCallback } from 'react'
import type { TelemetryEvent } from '../types/types';

type ConnectionStatus = 'connected' | 'disconnected' | 'reconnecting'

const MAX_EVENTS = 5000

export function useWebSocket(url: string) {
    // State: what do consuming components need to read?
    //   - the array of events
    //   - the connection status
    const [events, setEvents] = useState<TelemetryEvent[]>([])
    const [latestBatch, setLatestBatch] = useState<TelemetryEvent[]>([])
    const [status, setStatus] = useState<ConnectionStatus>('disconnected')

    // Refs: what do you need to persist across renders 
    //       WITHOUT triggering re-renders?
    //   - the WebSocket instance itself
    //   - a reconnect timer ID (so you can cancel it on unmount)
    //   - a flag to indicate if the component has been unmounted
    const wsRef = useRef<WebSocket | null>(null)
    const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
    const isCleanedUpRef = useRef(false)

    // A useCallback for your connect function:
    //   - create WebSocket
    //   - set up onopen → update status
    //   - set up onmessage → parse JSON, append to events (bounded!)
    //   - set up onclose → update status, schedule reconnect
    //   - set up onerror → log it
    const connect = useCallback(() => {
        if (wsRef.current) {
            wsRef.current.close()
            wsRef.current = null
        }

        isCleanedUpRef.current = false

        const ws = new WebSocket(url)
        wsRef.current = ws

        ws.onopen = () => {
            setStatus('connected')
            console.log('WebSocket connected')
        }

        ws.onmessage = (e) => {
            try {
                const newEvents = JSON.parse(e.data) as TelemetryEvent[]
                setLatestBatch(newEvents)
                setEvents(prev => [...prev, ...newEvents].slice(-MAX_EVENTS))
            } catch (err) {
                console.error('Failed to parse telemetry events:', err)
            }
        }

        ws.onclose = () => {
            setStatus('disconnected')
            console.log('WebSocket disconnected')
            if (wsRef.current === ws && !isCleanedUpRef.current) {
                reconnectTimerRef.current = setTimeout(() => {
                    connect()
                }, 1000)
            }
        }

        ws.onerror = (err) => {
            console.error('WebSocket error:', err)
            ws.close()
        }
    }, [url])

    // A useEffect that:
    //   - calls connect() on mount
    //   - cleans up (closes socket, clears timers) on unmount
    useEffect(() => {
        connect()

        return () => {
            if (reconnectTimerRef.current) {
                clearTimeout(reconnectTimerRef.current)
            }
            if (wsRef.current) {
                wsRef.current.close()
            }
            isCleanedUpRef.current = true
        }
    }, [connect])

    // Return the state that components need
    return { events, latestBatch, status }
}
