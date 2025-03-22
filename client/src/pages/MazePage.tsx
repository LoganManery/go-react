import { FC } from 'react'
import Maze from '../Maze'

const MazePage: FC = () => {
    return (
        <div className="container mx-auto p-4">
            <h1 className="text-3xl fojnt-bold mb-4">Maze Game</h1>
            <Maze />
        </div>
    )
}

export default MazePage