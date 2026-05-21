// SPAWNED = 0 | ROUTED = 1 | COMPLETED = 2 | ERRORED = 3
type ParticleState = 0 | 1 | 2 | 3;

interface Particle {
    ReqID: number;
    BackendID: number;
    spawnedAt: number;
    color: string;
    opacity: number;
    posX: number;
    posY: number;
    velX: number;
    velY: number;
    state: ParticleState;
    targetX: number;
    targetY: number;
    diedAt: number | null;
}

interface TelemetryEvent {
    ReqID: number,
    Type: number,
    BackendID: number,
    Metric: number,
    SourceHash: number,
    Padding: number,
}

interface BackendInfo {
    Routed: number,
    Completed: number,
    Errored: number,
    Latency: number,
}

type BackendStats = Record<number, BackendInfo>

export type { Particle, ParticleState, TelemetryEvent, BackendInfo, BackendStats };