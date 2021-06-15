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
)

type status struct {
  StatusName  StatusType
  SnakeOrder  int
  HasHazard   bool
  IsDanger    bool
  LikelySnake bool
}

type snakenav struct {
  Head Coord
  Dir  Coord
  Size int
}

type BoardRep struct {
  Board     [MaxSize][MaxSize]status
	Height    int 
	Width     int
  Offset    int
  OpenSlots int
  MyLength  int32
  Snakes    []snakenav
}

func AddCoord(xy1 Coord, xy2 Coord) Coord {
  return Coord{xy1.X+xy2.X, xy1.Y+xy2.Y}
}

func SubtrCoord(xy1 Coord, xy2 Coord) Coord {
  return Coord{xy1.X-xy2.X, xy1.Y-xy2.Y}
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

func GetDirections() map[Coord]string {
  return map[Coord]string{
    {-1, 0}:"left",
    { 1, 0}:"right",
    { 0,-1}:"down",
    { 0, 1}:"up",
  }
}

func MarkBoard(board BoardRep, xy Coord, what StatusType)BoardRep {
  board.Board[xy.X+board.Offset][xy.Y+board.Offset].StatusName=what
  return board
}

func GetValue(board BoardRep, xy Coord) status {
  return board.Board[xy.X+board.Offset][xy.Y+board.Offset]
}

func CreateRepresentation(b Board, me Battlesnake) BoardRep {
  var BRep BoardRep
  Offset:=2
  BRep.Offset=Offset
  BRep.Height=b.Height
  BRep.Width=b.Width
  BRep.MyLength=me.Length
  for x := 0; x < b.Width; x++ {
    for y := 0; y < b.Height; y++ {
      BRep=MarkBoard(BRep, Coord{x,y}, Inside)
      BRep.OpenSlots+=1
    }
  }
  for _, xy := range b.Hazards {
    BRep.Board[xy.X+Offset][xy.Y+Offset].HasHazard=true
  }
  for _, xy := range b.Food {
    BRep=MarkBoard(BRep, xy, Food)
  }
  for _, snake := range b.Snakes {
    if (snake.ID != me.ID){
      snakedir := SubtrCoord(snake.Body[0], snake.Body[1])
      nav := snakenav{snake.Head, snakedir, int(snake.Length)}
      BRep.Snakes = append(BRep.Snakes, nav)
    }
    if ((snake.Length>=me.Length) && (snake.ID != me.ID)){
      dirmap := GetDirections()
      for directions, _ := range dirmap {
        dangerCoord := AddCoord(snake.Head, directions)
        BRep.Board[dangerCoord.X+Offset][dangerCoord.Y+Offset].IsDanger = true
      }
    }
    for snakei, xy := range snake.Body {
      BRep.OpenSlots-=1
      BRep=MarkBoard(BRep, xy, Snake)
      BRep.Board[xy.X+Offset][xy.Y+Offset].SnakeOrder=int(snake.Length)-snakei
    }

  }
  return BRep
}

func DetermineValue(xy Coord, board BoardRep, i int) float64 {
  var status = GetValue(board, xy)
  spotpoints := float64(50)
  if board.MyLength>9{
    spotpoints-= math.Abs(float64(xy.X-(board.Width /2)))/float64(board.Width ) *3 -
                math.Abs(float64(xy.Y-(board.Height/2)))/float64(board.Height) *3
    if ((xy.X==0)||(xy.Y==0)||(board.Width-xy.X==1)||(board.Height-xy.Y==1)){
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
  } else if (status.StatusName==Outside) {
    spotpoints=0
  }
  if ((status.IsDanger && i==0)||(status.LikelySnake)) {
    spotpoints=math.Max(spotpoints-50,3)
  } 
  if (status.HasHazard) {
    spotpoints-=20
  }
  return spotpoints
}

func ScoreLocation(xy Coord, boardrep BoardRep, points float64, i int, MaxDepth int)(float64, int){
  headscore := DetermineValue(xy, boardrep, i) 
  if ((headscore>0.1)&&(i<MaxDepth)){
    points += (headscore * math.Exp(-0.25 * float64(i)) )
    i++
    boardrep=MarkBoard(boardrep, xy, Snake)
    boardrep.Board[xy.X+boardrep.Offset][xy.Y+boardrep.Offset].SnakeOrder=int(boardrep.MyLength)+i
    for snakei, snake := range boardrep.Snakes{
      newhead:=AddCoord(snake.Head, snake.Dir)
      if GetValue(boardrep, newhead).StatusName!=Outside{
        boardrep.Board[newhead.X+boardrep.Offset][newhead.Y+boardrep.Offset].LikelySnake=true
        boardrep.Snakes[snakei].Head=newhead
      }
    }
    points, i, _, _ = MoveSnake(xy, boardrep, points, i, MaxDepth)
    }
  return points, i
}

func MoveSnake(xy Coord, boardrep BoardRep, points float64, i int, MaxDepth int) (float64, int, string, map[string][]float32) {
  dirmap := GetDirections()
  scorelist := make(map[string][]float32)
  imax := i
  pointsmax := points
  bestdir := "down"
  for directions, name := range dirmap{
    newloc := AddCoord(xy, directions)
    points_new, i_new := ScoreLocation( newloc, boardrep, points, i, MaxDepth)
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
  points, i, move, scorelist := MoveSnake(r.You.Head, representation, 0, 0, Depth)
  
  fmt.Printf("Scorelist:\n %s \n", scorelist)
  fmt.Printf("Choosing '%s' with Score: %s and %s iterations\n", move, points, i)
  return move
}