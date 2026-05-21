import { useRef, useEffect } from "react";
import ParticleEngine from "../canvas/ParticleEngine";
import Renderer from "../canvas/Renderer";
import type { TelemetryEvent } from "../types/types";
import "./TrafficCanvas.css";

export default function TrafficCanvas({ latestBatch }: { latestBatch: TelemetryEvent[] }) {
    const canvasRef = useRef<HTMLCanvasElement>(null);
    const engineRef = useRef<ParticleEngine | null>(null);
    const rendererRef = useRef<Renderer | null>(null);
    const animFrameRef = useRef<number>(0);

    useEffect(() => {
        const canvas = canvasRef.current
        if (!canvas) return

        const engine = new ParticleEngine()
        engineRef.current = engine

        const renderer = new Renderer(canvas)
        rendererRef.current = renderer

        const observer = new ResizeObserver((entries) => {
            for (const entry of entries) {
                const { width, height } = entry.contentRect
                canvas.width = width
                canvas.height = height
                // engine.resize(width, height)
                // renderer.resize(width, height)
            }
        })

        observer.observe(canvas)
        canvas.width = canvas.clientWidth
        canvas.height = canvas.clientHeight

        let lastTime = 0;
        const loop = (currentTime: number) => {
            const deltaTime = currentTime - lastTime;
            if (lastTime != 0) {
                engine.update(deltaTime);
            }
            lastTime = currentTime;
            renderer.draw(engine.getParticles())
            animFrameRef.current = requestAnimationFrame(loop)
        }

        animFrameRef.current = requestAnimationFrame(loop)

        return () => {
            cancelAnimationFrame(animFrameRef.current)
            observer.disconnect()
        }
    }, [])

    useEffect(() => {
        const engine = engineRef.current
        if (!engine) return

        engine.processEvents(latestBatch)
    }, [latestBatch])

    return (
        <>
            <canvas ref={canvasRef} id="traffic-canvas" />
        </>
    )
}