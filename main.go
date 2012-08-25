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

type State struct {
	system     *twodee.System
	scene      *twodee.Scene
	env        *twodee.Env
	window     *twodee.Window
	char       *twodee.Sprite
	running    bool
	boundaries []*twodee.Sprite
	screenxmin float32
	screenxmax float32
	screenymin float32
	screenymax float32
}

func (s *State) HandleKeys(key, state int) {
	switch key {
	case twodee.KeyEsc:
		s.running = false
	}
}

func (s *State) CheckKeys() {
	var speed float32 = 2048
	switch {
	case s.system.Key(twodee.KeyUp) == 1 && s.system.Key(twodee.KeyDown) == 0:
		s.char.VelocityY = -speed
	case s.system.Key(twodee.KeyUp) == 0 && s.system.Key(twodee.KeyDown) == 1:
		//s.char.VelocityY = speed
	}
	switch {
	case s.system.Key(twodee.KeyLeft) == 1 && s.system.Key(twodee.KeyRight) == 0:
		s.char.VelocityX = -speed
	case s.system.Key(twodee.KeyLeft) == 0 && s.system.Key(twodee.KeyRight) == 1:
		s.char.VelocityX = speed
	default:
		s.char.VelocityX = 0
	}
}

func (s *State) Update(ms float32) {
	s.char.VelocityY += 128
	dX := s.char.VelocityX * ms
	dY := s.char.VelocityY * ms
	for _, b := range s.boundaries {
		if dX != 0 && !s.char.TestMove(dX, 0, b) {
			dX = 0
			s.char.VelocityX = 0
		}
		if dY != 0 && !s.char.TestMove(0, dY, b) {
			dY = 0
			s.char.VelocityY = 0
		}
	}
	s.char.X += dX
	s.char.Y += dY
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
		s.char = s.system.NewSprite("char-textures", 0, 0, 32, 64, 4)
		s.char.Frame = 2
		s.char.X = sprite.X
		s.char.Y = sprite.Y - 64
		sprite.Parent().AddChild(s.char)
		fallthrough
	case FLOOR:
		s.boundaries = append(s.boundaries, sprite)
	}
}

const (
	FLOOR = iota
	START
)

func Init(system *twodee.System) (state *State, err error) {
	var (
		env      *twodee.Env
		textures map[string]string
		opts     twodee.EnvOpts
	)
	state = &State{}
	state.boundaries = make([]*twodee.Sprite, 0)
	state.scene = &twodee.Scene{}
	state.window = &twodee.Window{
		Width:  640,
		Height: 480,
		Title:  "TDoS",
	}
	state.system = system
	state.system.Open(state.window)
	textures = map[string]string{
		"level-textures": "assets/level-textures.png",
		"char-textures":  "assets/char-textures.png",
	}
	for name, path := range textures {
		if err = system.LoadTexture(name, path, twodee.IntNearest); err != nil {
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
		BlockWidth:  32,
		BlockHeight: 32,
		Frames:      2,
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
	state.running = true
	return
}

func (s *State) Running() bool {
	return s.running && s.window.Opened()
}

func (s *State) Paint() {
	s.system.Paint(s.scene)
}

func main() {
	system, err := twodee.Init()
	Check(err)
	defer system.Terminate()

	state, err := Init(system)
	Check(err)
	tick := time.Now()
	for state.Running() {
		ms := float32(time.Since(tick)) / float32(time.Millisecond)
		state.CheckKeys()
		state.Update(ms)
		state.UpdateViewport()
		state.Paint()
		tick = time.Now()
	}
}
