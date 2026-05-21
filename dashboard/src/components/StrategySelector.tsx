import { useState } from 'react'
// import { Tooltip } from 'react-tooltip'
import './StrategySelector.css'

const API_URL = "" // TODO: Edit after deployment of backend to Render

export default function StrategySelector() {
  const [strategy, setStrategy] = useState('')

  const handleStrategyChange = async (event: React.ChangeEvent<HTMLSelectElement>) => {
    const newStrategy = event.target.value
    setStrategy(newStrategy)

    await fetch(`${API_URL}/api/start`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ strategy: newStrategy })
    })
  }

  return (
    <>
      <div>
        {strategy ? <p>Current Strategy: {strategy}</p> : null}
        {
          strategy == "random" ?
            <p>
              The Random strategy distributes traffic to servers randomly.
            </p> :
            strategy == "round_robin" ?
              <p>
                The Round Robin strategy distributes traffic to servers in a circular manner, from server 0 to server 9 and then back around.
              </p> :
              strategy == "least_connections" ?
                <p>
                  The Least Connections strategy distributes traffic to servers with the fewest active connections.
                </p> :
                strategy == "peak_ewma" ?
                  <p>
                    The Peak EWMA strategy distributes traffic to servers based on their exponentially weighted moving average of connection counts
                    and latency, prioritizing servers with lower combined load.
                  </p> :
                  strategy == "p2c" ?
                    <p>
                      The Power of 2 Choices strategy selects two servers randomly and sends the request to the server with the least load, factoring
                      in each server's latency and active connections.
                    </p> : null
        }
        <label htmlFor="strategy">Select Load Balancing Algorithm:</label>
        <select className='strategy-selector' id="strategy" value={strategy} onChange={handleStrategyChange}>
          <option value=""></option>
          <option value="random">Random</option>
          <option value="round_robin">Round Robin</option>
          <option value="least_connections">Least Connections</option>
          <option value="peak_ewma">Peak EWMA</option>
          <option value="p2c">Power of 2 Choices</option>
        </select>
      </div>

      {/* <Tooltip id="random-tooltip" />
      <Tooltip id="round-robin-tooltip" />
      <Tooltip id="least-connections-tooltip" />
      <Tooltip id="peak-ewma-tooltip" />
      <Tooltip id="p2c-tooltip" /> */}
    </>
  )
}