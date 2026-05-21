import './App.css'
import { useWebSocket } from './hooks/useWebSocket'
import { useBackendStats } from './hooks/useBackendStats'
import RequestsGraph from './components/RequestsGraph'
import TrafficCanvas from './components/TrafficCanvas'
import BackendPanel from './components/BackendPanel'
import StrategySelector from './components/StrategySelector'
import { TrafficController } from './components/TrafficController'

function App() {
  const URL = "" // TODO: Edit after deployment of backend to Render
  const { latestBatch } = useWebSocket(URL);
  const { stats, history } = useBackendStats();

  return (
    <>
      <h1>Welcome to the Traffic Control Observatory 🔬</h1>
      <p>This is a system that simulates traffic across ten servers, and allows you to control the traffic using different algorithms.</p>
      <div className="display-layout">
        <StrategySelector />
        <TrafficController />
        <TrafficCanvas latestBatch={latestBatch} />
        <BackendPanel stats={stats} />
        <RequestsGraph data={history} />
      </div>
    </>
  )
}

export default App
