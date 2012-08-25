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
	"os"
	"time"
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
	return float32(int32(a + 0.5))
}

type State struct {
	system     *twodee.System
	scene      *twodee.Scene
	hud        *twodee.Scene
	textscore  *twodee.Text
	textfps    *twodee.Text
	env        *twodee.Env
	window     *twodee.Window
	char       *twodee.Sprite
	running    bool
	score      int
	boundaries []*twodee.Sprite
	creatures  []*twodee.Sprite
	screenxmin float32
	screenxmax float32
	screenymin float32
	screenymax float32
}

func (s *State) SetScore(score int) {
	s.score = score
	s.textscore.SetText(fmt.Sprintf("%v", s.score))
	s.textscore.X = float32(s.window.Width - s.textscore.Width)
}

func (s *State) HandleKeys(key, state int) {
	switch key {
	case twodee.KeyEsc:
		s.running = false
	}
}

func (s *State) CheckKeys(ms float32) {
	var speed float32 = 1.3
	var minspeed float32 = 0.05
	var accel float32 = 0.001 * ms
	var decel float32 = 0.005 * ms
	switch {
	case s.system.Key(twodee.KeyUp) == 1 && s.system.Key(twodee.KeyDown) == 0:
		s.char.VelocityY = -speed
	case s.system.Key(twodee.KeyUp) == 0 && s.system.Key(twodee.KeyDown) == 1:
		//s.char.VelocityY = speed
	}
	switch {
	case s.system.Key(twodee.KeyLeft) == 1 && s.system.Key(twodee.KeyRight) == 0:
		s.char.VelocityX = Max(-speed, Min(-minspeed, s.char.VelocityX-accel))
	case s.system.Key(twodee.KeyLeft) == 0 && s.system.Key(twodee.KeyRight) == 1:
		s.char.VelocityX = Min(speed, Max(minspeed, s.char.VelocityX+accel))
	default:
		if Abs(s.char.VelocityX) <= decel {
			s.char.VelocityX = 0
		} else {
			if s.char.VelocityX > 0 {
				s.char.VelocityX -= decel
			} else {
				s.char.VelocityX += decel
			}
		}
	}
}

func (s *State) Visible(sprite *twodee.Sprite) bool {
	var (
		left   = 0 - s.env.X
		right  = left + float32(s.window.Width)
		top    = 0 - s.env.Y
		bottom = top + float32(s.window.Height)
	)
	var (
		inX = left <= sprite.X-float32(sprite.Width) && sprite.X <= right
		inY = top <= sprite.Y-float32(sprite.Height) && sprite.Y <= bottom
	)
	return inX && inY
}

func (s *State) UpdateSprite(sprite *twodee.Sprite, ms float32) {
	sprite.VelocityY += 0.3 // Gravity
	dX := Round(sprite.VelocityX * ms)
	dY := Round(sprite.VelocityY * ms)
	for _, b := range s.boundaries {
		if dX != 0 && !sprite.TestMove(dX, 0, b) {
			if sprite.TestMove(dX, float32(-b.Height), b) {
				sprite.Y -= float32(b.Height)
			} else {
				if dX < 0 {
					sprite.X = b.X + float32(b.Width)
				} else {
					sprite.X = b.X - float32(sprite.Width)
				}
				sprite.VelocityX = 0
				dX = 0
			}
		}
		if dY != 0 && !sprite.TestMove(0, dY, b) {
			if dY < 0 {
				sprite.Y = b.Y + float32(b.Height)
			} else {
				sprite.Y = b.Y - float32(sprite.Height)
			}
			sprite.VelocityY = 0
			dY = 0
		}
	}
	sprite.X = Round(sprite.X + dX)
	sprite.Y = Round(sprite.Y + dY)
}

func (s *State) Update(ms float32) {
	s.textfps.SetText(fmt.Sprintf("FPS %-5.1f", (1000.0 / ms)))
	s.UpdateSprite(s.char, ms)
	for _, c := range s.creatures {
		if s.Visible(c) {
			s.UpdateSprite(c, ms)
		}
	}
}

func (s *State) UpdateViewport() {
	s.env.X = 0 - s.char.X + (float32(s.window.Width) / 2)
	s.env.Y = 0 - s.char.Y + (float32(s.window.Height) / 2)
	if s.env.X < s.screenxmin {
		s.env.X = s.screenxmin
	}
	if s.env.X > s.screenxmax {
		s.env.X = s.screenxmax
	}
	if s.env.Y < s.screenymin {
		s.env.Y = s.screenymin
	}
	if s.env.Y > s.screenymax {
		s.env.Y = s.screenymax
	}
}

func (s *State) HandleAddBlock(env *twodee.Env, block *twodee.EnvBlock, sprite *twodee.Sprite, x float32, y float32) {
	switch block.Type {
	case START:
		s.char = s.system.NewSprite("char-textures", 0, 0, 32, 64, PLAYER)
		s.char.SetFrame(2)
		s.char.X = sprite.X
		s.char.Y = sprite.Y - 100
		env.AddChild(s.char)
		fallthrough
	case FLOOR:
		s.boundaries = append(s.boundaries, sprite)
	case BADGUY:
		badguy := s.system.NewSprite("char-textures", x, y - 64, 32, 64, BADGUY)
		badguy.SetFrame(1)
		s.creatures = append(s.creatures, badguy)
		env.AddChild(badguy)
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
	var (
		env  *twodee.Env
		opts twodee.EnvOpts
	)
	state = &State{}
	state.creatures = make([]*twodee.Sprite, 0)
	state.boundaries = make([]*twodee.Sprite, 0)
	state.hud = &twodee.Scene{}
	state.scene = &twodee.Scene{}
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
	}
	for _, t := range textures {
		if err = system.LoadTexture(t.Name, t.Path, twodee.IntNearest, t.Width); err != nil {
			return
		}
	}
	BlockHandler := func(env *twodee.Env, block *twodee.EnvBlock, sprite *twodee.Sprite, x float32, y float32) {
		state.HandleAddBlock(env, block, sprite, x, y)
	}
	opts = twodee.EnvOpts{
		Blocks: []*twodee.EnvBlock{
			&twodee.EnvBlock{
				Color:      color.RGBA{153, 102, 0, 255},
				Type:       FLOOR,
				FrameIndex: 0,
				Handler:    BlockHandler,
			},
			&twodee.EnvBlock{
				Color:      color.RGBA{0, 204, 51, 255},
				Type:       FLOOR,
				FrameIndex: 1,
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
		BlockWidth:  16,
		BlockHeight: 16,
	}
	if env, err = system.LoadEnv(opts); err != nil {
		return
	}
	state.system.SetClearColor(102, 204, 255, 255)
	state.env = env
	state.scene.AddChild(env)
	state.system.SetKeyCallback(func(k, s int) { state.HandleKeys(k, s) })
	state.screenxmin = float32(-state.env.Width + state.window.Width)
	state.screenxmax = 0
	state.screenymin = float32(-state.env.Height + state.window.Height)
	state.screenymax = 0

	// Do this later so that the hud renders last
	state.scene.AddChild(state.hud)
	state.textscore = system.NewText("font1-textures", 0, 0, 4, "")
	state.textfps = system.NewText("font1-textures", 0, float32(state.window.Height-32), 2, "")
	state.hud.AddChild(state.textscore)
	state.hud.AddChild(state.textfps)
	state.hud.Z = 0.5
	state.SetScore(0)
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
	for state.Running() {
		elapsed := time.Since(tick)
		//fmt.Printf("Elapsed: %v\n", float32(elapsed) / float32(time.Millisecond))
		tick = time.Now()
		ms := Min(float32(elapsed)/float32(time.Millisecond), 50)
		state.CheckKeys(ms)
		state.Update(ms)
		state.UpdateViewport()
		state.Paint(ms)
	}
}
