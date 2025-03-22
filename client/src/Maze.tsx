import React, { useState, useEffect } from 'react'

enum CellType {
    WALL = 0,
    PATH = 1,
    START = 2,
    END = 3,
    SOLUTION = 4
}

// Define type for positions and maze
type Position = [number, number]
type MazeGrid = CellType[][]
type VisitedGrid = boolean[][]

interface NeighborData {
    nextX: number,
    nextY: number,
    wallX: number,
    wallY: number,
}

const MazeGenerator = () => {
    const [maze, setMaze] = useState<MazeGrid>([])
    const [size, setSize] = useState<number>(15)
    const [solution, setSolution] = useState<Position[]>([])
    const [showSolution, setShowSolution] = useState<boolean>(false)

    // Directions for maze generation and solving
    const directions: Position[] = [
        [0, 2],
        [2, 0],
        [0, 2],
        [-2, 0]
    ]

    const solvingDirections: Position[] = [
        [0, -1],
        [1, 0],
        [0, 1],
        [-1, 0],
    ]

    // Directions for maze solving (moving 1 cell at a time)
    useEffect(() => {
        generateMaze();
    }, [size])

    const generateMaze = (): void => {
        // Initialize maze with walls
        const newMaze: MazeGrid = Array(size).fill(null).map(() =>
            Array(size).fill(CellType.WALL) 
        )

        // Use recursive backtracking to generate the maze
        const stack: Position[] = []
        const startX: number = 1
        const startY: number = 1

        // Mark start position
        newMaze[startY][startX] = CellType.PATH
        stack.push([startX, startY])

        while (stack.length > 0) {
            const currentPosition: Position = stack[stack.length - 1]
            const currentX: number = currentPosition[0]
            const currentY: number = currentPosition[1]

            // Get unvisited neighbors
            const neighbors: NeighborData[] = []


            for (const [dx, dy] of directions) {
                const newX: number = currentX + dx
                const newY: number = currentY + dy

                if (
                    newX > 0 &&
                    newX < size - 1 &&
                    newY > 0 &&
                    newY < size - 1 &&
                    newMaze[newY][newX] === CellType.WALL
                ) {
                    neighbors.push({
                        nextX: newX,
                        nextY: newY,
                        wallX: currentX + dx / 2,
                        wallY: currentY + dy/2
                    })
                }
            }

            if (neighbors.length > 0) {
                // Choose a random unvisited neighbor
                const { nextX, nextY, wallX, wallY } = neighbors[Math.floor(Math.random() * neighbors.length)]

                // Carve a path
                newMaze[wallY][wallX] = CellType.PATH
                newMaze[nextY][nextX] = CellType.PATH

                // Add to stack
                stack.push([nextX, nextY])
            } else {
                stack.pop()
            }
        }

        // Set start and end points
        newMaze[1][1] = CellType.START
        newMaze[size - 2][size - 2] = CellType.END

        setMaze(newMaze)
        setSolution([])
        setShowSolution(false)
    }

    const solveMaze = (): void => {
        if (maze.length === 0) return

        const vistied: VisitedGrid = maze.map(row => row.map(() => false))
        const startPos: Position = [1, 1]
        const path: Position[] = []

        const findPath = (x: number, y: number, visited: VisitedGrid, path: Position[]): boolean => {
            if (
                x < 0 ||
                y < 0 ||
                x >= size ||
                y >= size ||
                maze[x][y] === CellType.WALL ||
                visited[y][x]
            ) {
                return false
            }

            // Add current position to path
            path.push([x, y])
            visited[y][x] = true

            if (maze[y][x] === CellType.END) {
                return true
            }

            for (const [dx, dy] of solvingDirections) {
                const newX: number = x + dx
                const newY: number = y + dy

                if (findPath(newX, newY, visited, path)) {
                    return true
                }
            }
            // If no direction workd, backtrack
            path.pop()
            return false
        }

        const found: boolean = findPath(startPos[0], startPos[1], vistied, path)

        if (found) {
            setSolution(path)
            setShowSolution(true)
        }
    }

    const getCellStyle = (cellType: CellType): React.CSSProperties => {
        switch(cellType) {
            case CellType.WALL:
                return { backgroundColor: '#333' }
            case CellType.PATH:
                return { backgroundColor: '#fff' }
            case CellType.START:
                return { backgroundColor: '#4CAF50'}
            case CellType.END:
                return { backgroundColor: '#F44336' }
            case CellType.SOLUTION:
                return { backgroundColor: '#2196F3' }
            default:
                return { backgroundColor: '#fff' }
        }
    }

    return (
    <div className="flex flex-col items-center gap-4 p-4">
        <h1 className="text-2xl font-bold mb-4">Solvable Maze Generator</h1>
        <div className="mb-4 flex items-center gap-4">
            <label className="flex items-center gap-2">
                Size:
                <select
                    value={size}
                    onChange={(e) => setSize(parseInt(e.target.value, 10))}
                    className="p-2 border rounded">
                        <option value="11">Small (11x11)</option>
                        <option value="15">Medium (15x15)</option>
                        <option value="21">Large (21x21)</option>
                        <option value="31">X-Large (31x31)</option>
                    </select>
            </label>
            <button
                onClick={generateMaze}
                className="px-4 py-2 bg-blue-500 text-white rounded"
                type="button"
            >
                New Maze
            </button>
            <button
                onClick={solveMaze}
                className="px-4 py-2 bg-green-500 text-white rounded"
                type="button"
            >
                Solve Maze
            </button>
        </div>
        <div
            className="border border-gray-300"
            style={{
                display: 'grid',
                gridTemplateColumns: `repeat(${size}, 20px)`,
                gap: 0
            }}
            >
                {maze.map((row, y) =>
                    row.map((cell, x) => {
                        // Check if this cell is part of the solution
                        const isSolution: boolean = showSolution &&
                        solution.some(([sx, sy]) => sx === x && sy === y) &&
                        cell != CellType.START && cell !== CellType.END
                        return (
                            <div
                                key={`${x}-${y}`}
                                style={{
                                    ...getCellStyle(isSolution ? CellType.SOLUTION : cell),
                                    width: '20px',
                                    height: '20px'
                                }}
                            />
                        )
                    })
                )}
            </div>
            <div className="mt-4 flex gap-4">
                <div className="flex items-center gap-2">
                    <div className="w-4 h-4 bg-green-500"></div>
                    <span>Start</span>
                </div>
                <div className="flex items-center gap-2">
                    <div className="w-4 h-4 bg-red-500"></div>
                    <span>End</span>
                </div>
                <div className="flex items-center gap-2">
                    <div className="w-4 h-4 bg-blue-500"></div>
                    <span>Solution</span>
                </div>
            </div>
        </div>
    )
}

export default MazeGenerator


