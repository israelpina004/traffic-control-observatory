import { LineChart, Line, XAxis, YAxis, Tooltip, ResponsiveContainer, Legend } from 'recharts'
import type { BackendStats } from '../types/types'

export default function RequestsGraph({ data } : { data: BackendStats[] }) {
    if (!data) {
        return null
    }

    const chartData = []

    for (let i = 1; i < data.length; i++) {
        const prev = data[i - 1]
        const curr = data[i]

        const dataPoint:any = { time: i }

        for (const serverId in curr) {
            let prevCount = 0
            if (prev[serverId]) {
                prevCount = prev[serverId].Routed
            }
            let currCount = 0
            if (curr[serverId]) {
                currCount = curr[serverId].Routed
            }
            const diff = currCount - prevCount

            dataPoint[serverId] = diff
        }

        chartData.push(dataPoint)
    }

    const strokeColors = [
        "red",
        "green",
        "blue",
        "yellow",
        "purple",
        "orange",
        "pink",
        "cyan",
        "magenta",
        "brown",
        "gray",
        "black",
        "white"
    ]

    return (
        <>
            <div className="text-center text-xl font-bold">
                Change in Number of Requests Per Second by Backend Server
            </div>
            <ResponsiveContainer width="100%" height={300}>
                <LineChart data={chartData}>
                    <XAxis dataKey="time" />
                    <YAxis />
                    <Tooltip />
                    <Legend />
                    {
                        data.length > 0 && Object.keys(data[0]).map((serverId) => (
                            <Line
                                type="monotone"
                                dataKey={serverId}
                                strokeWidth={2}
                                stroke={strokeColors[parseInt(serverId) % strokeColors.length]}
                                dot={false}
                            />
                        ))
                    }
                </LineChart>
            </ResponsiveContainer>
        </>
    )
}