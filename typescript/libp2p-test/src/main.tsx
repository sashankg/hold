import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App.tsx'
import './index.css'
import './libp2p.ts'

if ('serviceWorker' in navigator) {
  navigator.serviceWorker.register(
    import.meta.env.MODE === 'production' ? '/service-worker.js' : '/dev-sw.js?dev-sw'
  )
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)
