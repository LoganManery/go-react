import { FC } from 'react'
import { Link } from 'react-router-dom'

const HomePage: FC = () => {
    return (
        <div className="container mx-auto p-4">
            <h1 className="text-3xl font-bold mb-6">Welcome to Maze Game</h1>
            <Link
                to="/maze"
                className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded"
        > Plau Maze Game</Link>
        </div>
    )
}

export default HomePage