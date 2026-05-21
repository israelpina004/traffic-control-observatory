import type { Particle } from "../types/types";

class Renderer {
    private ctx: CanvasRenderingContext2D
    private canvas: HTMLCanvasElement

    constructor(canvas: HTMLCanvasElement) {
        this.canvas = canvas
        this.ctx = canvas.getContext('2d')!
    }

    public draw(particles: Map<number, Particle>) {
        this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);

        // Draw servers
        for (let i = 0; i < 10; i++) {
            this.drawServer(i * 100 + 40, 650, `Server ${i}`);
        }

        for (const particle of particles.values()) {
            this.drawParticle(particle);
        }
    }

    private drawParticle(particle: Particle) {
        const particleColor = particle.color;
        const particleOpacity = particle.opacity;

        this.ctx.beginPath();
        this.ctx.arc(particle.posX, particle.posY, 5, 0, Math.PI * 2);
        this.ctx.fillStyle = particleColor;
        this.ctx.globalAlpha = particleOpacity;
        this.ctx.fill();
        this.ctx.globalAlpha = 1.0;
    }

    private drawServer(x: number, y: number, label: string) {
        this.ctx.beginPath();

        // Draw rectangle with rounded corners representing a server
        this.ctx.roundRect(x, y, 100, 50, 10);
        this.ctx.fillStyle = "#008535ff";
        this.ctx.fill();
        this.ctx.stroke();

        // Draw label
        this.ctx.font = "12px Arial";
        this.ctx.fillStyle = "#fff";
        this.ctx.textAlign = "center";
        this.ctx.textBaseline = "middle";
        this.ctx.fillText(label, x + 50, y + 25);
    }
}

export default Renderer;