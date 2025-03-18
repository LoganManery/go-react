import { useState, useEffect } from 'react'
import './App.css'

function App() {
  const [message, setMessage] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    // Fetch the hello message from the Go backend
    fetch('/api/hello')
      .then(response => response.json())
      .then(data => {
        setMessage(data.message)
        setLoading(false)
      })
      .catch(error => {
        console.error('Error fetching data:', error)
        setMessage('Error connecting to server')
        setLoading(false)
      })
  }, [])

  return (
    <div className="App">
      <div className="container">
        <h1>React + Vite + Go with Gin</h1>
        {loading ? (
          <p>Loading...</p>
        ) : (
          <div className="card">
            <p>Message from server: <strong>{message}</strong></p>
          </div>
        )}
      </div>
    </div>
  )
}

export default App
