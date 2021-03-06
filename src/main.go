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
	"./twodee"
	"image/color"
	"math"
	"math/rand"
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

const (
	DEBUG = false
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
	avail      int
	max        int
	Availframe int
	Emptyframe int
	system     *twodee.System
}

func NewLivesBar(system *twodee.System, avail int, max int) *LivesBar {
	bar := &LivesBar{
		avail:      avail,
		max:        max,
		system:     system,
		Availframe: 0,
		Emptyframe: 1,
	}
	bar.Render()
	return bar
}

func (l *LivesBar) SetAvailable(avail int) int {
	if avail > l.max {
		avail = l.max
	}
	if avail < 0 {
		avail = 0
	}
	l.avail = avail
	l.Render()
	return avail
}

func (l *LivesBar) Available() int {
	return l.avail
}

func (l *LivesBar) SetMax(max int) int {
	if max < 0 {
		max = 0
	}
	if l.avail > max {
		l.avail = max
	}
	l.max = max
	l.Render()
	return max
}

func (l *LivesBar) Max() int {
	return l.max
}

func (l *LivesBar) Render() {
	l.Clear()
	var (
		x  int             = 0
		t  *twodee.Texture = l.system.Textures["powerups-textures"]
		we int             = 2 * (t.Frames[l.Emptyframe][1] - t.Frames[l.Emptyframe][0])
		wa int             = 2 * (t.Frames[l.Availframe][1] - t.Frames[l.Availframe][0])
		h  int             = 2 * t.Height
		y  float32         = -24
	)
	for i := 0; i < l.avail; i++ {
		s := l.system.NewSprite("powerups-textures", float32(x), y, wa, h, 0)
		s.SetFrame(l.Availframe)
		l.AddChild(s)
		x += wa + 2
	}
	for i := l.avail; i < l.max; i++ {
		s := l.system.NewSprite("powerups-textures", float32(x), y, we, h, 0)
		s.SetFrame(l.Emptyframe)
		l.AddChild(s)
		x += we + 2
	}
}

const (
	FACING_LEFT    = 1 << iota
	FACING_RIGHT   = 1 << iota
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
	StartX       float32
	StartY       float32
	invincible   bool
	vincibleat   time.Time
}

func (s *State) NewPlayer(x float32, y float32) (p *Player) {
	var (
		texture = s.system.Textures["darwin-textures"]
		width   = (texture.Frames[0][1] - texture.Frames[0][0]) * 2
		height  = texture.Height * 2
		starty  = y - float32(height)
	)
	a := map[int]*Animation{
		PLAYER_STOPPED | FACING_LEFT:  Anim([]int{4, 5}, 400),
		PLAYER_STOPPED | FACING_RIGHT: Anim([]int{0, 1}, 400),
		PLAYER_WALKING | FACING_LEFT:  Anim([]int{3, 5}, 80),
		PLAYER_WALKING | FACING_RIGHT: Anim([]int{0, 2}, 80),
		PLAYER_JUMPING | FACING_LEFT:  Anim([]int{5}, 80),
		PLAYER_JUMPING | FACING_RIGHT: Anim([]int{0}, 80),
	}
	p = &Player{
		Sprite:       s.system.NewSprite("darwin-textures", x, starty, width, height, PLAYER),
		State:        PLAYER_STOPPED | FACING_RIGHT,
		LastState:    PLAYER_STOPPED | FACING_RIGHT,
		NextFrame:    time.Now(),
		Animations:   a,
		StartX:       x,
		StartY:       y,
		FrameCounter: 0,
		JumpSpeed:    1.2,
		WalkSpeed:    0.03,
		RunSpeed:     0.6,
		Acceleration: 0.001,
		Deceleration: 0.001,
		invincible:   false,
	}
	p.Sprite.SetZ(1)
	p.Sprite.SetFrame(0)
	return p
}

func (p *Player) Invincible() bool {
	return p.invincible
}

func (p *Player) SetInvincible() {
	p.invincible = true
	p.vincibleat = time.Now().Add(time.Duration(200) * time.Millisecond)
}

func (p *Player) Respawn() {
	p.Sprite.Collide = true
	p.Sprite.VelocityY = 0
	p.Sprite.VelocityX = 0
	p.Sprite.MoveTo(twodee.Pt(p.StartX, p.StartY))
}

func (p *Player) Die() {
	p.Sprite.Collide = false
	p.Sprite.VelocityY = -p.JumpSpeed
	p.Sprite.VelocityX = 0
}

func (p *Player) Jump() {
	if p.State&PLAYER_JUMPING != PLAYER_JUMPING {
		p.Sprite.VelocityY = -p.JumpSpeed
		p.State &= 511 ^ (PLAYER_STOPPED | PLAYER_WALKING)
		p.State |= (PLAYER_JUMPING)
	}
}

func (p *Player) Left(ms float32) {
	var v = p.Sprite.VelocityX - p.Acceleration*ms
	p.Sprite.VelocityX = Max(-p.RunSpeed, Min(-p.WalkSpeed, v))
	p.State &= 511 ^ (FACING_RIGHT | PLAYER_STOPPED)
	p.State |= (FACING_LEFT | PLAYER_WALKING)
}

func (p *Player) Right(ms float32) {
	var v = p.Sprite.VelocityX + p.Acceleration*ms
	p.Sprite.VelocityX = Min(p.RunSpeed, Max(p.WalkSpeed, v))
	p.State &= 511 ^ (FACING_LEFT | PLAYER_STOPPED)
	p.State |= (FACING_RIGHT | PLAYER_WALKING)
}

func (p *Player) Slow(ms float32) {
	if Abs(p.Sprite.VelocityX) <= p.Deceleration*ms {
		p.Sprite.VelocityX = 0
		p.State &= 511 ^ (PLAYER_WALKING)
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
	if c.Sprite.X() >= p.Sprite.X() {
		p.Sprite.VelocityX = -p.RunSpeed
	} else {
		p.Sprite.VelocityX = p.RunSpeed
	}
	if c.Sprite.Y() >= p.Sprite.Y() {
		p.Sprite.VelocityY = -p.JumpSpeed
	} else {
		p.Sprite.VelocityY = p.JumpSpeed
	}
	p.State &= 511 ^ (PLAYER_STOPPED | PLAYER_WALKING)
	p.State |= (PLAYER_JUMPING)
}

func (p *Player) Bounce(c *Creature) {
	if c.Sprite.Y() >= p.Sprite.Y() {
		p.Sprite.VelocityY = -p.JumpSpeed
		p.Sprite.Move(twodee.Pt(0,-2)) // Clear collision zone
	} else {
		p.Sprite.VelocityY = p.JumpSpeed
		p.Sprite.Move(twodee.Pt(0,2)) // Clear collision zone
	}
	p.Sprite.VelocityX = 0
	p.State &= 511 ^ (PLAYER_STOPPED | PLAYER_WALKING)
	p.State |= (PLAYER_JUMPING)

}

func (p *Player) Update(result int, ms float32) {
	if result&HITBOTTOM == HITBOTTOM {
		p.State &= 511 ^ (PLAYER_JUMPING)
	}
	if time.Now().After(p.NextFrame) || p.LastState != p.State {
		if anim, ok := p.Animations[p.State]; ok {
			i := p.FrameCounter % anim.Len()
			p.Sprite.SetFrame(anim.Frames[i])
			p.NextFrame = time.Now().Add(anim.Duration)
		}
		p.FrameCounter = (p.FrameCounter + 1) % 1000
		p.LastState = p.State
	}
	if p.invincible && time.Now().After(p.vincibleat) {
		p.invincible = false
	}
}

const (
	MUSHROOM = iota
	SMALL_MUSHROOM
)

type Creature struct {
	Sprite       *twodee.Sprite
	Type         int
	Points       int
	State        int
	LastState    int
	Speed        float32
	JumpSpeed    float32
	NextFrame    time.Time
	FrameCounter int
	Animations   map[int]*Animation
	LastSpawn    time.Time
}

func (s *State) NewCreature(t string, x float32, y float32, z int) (c *Creature) {
	var (
		texture = s.system.Textures[t]
		width   = (texture.Frames[0][1] - texture.Frames[0][0]) * 2
		height  = texture.Height * 2
		starty  = y - float32(height)
	)
	a := map[int]*Animation{}
	c = &Creature{
		Sprite:       s.system.NewSprite(t, x, starty, width, height, BADGUY),
		Type:         z,
		State:        FACING_LEFT,
		LastState:    FACING_RIGHT,
		NextFrame:    time.Now(),
		LastSpawn:    time.Now(),
		Animations:   a,
		FrameCounter: 0,
		JumpSpeed:    0.8,
		Speed:        0.05,
		Points:       5,
	}
	c.Sprite.SetFrame(0)
	c.Sprite.VelocityX = -c.Speed
	return
}

func (c *Creature) Update(result int, ms float32) {
	if time.Now().After(c.NextFrame) || c.LastState != c.State {
		if anim, ok := c.Animations[c.State]; ok {
			i := c.FrameCounter % anim.Len()
			c.Sprite.SetFrame(anim.Frames[i])
			c.NextFrame = time.Now().Add(anim.Duration)
		}
		c.FrameCounter = (c.FrameCounter + 1) % 1000
		c.LastState = c.State
	}
	switch {
	case result&HITRIGHT == HITRIGHT:
		c.State &= 511 ^ (FACING_RIGHT)
		c.State |= (FACING_LEFT)
		c.Sprite.VelocityX = -c.Speed
	case result&HITLEFT == HITLEFT:
		c.State &= 511 ^ (FACING_LEFT)
		c.State |= (FACING_RIGHT)
		c.Sprite.VelocityX = c.Speed
	}
	if diff := Abs(c.Sprite.VelocityX) - c.Speed; diff != 0 {
		damp := diff / 10
		if c.Sprite.VelocityX > 0 {
			c.Sprite.VelocityX -= damp
		} else {
			c.Sprite.VelocityX += damp
		}
	}
}

func (s *State) NewMushroom(x float32, y float32) *Creature {
	c := s.NewCreature("enemy-textures", x, y, MUSHROOM)
	c.Speed = 0.05
	c.JumpSpeed = 0.1
	c.Points = 100
	c.Animations = map[int]*Animation{
		FACING_LEFT:  Anim([]int{0, 1}, 120),
		FACING_RIGHT: Anim([]int{2, 3}, 120),
	}
	return c
}

func (s *State) NewSmallMushroom(x float32, y float32) *Creature {
	c := s.NewCreature("enemy-sm-textures", x, y, SMALL_MUSHROOM)
	c.Speed = 0.08
	c.JumpSpeed = 0.3
	c.Points = 250
	c.Animations = map[int]*Animation{
		FACING_LEFT:  Anim([]int{0, 1}, 120),
		FACING_RIGHT: Anim([]int{2, 3}, 120),
	}
	return c
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
	healthbar  *LivesBar
	running    bool
	Victory    bool
	score      int
	nextlife   int
	boundaries []*twodee.Sprite
	creatures  []*Creature
	screenxmin float32
	screenxmax float32
	screenymin float32
	screenymax float32
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

func (s *State) SetMaxHealth(health int) {
	s.healthbar.SetMax(health)
}

func (s *State) ChangeHealth(change int) int {
	if change < 0 && s.player.Invincible() {
		return s.healthbar.Available()
	}
	s.player.SetInvincible()
	var health = s.healthbar.SetAvailable(s.healthbar.Available() + change)
	return health
}

func (s *State) ChangeMaxLives(i int) {
	s.livesbar.SetMax(s.livesbar.Max() + i)
}

func (s *State) ChangeLives(i int) int {
	var lives = s.livesbar.SetAvailable(s.livesbar.Available() + i)
	if lives == 0 {
		s.running = false
	}
	return lives
}

func (s *State) SetScore(score int) {
	s.score = score
	s.textscore.SetText(fmt.Sprintf("%v", s.score))
	s.textscore.MoveTo(twodee.Pt(s.window.View.Max.X-s.textscore.Width(), 0))
	if s.score >= s.nextlife {
		s.ChangeMaxLives(1)
		s.ChangeLives(1)
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
		wb     = s.window.View.Sub(s.env.Bounds().Min)
		sb     = sprite.RelativeBounds(s.env)
		buffer = float32(256)
	)
	wb.Min.X -= buffer
	wb.Min.Y -= buffer
	wb.Max.X += buffer
	wb.Max.X += buffer
	return sb.Overlaps(wb)
}

func (s *State) UpdateSprite(sprite *twodee.Sprite, ms float32) (result int) {
	sprite.VelocityY += 0.005 * ms // Gravity
	var (
		dX = sprite.VelocityX * ms
		dY = sprite.VelocityY * ms
		b  = sprite.RelativeBounds(s.env)
	)
	if b.Min.X+dX < 0 {
		result |= HITLEFT
		sprite.VelocityX = 0
		sprite.Move(twodee.Pt(1, 0))
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
		sprite.Move(twodee.Pt(-1, 0))
		dX = 0
	}
	if sprite.Collide {
		for _, block := range s.boundaries {
			if dX != 0 && !sprite.TestMove(dX, 0, block) {
				if sprite.TestMove(dX, -block.Height(), block) {
					// Allows running up small bumps
					sprite.Move(twodee.Pt(0, -block.Height()))
				} else {
					if dX < 0 {
						sprite.MoveTo(twodee.Pt(block.X()+block.Width(), sprite.Y()))
						result |= HITLEFT
					} else {
						sprite.MoveTo(twodee.Pt(block.X()-sprite.Width(), sprite.Y()))
						result |= HITRIGHT
					}
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
				sprite.VelocityY = 0
				dY = 0
			}
		}
	}
	if dX != 0 || dY != 0 {
		sprite.Move(twodee.Pt(dX, dY))
	}
	//sprite.MoveTo(twodee.Pt(Round(sprite.X()), Round(sprite.Y())))
	return
}

func (s *State) IsKillShot(c *Creature) bool {
	var (
		downward = s.player.Sprite.VelocityY > 0.1
		//jumping = s.player.State&PLAYER_JUMPING == PLAYER_JUMPING
	)
	//return downward && jumping
	return downward //jumping wouldn't let you walk off cliffs
}

func (s *State) Update(ms float32) {
	if DEBUG {
		s.textfps.SetText(fmt.Sprintf("FPS %-5.1f", (1000.0 / ms)))
	}
	for _, c := range s.creatures {
		if s.player.Sprite.Collide {
			if s.player.Sprite.CollidesWith(c.Sprite) {
				if s.IsKillShot(c) {
					s.SetScore(s.Score() + c.Points)
					s.KillCreature(c)
					s.player.Bounce(c)
				} else {
					health := s.healthbar.Available()
					if !s.player.Invincible() {
						s.player.Rebound(c)
						health = s.ChangeHealth(-1)
					}
					if health == 0 {
						s.player.Die()
					}
				}
			}
		}
		if s.Visible(c.Sprite) {
			result := s.UpdateSprite(c.Sprite, ms)
			c.Update(result, ms)
			switch c.Type {
			case MUSHROOM:
				thresh := time.Duration(5) * time.Second
				if time.Now().After(c.LastSpawn.Add(thresh)) {
					if rand.Float32() > 0.95 {
						c2 := s.NewSmallMushroom(c.Sprite.X(), c.Sprite.Y())
						c2.Sprite.VelocityX *= rand.Float32() * 2.0
						if rand.Float32() > 0.5 {
							c2.Sprite.VelocityX *= -1
						}
						c2.Sprite.VelocityY = -c2.JumpSpeed
						s.creatures = append(s.creatures, c2)
						s.env.AddChild(c2.Sprite)
						c.LastSpawn = time.Now()
					}
				}
			}
		}
	}

	result := s.UpdateSprite(s.player.Sprite, ms)
	s.player.Update(result, ms)

	var b = s.player.Sprite.RelativeBounds(s.env)
	if b.Max.Y > s.env.Height()+1000 {
		//Player has fallen off the map
		lives := s.ChangeLives(-1)
		if lives > 0 {
			s.ChangeHealth(s.healthbar.Max())
			s.player.Respawn()
			s.UpdateViewport(0)
		}
	}
	if b.Max.X >= s.env.Width() - 100 {
		// Poor man's victory
		s.running = false
		s.Victory = true
	}

}

func (s *State) UpdateViewport(ms float32) {
	var (
		r  = 0.1 * ms
		b  = s.env.RelativeBounds(s.player.Sprite)
		v  = s.window.View
		x  = Min(Max(b.Min.X+v.Dx()/2, s.screenxmin), s.screenxmax)
		y  = Min(Max(b.Min.Y+v.Dy()/2-s.player.Sprite.Height()/2, s.screenymin), s.screenymax)
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
	if s.player.Sprite.Collide {
		// Only smooth motion if the player isn't dying
		if ms == 0 || (dy < 1 && dy > -1) {
			s.env.MoveTo(twodee.Pt(x, y))
			return
		}
		if dy > 0 {
			d.Y = Max(1, dy/30)
		} else {
			d.Y = Min(-1, dy/30)
		}
	}
	s.env.Move(d)
	s.env.MoveTo(twodee.Pt(Round(s.env.X()), Round(s.env.Y())))
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
		c := s.NewMushroom(x, y)
		s.creatures = append(s.creatures, c)
		s.env.AddChild(c.Sprite)
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

func Init(system *twodee.System, window *twodee.Window) (state *State, err error) {
	state = &State{}
	state.creatures = make([]*Creature, 0)
	state.boundaries = make([]*twodee.Sprite, 0)
	state.hud = &twodee.Scene{}
	state.scene = &twodee.Scene{}
	state.env = &twodee.Env{}
	state.window = window
	state.system = system
	textures := []TexInfo{
		TexInfo{"level-textures", "assets/level-textures.png", 16},
		TexInfo{"enemy-sm-textures", "assets/enemy-sm-textures-fw.png", 0},
		TexInfo{"enemy-textures", "assets/enemy-textures-fw.png", 0},
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
				Color:      color.RGBA{153, 153, 153, 255}, // Rock
				Type:       FLOOR,
				FrameIndex: 8,
				Handler:    BlockHandler,
			},
			&twodee.EnvBlock{
				Color:      color.RGBA{118, 118, 118, 255}, // Rock left
				Type:       FLOOR,
				FrameIndex: 9,
				Handler:    BlockHandler,
			},
			&twodee.EnvBlock{
				Color:      color.RGBA{84, 84, 84, 255}, // Rock right
				Type:       FLOOR,
				FrameIndex: 10,
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

	// Do this later so that the hud renders on top of things
	state.scene.AddChild(state.hud)
	state.livesbar = NewLivesBar(system, 0, 0)
	state.hud.AddChild(state.livesbar)

	state.healthbar = NewLivesBar(system, 0, 0)
	state.healthbar.Availframe = 3
	state.healthbar.Emptyframe = 2
	state.healthbar.MoveTo(twodee.Pt(0, 24))
	state.hud.AddChild(state.healthbar)

	state.textscore = system.NewText("font1-textures", 0, 0, 2, "")
	state.hud.AddChild(state.textscore)

	state.textfps = system.NewText("font1-textures", 0, float32(state.window.View.Max.Y-32), 1, "")
	state.hud.AddChild(state.textfps)
	state.hud.SetZ(0.5)
	state.nextlife = 400
	state.SetScore(0)
	state.ChangeMaxLives(1)
	state.ChangeLives(1)
	state.SetMaxHealth(3)
	state.ChangeHealth(3)
	state.running = true
	state.Victory = false
	return
}

type Splash struct {
	running bool
	window  *twodee.Window
	system  *twodee.System
	scene   *twodee.Scene
	sprite  *twodee.Sprite
	started time.Time
}

func InitSplash(system *twodee.System, window *twodee.Window, frame int) (splash *Splash, err error) {
	if system.LoadTexture("splash", "assets/splash-fw.png", twodee.IntNearest, 0); err != nil {
		return
	}
	splash = &Splash{
		running: true,
		window:  window,
		system:  system,
		scene:   &twodee.Scene{},
	}
	system.SetKeyCallback(func(k, s int) {
		threshold := time.Duration(500) * time.Millisecond
		if time.Now().After(splash.started.Add(threshold)) {
			splash.running = false
		}
	})
	fmt.Println(system.Textures)
	splash.sprite = system.NewSprite("splash", 0, 0, int(window.View.Dx()), int(window.View.Dy()), 0)
	splash.sprite.SetFrame(frame)
	splash.scene.AddChild(splash.sprite)
	splash.started = time.Now()
	return
}

func (s *Splash) Running() bool {
	return s.running && s.window.Opened()
}

func (s *Splash) Paint() {
	threshold := time.Duration(5) * time.Second
	if time.Now().After(s.started.Add(threshold)) {
		s.running = false
	}
	s.system.Paint(s.scene)
}

func main() {
	var (
		splash *Splash
		system *twodee.System
		window *twodee.Window
		err    error
	)
	system, err = twodee.Init()
	Check(err)
	defer system.Terminate()

	window = &twodee.Window{
		Width:  800,
		Height: 600,
		Title:  "TDoS",
		Fullscreen: false,
	}
	system.Open(window)

	if !DEBUG {
		splash, err = InitSplash(system, window, 0)
		Check(err)
		for splash.Running() {
			splash.Paint()
		}
		splash = nil
	}

	state, err := Init(system, window)
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

	if !DEBUG {
		frame := 1
		if state.Victory {
			frame = 2
		}
		splash, err = InitSplash(system, window, frame)
		Check(err)
		for splash.Running() {
			splash.Paint()
		}
	}
}
