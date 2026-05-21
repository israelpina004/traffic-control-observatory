import type { Particle, TelemetryEvent } from "../types/types";

class ParticleEngine {
    private particles: Map<number, Particle>;

    private readonly SERVER_X_POSITIONS: number[] = [
        40, 140, 240, 340, 440, 540, 640, 740, 840, 940
    ];

    private readonly SERVER_Y_POSITION: number = 650;

    private readonly FADE_OUT_DURATION: number = 2000;

    private readonly MAX_SPAWN_WAIT_TIME: number = 5000;

    private generateBrightBlueYellowHex = (minLightness: number = 40): string => {
        // Total safe degrees = 50 (Yellow/Orange) + 180 (Blue/Purple/Pink) = 230
        let hue = Math.floor(Math.random() * 230);

        // Distribute the random number into our two safe zones
        if (hue < 50) {
            // Map 0-49 to the 30°-79° range (Oranges to Yellows)
            hue += 30;
        } else {
            // Map 50-229 to the 150°-329° range (Cyans to Pinks)
            // Formula: (hue - 50) + 150 -> simplifies to hue + 100
            hue += 100;
        }

        // Calculate Saturation (0-100)
        const saturation = Math.floor(Math.random() * 100);

        // Calculate Lightness (avoiding black/darkness)
        const lightness = Math.floor(Math.random() * (100 - minLightness) + minLightness);

        return this.hslToHex(hue, saturation, lightness);
    };

    private hslToHex = (h: number, s: number, l: number): string => {
        l /= 100;
        const a = (s * Math.min(l, 1 - l)) / 100;
        const f = (n: number) => {
            const k = (n + h / 30) % 12;
            const color = l - a * Math.max(Math.min(k - 3, 9 - k, 1), -1);
            return Math.round(255 * color)
                .toString(16)
                .padStart(2, "0");
        };
        return `#${f(0)}${f(8)}${f(4)}`.toUpperCase();
    };

    constructor() {
        this.particles = new Map<number, Particle>();
    }

    public getParticles() {
        return this.particles;
    }

    private getBackendPosition(backendID: number): { x: number, y: number } {
        return {
            x: this.SERVER_X_POSITIONS[backendID],
            y: this.SERVER_Y_POSITION
        }
    }


    processEvents(events: TelemetryEvent[]) {
        for (const event of events) {
            switch (event.Type) {
                case 0: // SPAWNED
                    this.particles.set(event.ReqID, {
                        ReqID: event.ReqID,
                        BackendID: event.BackendID,
                        spawnedAt: performance.now(),
                        color: this.generateBrightBlueYellowHex(),
                        opacity: 1,
                        posX: Math.floor(Math.random() * 981) + 40,
                        posY: 40,
                        velX: 0,
                        velY: 0,
                        state: 0,
                        targetX: -1,
                        targetY: -1,
                        diedAt: null,
                    });
                    break;

                case 1: // ROUTED
                    const particle = this.particles.get(event.ReqID);
                    if (!particle) break;

                    particle.state = 1;
                    const backendPosition = this.getBackendPosition(event.BackendID);
                    const targetX = backendPosition.x + 10 + Math.floor(Math.random() * (81))
                    const targetY = backendPosition.y + 5 + Math.floor(Math.random() * (31))
                    particle.targetX = targetX;
                    particle.targetY = targetY;
                    break;

                case 2: // COMPLETED
                    const completedParticle = this.particles.get(event.ReqID);
                    if (!completedParticle) break;

                    completedParticle.state = 2;
                    // completedParticle.color = "#00ff66ff";
                    completedParticle.diedAt = performance.now();

                    // if (completedParticle.posX === -1 || completedParticle.posY === -1) {
                    //     completedParticle.posX = completedParticle.targetX;
                    //     completedParticle.posY = completedParticle.targetY;
                    // }

                    break;

                case 3: // ERRORED
                    const erroredParticle = this.particles.get(event.ReqID);
                    if (!erroredParticle) break;

                    erroredParticle.state = 3;
                    erroredParticle.color = "#ff0000ff";
                    erroredParticle.diedAt = performance.now();
                    break;

                default:
                    console.warn(`Unknown event type: ${event.Type}`);
                    break;
            }
        }
    }

    public update(deltaTime: number) {
        for (const [reqID, particle] of this.particles) {
            const targetX = particle.targetX;
            const targetY = particle.targetY;
            const particleX = particle.posX;
            const particleY = particle.posY;

            const dx = targetX - particleX;
            const dy = targetY - particleY;

            const dist = Math.sqrt(dx * dx + dy * dy);

            switch (particle.state) {
                case 0: // SPAWNED
                    if (performance.now() - particle.spawnedAt >= this.MAX_SPAWN_WAIT_TIME) {
                        this.particles.delete(reqID);
                    }
                    break;

                case 1: // ROUTED
                    if (performance.now() - particle.spawnedAt >= this.MAX_SPAWN_WAIT_TIME) {
                        this.particles.delete(reqID);
                    }

                    if (dist > 1) {
                        const velX = dx / dist;
                        const velY = dy / dist;

                        particle.posX += velX * deltaTime;
                        particle.posY += velY * deltaTime;
                    }

                    break;

                case 2: // COMPLETED
                    if (!particle.diedAt) {
                        particle.diedAt = performance.now();
                    }

                    if (particle.targetX !== -1 && particle.targetY !== -1) {
                        if (dist > 1) {
                            const velX = dx / dist;
                            const velY = dy / dist;

                            particle.posX += velX * deltaTime;
                            particle.posY += velY * deltaTime;
                        }
                    }

                    particle.opacity = 1 - (performance.now() - particle.diedAt) / this.FADE_OUT_DURATION;
                    if (particle.opacity < 0) {
                        this.particles.delete(reqID);
                    }
                    break;

                case 3: // ERRORED
                    if (!particle.diedAt) {
                        particle.diedAt = performance.now();
                    }

                    particle.opacity = 1 - (performance.now() - particle.diedAt) / this.FADE_OUT_DURATION;
                    if (particle.opacity < 0) {
                        this.particles.delete(reqID);
                    }
                    break;

                default:
                    console.warn(`Unknown particle state: ${particle.state}`);
                    break;
            }
        }
    }
}

export default ParticleEngine;