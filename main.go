// Copyright 2012 Arne Roomann-Kurrik
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"github.com/kurrik/twodee"
	"image/color"
	"math"
	"os"
	"time"
)

const (
	HITLEFT   = 1 << iota
	HITRIGHT  = 1 << iota
	HITTOP    = 1 << iota
	HITBOTTOM = 1 << iota
	OK        = 1 << iota
)

func Check(err error) {
	if err != nil {
		fmt.Printf("[error]: %v\n", err)
		os.Exit(1)
	}
}

func Min(a float32, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func Max(a float32, b float32) float32 {
	if a < b {
		return b
	}
	return a
}

func Abs(a float32) float32 {
	if a < 0 {
		return -a
	}
	return a
}

func Round(a float32) float32 {
	if a > 0 {
		a += 0.5
	} else {
		a -= 0.5
	}
	return float32(math.Floor(float64(a)))
}

type LivesBar struct {
	twodee.Element
	used   int
	lives  int
	system *twodee.System
}

func NewLivesBar(system *twodee.System, lives int, used int) *LivesBar {
	bar := &LivesBar{
		used:   used,
		lives:  lives,
		system: system,
	}
	bar.Render()
	return bar
}

func (l *LivesBar) SetUsed(used int) {
	l.used = used
	l.Render()
}

func (l *LivesBar) Used() int {
	return l.used
}

func (l *LivesBar) SetLives(lives int) {
	l.lives = lives
	l.Render()
}

func (l *LivesBar) Lives() int {
	return l.lives
}

func (l *LivesBar) Render() {
	l.Clear()
	var (
		x  int             = 0
		t  *twodee.Texture = l.system.Textures["powerups-textures"]
		w0 int             = 2 * (t.Frames[0][1] - t.Frames[0][0])
		w1 int             = 2 * (t.Frames[1][1] - t.Frames[1][0])
		h  int             = 2 * t.Height
		y  float32         = -24
	)
	for i := 0; i < l.used; i++ {
		s := l.system.NewSprite("powerups-textures", float32(x), y, w1, h, 0)
		s.SetFrame(1)
		l.AddChild(s)
		x += w1
	}
	for i := 0; i < l.lives; i++ {
		s := l.system.NewSprite("powerups-textures", float32(x), y, w0, h, 0)
		s.SetFrame(0)
		l.AddChild(s)
		x += w0
	}
}

const (
	PLAYER_LEFT = 1 << iota
	PLAYER_RIGHT = 1 << iota
	PLAYER_STOPPED = 1 << iota
	PLAYER_WALKING = 1 << iota
	PLAYER_JUMPING = 1 << iota
)

type Animation struct {
	Frames   []int
	Duration time.Duration
}

func Anim(frames []int, ms int) *Animation {
	return &Animation{
		Frames:   frames,
		Duration: time.Duration(ms) * time.Millisecond,
	}
}

func (a *Animation) Len() int {
	return len(a.Frames)
}

type Player struct {
	Sprite       *twodee.Sprite
	State        int
	LastState    int
	JumpSpeed    float32
	WalkSpeed    float32
	RunSpeed     float32
	Acceleration float32
	Deceleration float32
	NextFrame    time.Time
	FrameCounter int
	Animations   map[int]*Animation
}

func (s *State) NewPlayer(x float32, y float32) (p *Player) {
	var (
		texture = s.system.Textures["darwin-textures"]
		width   = (texture.Frames[0][1] - texture.Frames[0][0]) * 2
		height  = texture.Height * 2
		starty  = y - float32(height)
	)
	a := map[int]*Animation{
		PLAYER_STOPPED | PLAYER_LEFT:  Anim([]int{4,5}, 400),
		PLAYER_STOPPED | PLAYER_RIGHT: Anim([]int{0,1}, 400),
		PLAYER_WALKING | PLAYER_LEFT:  Anim([]int{3,5}, 80),
		PLAYER_WALKING | PLAYER_RIGHT: Anim([]int{0,2}, 80),
		PLAYER_JUMPING | PLAYER_LEFT:  Anim([]int{5}, 80),
		PLAYER_JUMPING | PLAYER_RIGHT: Anim([]int{0}, 80),
	}
	p = &Player{
		Sprite:       s.system.NewSprite("darwin-textures", x, starty, width, height, PLAYER),
		State:        PLAYER_STOPPED | PLAYER_RIGHT,
		LastState:    PLAYER_STOPPED | PLAYER_RIGHT,
		NextFrame:    time.Now(),
		Animations:   a,
		FrameCounter: 0,
		JumpSpeed:    0.8,
		WalkSpeed:    0.05,
		RunSpeed:     0.8,
		Acceleration: 0.001,
		Deceleration: 0.001,
	}
	p.Sprite.SetFrame(0)
	return p
}

func (p *Player) Jump() {
	p.Sprite.VelocityY = -p.JumpSpeed
	p.State &= 511 ^ (PLAYER_STOPPED | PLAYER_JUMPING)
	p.State |= (PLAYER_JUMPING)
}

func (p *Player) Left(ms float32) {
	var v = p.Sprite.VelocityX - p.Acceleration*ms
	p.Sprite.VelocityX = Max(-p.RunSpeed, Min(-p.WalkSpeed, v))
	p.State &= 511 ^ (PLAYER_RIGHT | PLAYER_STOPPED | PLAYER_JUMPING)
	p.State |= (PLAYER_LEFT | PLAYER_WALKING)
}

func (p *Player) Right(ms float32) {
	var v = p.Sprite.VelocityX + p.Acceleration*ms
	p.Sprite.VelocityX = Min(p.RunSpeed, Max(p.WalkSpeed, v))
	p.State &= 511 ^ (PLAYER_LEFT | PLAYER_STOPPED | PLAYER_JUMPING)
	p.State |= (PLAYER_RIGHT | PLAYER_WALKING)
}

func (p *Player) Slow(ms float32) {
	if Abs(p.Sprite.VelocityX) <= p.Deceleration*ms {
		p.Sprite.VelocityX = 0
		p.State &= 511 ^ (PLAYER_WALKING | PLAYER_JUMPING)
		p.State |= PLAYER_STOPPED
	} else {
		if p.Sprite.VelocityX > 0 {
			p.Sprite.VelocityX -= p.Deceleration * ms
		} else {
			p.Sprite.VelocityX += p.Deceleration * ms
		}
	}
}

func (p *Player) Rebound(c *Creature) {
	if c.Sprite.X() > p.Sprite.X() {
		p.Sprite.VelocityX = -p.RunSpeed
	} else {
		p.Sprite.VelocityX = p.RunSpeed
	}
	if c.Sprite.Y() > p.Sprite.Y() {
		p.Sprite.VelocityY = -p.JumpSpeed
	} else {
		p.Sprite.VelocityY = p.JumpSpeed
	}
}

func (p *Player) Bounce(c *Creature) {
	if c.Sprite.Y() > p.Sprite.Y() {
		p.Sprite.VelocityY = -p.JumpSpeed
	} else {
		p.Sprite.VelocityY = p.JumpSpeed
	}
	p.Sprite.VelocityX = 0
}

func (p *Player) Update(ms float32) {
	if time.Now().After(p.NextFrame) || p.LastState != p.State {
		if anim, ok := p.Animations[p.State]; ok {
			i := p.FrameCounter % anim.Len()
			p.Sprite.SetFrame(anim.Frames[i])
			p.NextFrame = time.Now().Add(anim.Duration)
		}
		p.FrameCounter = (p.FrameCounter + 1) % 1000
		p.LastState = p.State
	}
}

type Creature struct {
	Sprite *twodee.Sprite
	Speed  float32
	Points int
}

type State struct {
	system     *twodee.System
	scene      *twodee.Scene
	hud        *twodee.Scene
	textscore  *twodee.Text
	textfps    *twodee.Text
	env        *twodee.Env
	window     *twodee.Window
	player     *Player
	livesbar   *LivesBar
	running    bool
	score      int
	nextlife   int
	boundaries []*twodee.Sprite
	creatures  []*Creature
	screenxmin float32
	screenxmax float32
	screenymin float32
	screenymax float32
}

func (s *State) SpawnCreature(t int, x float32, y float32) {
	var (
		c *Creature
	)
	switch t {
	case BADGUY:
		c = &Creature{
			Sprite: s.system.NewSprite("char-textures", x, y-64, 32, 64, t),
			Speed:  0.1,
			Points: 100,
		}
	}
	c.Sprite.SetFrame(1)
	c.Sprite.VelocityX = -c.Speed
	s.creatures = append(s.creatures, c)
	s.env.AddChild(c.Sprite)
}

func (s *State) KillCreature(c *Creature) {
	for i, d := range s.creatures {
		if d == c {
			s.creatures = append(s.creatures[:i], s.creatures[i+1:]...)
			break
		}
	}
	s.env.RemoveChild(c.Sprite)
	return

}

func (s *State) AddLife() {
	var (
		lives = s.livesbar.Lives()
	)
	s.livesbar.SetLives(lives + 1)
}

func (s *State) LoseLife() {
	var (
		lives = s.livesbar.Lives()
		used  = s.livesbar.Used()
	)
	s.livesbar.SetLives(lives - 1)
	s.livesbar.SetUsed(used + 1)
}

func (s *State) SetScore(score int) {
	s.score = score
	s.textscore.SetText(fmt.Sprintf("%v", s.score))
	s.textscore.MoveTo(twodee.Pt(s.window.View.Max.X-s.textscore.Width(), 0))
	if s.score >= s.nextlife {
		s.AddLife()
		s.nextlife *= 2
	}
}

func (s *State) Score() int {
	return s.score
}

func (s *State) HandleKeys(key, state int) {
	switch key {
	case twodee.KeyEsc:
		s.running = false
	}
}

func (s *State) CheckKeys(ms float32) {
	switch {
	case s.system.Key(twodee.KeyUp) == 1 && s.system.Key(twodee.KeyDown) == 0:
		s.player.Jump()
	case s.system.Key(twodee.KeyUp) == 0 && s.system.Key(twodee.KeyDown) == 1:
		//s.char.VelocityY = speed
	}
	switch {
	case s.system.Key(twodee.KeyLeft) == 1 && s.system.Key(twodee.KeyRight) == 0:
		s.player.Left(ms)
	case s.system.Key(twodee.KeyLeft) == 0 && s.system.Key(twodee.KeyRight) == 1:
		s.player.Right(ms)
	default:
		s.player.Slow(ms)
	}
}

func (s *State) Visible(sprite *twodee.Sprite) bool {
	var (
		wb = s.window.View.Sub(s.env.Bounds().Min)
		sb = sprite.RelativeBounds(s.env)
	)
	return sb.Overlaps(wb)
}

func (s *State) UpdateSprite(sprite *twodee.Sprite, ms float32) (result int) {
	sprite.VelocityY += 0.2 // Gravity
	var (
		dX    = sprite.VelocityX * ms
		dY    = sprite.VelocityY * ms
		b     = sprite.RelativeBounds(s.env)
		moved = false
	)
	if b.Min.X+dX < 0 {
		result |= HITLEFT
		sprite.VelocityX = 0
		dX = 0
	}
	if b.Max.X+dX > s.env.Width() {
		/*
			fmt.Printf("HITRIGHT\n")
			fmt.Printf("sprite.RelativeBounds(s.env) %v\n", sprite.RelativeBounds(s.env))
			fmt.Printf("sprite.LocalBounds() %v\n", sprite.Bounds())
		*/
		result |= HITRIGHT
		sprite.VelocityX = 0
		dX = 0
	}
	for _, block := range s.boundaries {
		if dX != 0 && !sprite.TestMove(dX, 0, block) {
			if sprite.TestMove(dX, -block.Height(), block) {
				// Allows running up small bumps
				sprite.Move(twodee.Pt(0, -block.Height()))
				moved = true
			} else {
				if dX < 0 {
					sprite.MoveTo(twodee.Pt(block.X()+block.Width(), sprite.Y()))
					result |= HITLEFT
				} else {
					sprite.MoveTo(twodee.Pt(block.X()-sprite.Width(), sprite.Y()))
					result |= HITRIGHT
				}
				moved = true
				sprite.VelocityX = 0
				dX = 0
			}
		}
		if dY != 0 && !sprite.TestMove(0, dY, block) {
			if dY < 0 {
				sprite.MoveTo(twodee.Pt(sprite.X(), block.Y()+block.Height()))
				result |= HITTOP
			} else {
				sprite.MoveTo(twodee.Pt(sprite.X(), block.Y()-sprite.Height()))
				result |= HITBOTTOM
			}
			moved = true
			sprite.VelocityY = 0
			dY = 0
		}
	}
	if dX != 0 || dY != 0 {
		sprite.Move(twodee.Pt(dX, dY))
		moved = true
	}
	sprite.MoveTo(twodee.Pt(Round(sprite.X()), Round(sprite.Y())))
	if moved {
	}
	return
}

func (s *State) IsKillShot(c *Creature) bool {
	var (
		buffery  = float32(c.Sprite.Height()) / 2
		downward = s.player.Sprite.VelocityY > 0
		hitshead = s.player.Sprite.Y()+float32(s.player.Sprite.Height())-c.Sprite.Y() < buffery
	)
	return downward && hitshead
}

func (s *State) Update(ms float32) {
	s.textfps.SetText(fmt.Sprintf("FPS %-5.1f", (1000.0 / ms)))
	for _, c := range s.creatures {
		if s.player.Sprite.CollidesWith(c.Sprite) {
			if s.IsKillShot(c) {
				s.SetScore(s.Score() + c.Points)
				fmt.Println("Kill")
				s.KillCreature(c)
				s.player.Bounce(c)
			} else {
				s.LoseLife()
				s.player.Rebound(c)
			}
		}
		if s.Visible(c.Sprite) {
			result := s.UpdateSprite(c.Sprite, ms)
			switch {
			case result&HITRIGHT == HITRIGHT:
				c.Sprite.VelocityX = -c.Speed
			case result&HITLEFT == HITLEFT:
				c.Sprite.VelocityX = c.Speed
			}
		}
	}
	s.player.Update(ms)
	s.UpdateSprite(s.player.Sprite, ms)
}

func (s *State) UpdateViewport(ms float32) {
	var (
		r  = 0.1 * ms
		b  = s.env.RelativeBounds(s.player.Sprite)
		v  = s.window.View
		x  = Min(Max(b.Min.X+v.Dx()/2, s.screenxmin), s.screenxmax)
		y  = Min(Max(b.Min.Y+v.Dy()/2, s.screenymin), s.screenymax)
		dy = y - s.env.Y()
		dx = x - s.env.X()
		d  = twodee.Pt(dx/r, dy/r)
	)
	/*
		fmt.Printf("Bounds %v\n", s.char.GlobalBounds())
		fmt.Printf("s.env.RelativeBounds(s.char) %v\n", s.env.RelativeBounds(s.char))
		fmt.Printf("s.char.RelativeBounds(s.env) %v\n", s.char.RelativeBounds(s.env))
		fmt.Printf("Moving viewport to %v, %v\n", x, y)
	*/
	if ms == 0 {
		s.env.MoveTo(twodee.Pt(x, y))
		return
	}
	if dy < 200 && dy > -200 {
		d.Y /= 10
	}
	s.env.Move(d)
}

func (s *State) HandleAddBlock(block *twodee.EnvBlock, sprite *twodee.Sprite, x float32, y float32) {
	switch block.Type {
	case START:
		s.player = s.NewPlayer(x, y)
		s.env.AddChild(s.player.Sprite)
		fallthrough
	case FLOOR:
		s.boundaries = append(s.boundaries, sprite)
	case BADGUY:
		s.SpawnCreature(block.Type, x, y)
	}
}

func (s *State) Running() bool {
	return s.running && s.window.Opened()
}

func (s *State) Paint(ms float32) {
	s.system.Paint(s.scene)
}

const (
	FLOOR = iota
	START
	PLAYER
	BADGUY
)

type TexInfo struct {
	Name  string
	Path  string
	Width int
}

func Init(system *twodee.System) (state *State, err error) {
	state = &State{}
	state.creatures = make([]*Creature, 0)
	state.boundaries = make([]*twodee.Sprite, 0)
	state.hud = &twodee.Scene{}
	state.scene = &twodee.Scene{}
	state.env = &twodee.Env{}
	state.window = &twodee.Window{
		Width:  640,
		Height: 480,
		Title:  "TDoS",
	}
	state.system = system
	state.system.Open(state.window)
	textures := []TexInfo{
		TexInfo{"level-textures", "assets/level-textures.png", 16},
		TexInfo{"char-textures", "assets/char-textures.png", 16},
		TexInfo{"font1-textures", "assets/font1-textures.png", 0},
		TexInfo{"darwin-textures", "assets/darwin-textures.png", 0},
		TexInfo{"powerups-textures", "assets/powerups-textures-fw.png", 0},
	}
	for _, t := range textures {
		if err = system.LoadTexture(t.Name, t.Path, twodee.IntNearest, t.Width); err != nil {
			return
		}
	}
	BlockHandler := func(block *twodee.EnvBlock, sprite *twodee.Sprite, x float32, y float32) {
		state.HandleAddBlock(block, sprite, x, y)
	}
	opts := twodee.EnvOpts{
		Blocks: []*twodee.EnvBlock{
			&twodee.EnvBlock{
				Color:      color.RGBA{153, 102, 0, 255}, // Dirt
				Type:       FLOOR,
				FrameIndex: 0,
				Handler:    BlockHandler,
			},
			&twodee.EnvBlock{
				Color:      color.RGBA{0, 204, 51, 255}, // Green top
				Type:       FLOOR,
				FrameIndex: 1,
				Handler:    BlockHandler,
			},
			&twodee.EnvBlock{
				Color:      color.RGBA{51, 102, 0, 255}, // Top left corner
				Type:       FLOOR,
				FrameIndex: 2,
				Handler:    BlockHandler,
			},
			&twodee.EnvBlock{
				Color:      color.RGBA{51, 153, 0, 255}, // Top right corner
				Type:       FLOOR,
				FrameIndex: 3,
				Handler:    BlockHandler,
			},
			&twodee.EnvBlock{
				Color:      color.RGBA{153, 153, 51, 255}, // Left dirt wall
				Type:       FLOOR,
				FrameIndex: 4,
				Handler:    BlockHandler,
			},
			&twodee.EnvBlock{
				Color:      color.RGBA{153, 153, 102, 255}, // Right dirt wall
				Type:       FLOOR,
				FrameIndex: 5,
				Handler:    BlockHandler,
			},
			&twodee.EnvBlock{
				Color:      color.RGBA{204, 204, 51, 255}, // Left grass cap
				Type:       FLOOR,
				FrameIndex: 6,
				Handler:    BlockHandler,
			},
			&twodee.EnvBlock{
				Color:      color.RGBA{204, 204, 102, 255}, // Right grass cap
				Type:       FLOOR,
				FrameIndex: 7,
				Handler:    BlockHandler,
			},
			&twodee.EnvBlock{
				Color:      color.RGBA{0, 0, 0, 255},
				Type:       START,
				FrameIndex: 1,
				Handler:    BlockHandler,
			},
			&twodee.EnvBlock{
				Color:      color.RGBA{51, 51, 51, 255},
				Type:       BADGUY,
				FrameIndex: -1,
				Handler:    BlockHandler,
			},
		},
		TextureName: "level-textures",
		MapPath:     "assets/level-fw.png",
		BlockWidth:  32,
		BlockHeight: 32,
	}
	if err = state.env.Load(system, opts); err != nil {
		return
	}
	state.system.SetClearColor(102, 204, 255, 255)
	state.scene.AddChild(state.env)
	state.system.SetKeyCallback(func(k, s int) { state.HandleKeys(k, s) })
	state.screenxmin = float32(-state.env.Width()) + state.window.View.Max.X
	state.screenxmax = 0
	state.screenymin = float32(-state.env.Height()) + state.window.View.Max.Y
	state.screenymax = 0

	// Do this later so that the hud renders last
	state.scene.AddChild(state.hud)
	state.livesbar = NewLivesBar(system, 0, 0)
	state.textscore = system.NewText("font1-textures", 0, 0, 2, "")
	state.textfps = system.NewText("font1-textures", 0, float32(state.window.View.Max.Y-32), 1, "")
	state.hud.AddChild(state.livesbar)
	state.hud.AddChild(state.textscore)
	state.hud.AddChild(state.textfps)
	state.hud.SetZ(0.5)
	state.nextlife = 100
	state.SetScore(0)
	state.AddLife()
	state.running = true
	return
}

func main() {
	system, err := twodee.Init()
	Check(err)
	defer system.Terminate()

	state, err := Init(system)
	Check(err)
	tick := time.Now()
	state.UpdateViewport(0)
	for state.Running() {
		elapsed := time.Since(tick)
		//fmt.Printf("Elapsed: %v\n", float32(elapsed) / float32(time.Millisecond))
		tick = time.Now()
		ms := Min(float32(elapsed)/float32(time.Millisecond), 50)
		state.CheckKeys(ms)
		state.Update(ms)
		state.UpdateViewport(ms)
		state.Paint(ms)
	}
}
