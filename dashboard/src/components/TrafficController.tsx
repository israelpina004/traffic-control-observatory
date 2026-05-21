import { useState } from "react"
import "./TrafficController.css"

const API_URL = "https://lb-simulator-backend.onrender.com" // TODO: Edit after deployment of backend to Render

export const TrafficController = () => {
    const [toggleWorkers, setToggleWorkers] = useState(false)

    const toggleWorkersHandler = async () => {
        const newState = !toggleWorkers
        setToggleWorkers(newState)

        await fetch(`${API_URL}/api/workers/toggle`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({ running: newState })
        })
    }

    return (
        <div className="traffic-controller">
            <button onClick={toggleWorkersHandler}>
                {toggleWorkers ? "Stop Workers" : "Start Workers"}
            </button>
        </div>
    )
}