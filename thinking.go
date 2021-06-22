package main

import (
	"fmt"
  "math"
  "time"
)

const cutoff = time.Second/20

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
  DangerLvl   float64 // Tracks radius around dangerous snakes
  LikelySnake bool // Tracks progress of dangerous snakes assuming they move in the same direction
}

type snakenav struct {
  Head Coord
  Dir  Coord
  Size int
}

type BoardRep struct {
  Board     [][]status
	Height    int 
	Width     int
  Offset    int
  OpenSlots int
  MyLength  int32
  Snakes    []snakenav
}

func copyBoard(arr BoardRep) BoardRep {
  tmp := arr
	tmp_slice := make([][]status, len(arr.Board))
	for i:=0;i<len(arr.Board);i++ {
	    tmp_slice[i] = make([]status, len(arr.Board[i]))
      copy(tmp_slice[i], arr.Board[i])
	}
  tmp.Board=tmp_slice
	return tmp
}

func MakeEmptyBoardRep(brd Board, offset int) BoardRep {
	var BRep BoardRep
	BRep.Board = make([][]status, brd.Width+offset*2)	
	for i:=0;i<brd.Width+offset*2;i++ {
	    BRep.Board[i] = make([]status, brd.Height+offset*2)
	}
  BRep.Offset=offset
  BRep.Height=brd.Height
  BRep.Width=brd.Width
	return BRep
}

func AddCoord(xy1 Coord, xy2 Coord) Coord {
  return Coord{xy1.X+xy2.X, xy1.Y+xy2.Y}
}

func SubtrCoord(xy1 Coord, xy2 Coord) Coord {
  return Coord{xy1.X-xy2.X, xy1.Y-xy2.Y}
}

func IsCoordOnBoard(xy Coord, board BoardRep) bool {
  if (xy.X<0 || xy.Y<0) {
    return false
  } else if (xy.X>=board.Width || xy.Y>=board.Height){
    return false
  } else {
    return true
  }
}

func GetDepth(slots int) int {
  if (slots>105){
    return 5
  } else if (slots>100){
    return 6
  } else if (slots>80){
    return 7
  } else if (slots>60){
    return 8
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
  Offset:=1
  BRep := MakeEmptyBoardRep(b, Offset)
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
    snakedir := SubtrCoord(snake.Body[0], snake.Body[1])
    if (snake.ID != me.ID){
      nav := snakenav{snake.Head, snakedir, int(snake.Length)}
      BRep.Snakes = append(BRep.Snakes, nav)
    }
    if ((snake.Length>=me.Length) && (snake.ID != me.ID)){
      dirmap := GetDirections()
      toAdd:= map[bool]float64{true:1.34, false:.67}
      for directions, _ := range dirmap {
        dangerCoord := AddCoord(snake.Head, directions)
        for directions2, _ := range dirmap {
          dangerCoord2 := AddCoord(dangerCoord, directions2)
          if IsCoordOnBoard(dangerCoord2, BRep) {
            BRep.Board[dangerCoord2.X+Offset][dangerCoord2.Y+Offset].DangerLvl += 2*toAdd[directions==snakedir||directions2==snakedir]
          }
          for directions3, _ := range dirmap {
            dangerCoord3 := AddCoord(dangerCoord2, directions3)
            if IsCoordOnBoard(dangerCoord3, BRep) {
              BRep.Board[dangerCoord3.X+Offset][dangerCoord3.Y+Offset].DangerLvl  += toAdd[directions3==snakedir]
            }
          }
        }
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
  spotpoints := float64(100)
  if board.MyLength>9{
    spotpoints-= math.Abs(float64(xy.X-(board.Width /2)))/float64(board.Width ) *3 -
                math.Abs(float64(xy.Y-(board.Height/2)))/float64(board.Height) *3
    if ((xy.X==0)||(xy.Y==0)||(board.Width-xy.X==1)||(board.Height-xy.Y==1)){
      spotpoints = 75
    }
  }
  contested := status.DangerLvl>=10 && i==0
  if (contested||status.LikelySnake) {
    spotpoints-=50
  } else {
    spotpoints-=status.DangerLvl
  }
  if (status.HasHazard) {
    spotpoints-=20
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
    if (contested){
      spotpoints-=10
    } else {
      spotpoints+=30
    }    
  } else if (status.StatusName==Outside) {
    spotpoints=0
  }
  return spotpoints
}

func ScoreLocation(xy Coord, boardrep BoardRep, points float64, i int, MaxDepth int, t0 time.Time)(float64, int){
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
    points, i, _, _ = MoveSnake(xy, boardrep, points, i, MaxDepth, t0)
    }
  return points, i
}

func MoveSnake(xy Coord, boardrep BoardRep, points float64, i int, MaxDepth int, t0 time.Time) (float64, int, string, map[string][]float32) {
  dirmap := GetDirections()
  scorelist := make(map[string][]float32)
  imax := i
  pointsmax := points
  bestdir := "down"
  for directions, name := range dirmap{
    if i==0 {
      t0 = time.Now()
    }
    if (time.Since(t0)<cutoff ){
      newloc := AddCoord(xy, directions)
      boardrepCopy := copyBoard(boardrep)
      points_new, i_new := ScoreLocation( newloc, boardrepCopy, points, i, MaxDepth, t0)
      scorelist[name] = []float32{float32(points_new), float32(i_new)}
      if (i_new>imax)||((i_new==imax)&&(points_new>pointsmax)){
        imax = i_new
        pointsmax = points_new
        bestdir=name
        }
    }
  }
  return pointsmax, imax, bestdir, scorelist
}

func Decide(r GameRequest) string {
  fmt.Printf("Current Location: %s\n",r.You.Head)
  representation := CreateRepresentation(r.Board, r.You)
  Depth := GetDepth(representation.OpenSlots)
  fmt.Printf("\n###################Depth: %s; Openslots: %s \n", Depth, representation.OpenSlots)
  for rowi, row := range representation.Board{
    fmt.Printf("\n Row number: %s\n %s \n", rowi, row)
  }
  points, i, move, scorelist := MoveSnake(r.You.Head, representation, 0, 0, Depth, time.Now())
  fmt.Printf("Scorelist:\n %s \n", scorelist)
  fmt.Printf("Choosing '%s' with Score: %s and %s iterations\n", move, points, i)
  return move
}