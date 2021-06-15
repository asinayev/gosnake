package main

import (
	"fmt"
  "math"
)

const MaxSize=21
const Maxi=1000

type StatusType int
const (
    Outside StatusType = iota
    Inside
    Snake
    Food
    Danger
    Hazard
)

type status struct {
  StatusName StatusType
  SnakeOrder int
}

type BoardRep struct {
  Board     [MaxSize][MaxSize]status
	Height    int 
	Width     int
  Offset    int
  OpenSlots int
  MyLength  int32
}

func GetDepth(slots int) int {
  if (slots>85){
    return 8
  } else if (slots>60){
    return 9
  } else {
    return 10
  }
}

func GetDirections() map[string][]int {
  return map[string][]int{
    "left":{-1, 0},
    "right":{ 1, 0},
    "down":{ 0,-1},
    "up":{ 0, 1},
  }
}

func CreateRepresentation(b Board, me Battlesnake) BoardRep {
  var BoardRepresentation BoardRep
  Offset:=2
  BoardRepresentation.Offset=Offset
  BoardRepresentation.Height=b.Height
  BoardRepresentation.Width=b.Width
  BoardRepresentation.MyLength=me.Length
  for x := 0; x < b.Width; x++ {
    for y := 0; y < b.Height; y++ {
      BoardRepresentation.Board[x+Offset][y+Offset].StatusName=Inside
      BoardRepresentation.OpenSlots+=1
    }
  }
  for _, xy := range b.Hazards {
    BoardRepresentation.Board[xy.X+Offset][xy.Y+Offset].StatusName=Hazard
  }
  for _, xy := range b.Food {
    BoardRepresentation.Board[xy.X+Offset][xy.Y+Offset].StatusName=Food
  }
  for _, snake := range b.Snakes {
    if ((snake.Length>=me.Length) && (snake.ID != me.ID)){
      dirmap := GetDirections()
      for _, directions := range dirmap {
        directionX := snake.Head.X+directions[0]+Offset
        directionY := snake.Head.Y+directions[1]+Offset
        BoardRepresentation.Board[directionX][directionY].StatusName = Danger
      }
    }
    for snakei, xy := range snake.Body {
      BoardRepresentation.OpenSlots-=1
      BoardRepresentation.Board[xy.X+Offset][xy.Y+Offset].StatusName=Snake
      BoardRepresentation.Board[xy.X+Offset][xy.Y+Offset].SnakeOrder=int(snake.Length)-snakei
    }

  }
  return BoardRepresentation
}

func DetermineValue(x int, y int, board BoardRep, i int) float64 {
  var status = board.Board[x+board.Offset][y+board.Offset]
  spotpoints := float64(50)
  if board.MyLength>9{
    spotpoints-= math.Abs(float64(x-(board.Width /2)))/float64(board.Width ) *3 -
                math.Abs(float64(y-(board.Height/2)))/float64(board.Height) *3
    if ((x==0)||(y==0)||(board.Width-x==1)||(board.Height-y==1)){
      spotpoints = 35
    }
  }
  if (status.StatusName==Snake) {
    if (i-status.SnakeOrder==-1){
      spotpoints-=20
    } else if (i-status.SnakeOrder< -1){
      spotpoints=0
    } else {
      spotpoints-=10
    }
  } else if (status.StatusName==Food) {
    spotpoints+=30
  } else if (status.StatusName==Hazard) {
    spotpoints-=20
  } else if (status.StatusName==Danger && i==0) {
    spotpoints=3
  } else if (status.StatusName==Outside) {
    spotpoints=0
  }
  return spotpoints
}

func ScoreLocation(x int, y int, boardrep BoardRep, points float64, i int, MaxDepth int)(float64, int){
  headscore := DetermineValue(x, y, boardrep, i) 
  if ((headscore>0.1)&&(i<MaxDepth)){
    points += (headscore * math.Exp(-0.25 * float64(i)) )
    i++
    boardrep.Board[x+boardrep.Offset][y+boardrep.Offset].StatusName=Snake
    boardrep.Board[x+boardrep.Offset][y+boardrep.Offset].SnakeOrder=int(boardrep.MyLength)+i
    points, i, _, _ = MoveSnake(x, y, boardrep, points, i, MaxDepth)
    }
  return points, i
}

func MoveSnake(x int, y int, boardrep BoardRep, points float64, i int, MaxDepth int) (float64, int, string, map[string][]float32) {
  dirmap := GetDirections()
  scorelist := make(map[string][]float32)
  imax := i
  pointsmax := points
  bestdir := "down"
  for name, directions := range dirmap{
    points_new, i_new := ScoreLocation(x+directions[0], y+directions[1], boardrep, points, i, MaxDepth)
    scorelist[name] = []float32{float32(points_new), float32(i_new)}
    if (i_new>imax)||((i_new==imax)&&(points_new>pointsmax)){
      imax = i_new
      pointsmax = points_new
      bestdir=name
      }
  }
  return pointsmax, imax, bestdir, scorelist
}

func Decide(r GameRequest) string {
  fmt.Printf("Current Location: %s\n",r.You.Head)
  representation := CreateRepresentation(r.Board, r.You)
  Depth := GetDepth(representation.OpenSlots)
  fmt.Printf("\n###################Depth: %s; Openslots: %s \n", Depth, representation.OpenSlots)
  points, i, move, scorelist := MoveSnake(r.You.Head.X, r.You.Head.Y, representation, 0, 0, Depth)
  
  fmt.Printf("Scorelist:\n %s \n", scorelist)
  fmt.Printf("Choosing '%s' with Score: %s and %s iterations\n", move, points, i)
  return move
}