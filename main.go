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
	"os"
)

func Check(err error) {
	if err != nil {
		fmt.Printf("[error]: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	var (
		err    error
		system *twodee.System
		window *twodee.Window
		scene  *twodee.Scene
		env    *twodee.Environment
	)
	system, err = twodee.Init()
	Check(err)
	defer system.Terminate()

	window = &twodee.Window{
		Width:  640,
		Height: 480,
		Title:  "The Destiny of Species",
	}
	err = system.Open(window)
	Check(err)

	textures := map[string]string{
		"level-textures": "assets/level-textures.png",
	}
	for name, path := range textures {
		err = system.LoadTexture(name, path, twodee.IntNearest)
		Check(err)
	}

	scene = &twodee.Scene{}
	env, err = system.LoadEnvironment("level-textures", "assets/level-fw.png")
	Check(err)
	scene.AddChild(env)
	run := true
	var keystep float32 = 32.0
	for run {
		system.Paint(scene)
		if system.Key(twodee.KeyUp) == 1 { env.Y -= keystep }
		if system.Key(twodee.KeyDown) == 1 { env.Y += keystep }
		if system.Key(twodee.KeyLeft) == 1 { env.X -= keystep }
		if system.Key(twodee.KeyRight) == 1 { env.X += keystep }
		run = system.Key(twodee.KeyEsc) == 0 && window.Opened()
	}
}
