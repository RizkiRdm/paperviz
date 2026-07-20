// Application entry point. Mounts the single top-level App component into
// the #root div declared in index.html. This file should stay this small —
// all real logic lives in App.jsx and the components it renders.
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.jsx'

createRoot(document.getElementById('root')).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
