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

func (s *State) CheckKeys(us float32) {
	var speed float32 = 20
	var minspeed float32 = 5
	var accel float32 = 0.01 * us
	var decel float32 = 0.5 * us
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

func (s *State) Update(us float32) {
	s.textfps.SetText(fmt.Sprintf("%6.0f FPS", (1000000.0 / us)))

	s.char.VelocityY += 5
	dX := Round(s.char.VelocityX * us)
	dY := Round(s.char.VelocityY * us)
	for _, b := range s.boundaries {
		if dX != 0 && !s.char.TestMove(dX, 0, b) {
			if s.char.TestMove(dX, float32(-b.Height), b) {
				s.char.Y -= float32(b.Height)
			} else {
				if dX < 0 {
					s.char.X = b.X + float32(b.Width)
				} else {
					s.char.X = b.X - float32(s.char.Width)
				}
				s.char.VelocityX = 0
				dX = 0
			}
		}
		if dY != 0 && !s.char.TestMove(0, dY, b) {
			if dY < 0 {
				s.char.Y = b.Y + float32(b.Height)
			} else {
				s.char.Y = b.Y - float32(s.char.Height)
			}
			s.char.VelocityY = 0
			dY = 0
		}
	}
	s.char.X = Round(s.char.X + dX)
	s.char.Y = Round(s.char.Y + dY)
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

func (s *State) HandleAddBlock(sprite *twodee.Sprite, block *twodee.EnvBlock) {
	switch block.Type {
	case START:
		s.char = s.system.NewSprite("char-textures", 0, 0, 32, 64)
		s.char.SetFrame(2)
		s.char.X = sprite.X
		s.char.Y = sprite.Y - 100
		sprite.Parent().AddChild(s.char)
		fallthrough
	case FLOOR:
		s.boundaries = append(s.boundaries, sprite)
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
)

type TexInfo struct {
	Name  string
	Path  string
	Width int
}

func Init(system *twodee.System) (state *State, err error) {
	var (
		env      *twodee.Env
		opts     twodee.EnvOpts
	)
	state = &State{}
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
		TexInfo{"level-textures", "assets/level-textures.png", 8},
		TexInfo{"char-textures", "assets/char-textures.png", 16},
		TexInfo{"font1-textures", "assets/font1-textures.png", 0},
	}
	for _, t := range textures {
		if err = system.LoadTexture(t.Name, t.Path, twodee.IntNearest, t.Width); err != nil {
			return
		}
	}
	BlockHandler := func(sprite *twodee.Sprite, block *twodee.EnvBlock) {
		state.HandleAddBlock(sprite, block)
	}
	opts = twodee.EnvOpts{
		Blocks: []*twodee.EnvBlock{
			&twodee.EnvBlock{
				Color:      color.RGBA{153, 102, 0, 255},
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
	state.textfps = system.NewText("font1-textures", 0, float32(state.window.Height - 16), 1, "")
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
		us := Min(float32(time.Since(tick))/float32(time.Microsecond), 1)
		state.CheckKeys(us)
		state.Update(us)
		state.UpdateViewport()
		state.Paint(us)
		tick = time.Now()
	}
}
