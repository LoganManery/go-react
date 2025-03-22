import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route} from 'react-router-dom'
import './index.css'
import App from './App.tsx'
import MazePage from './pages/MazePage.tsx'
import HomePage from './pages/HomePage.tsx'


createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<HomePage /> } />
        <Route path="/maze" element={ <MazePage /> } />
      </Routes>
    </BrowserRouter>
    <App />
  </StrictMode>,
)
