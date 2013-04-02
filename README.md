The Destiny of Species
======================

The year is 1835.  Charles Darwin has just reached the Galapagos, only to find
that the island hides a terrifying secret...

LD48 submission here: http://www.ludumdare.com/compo/ludum-dare-24/?action=preview&uid=7913

Builds available for OSX and Windows

Please contact twitter.com/kurrik with feedback, especially if it doesn't work!

Building
--------
Install deps, as listed in https://github.com/kurrik/twodee

    make init
    make run

Tasks
-----
* Load a level and construct a scene (DONE)
* Actor class for monsters, player (DONE)
* Player controls (DONE)
* Collision detection (DONE)
* Timer (DONE)
* Screen overlay (DONE)
* Physics (DONE)
* Monsters (DONE)
* Collisions (DONE)
* Points (DONE)
* Death (DONE)
* 1ups (DONE)
* Better sprites (DONE)
* Hit points (DONE)
* Death animation (DONE)
* Tune controls (DONE)
* Ending? (DONE)
* Don't jump forever (DONE)
* Enemies spawn others (DONE)
* Bigger map
* Music
* Parallax

Bugs
----
* Player z-index (FIXED)

Ideas (spoilers!)
-----------------
* Powerup by losing a life
  * This will make the mechanic mostly around trying to get 1ups
  * Start with none, so the player won't be so powerful
  * On your 10th life or so you should be pretty awesome
* Obviously monsters need to spawn a lot of small monsters
* Fight through increasingly complex monsters
* Giant boss monster, spawns tons of small monsters

Building on Windows
-------------------
64bit!  http://tdm-gcc.tdragon.net/

Seems to work from the "Git Bash" env
Build GLEW
http://stackoverflow.com/questions/6005076/building-glew-on-windows-with-mingw
Copy GLEW includes to C:\Mingw64 includes
Copy GLEW /lib output (.a files!) to C:\MinGW64\lib
Copy GLEW DLLs to C:\Windows\System32

Get GLFW
Binary dist
Copy glfw-2.7.6.bin.WIN32/lib-mingw/x64 stuff  to C:\MinGW64\lib
Copy dist include to gcc include
Copy glfw.dll to C:\Windows\System32

go get github.com/banthar/gl

go get github.com/metaleap/go-glfwdll-win64
//go get github.com/jteeuwen/glfw


Ffffffff http://stackoverflow.com/questions/10369513/dll-linking-via-windows-cgo-gcc-ld-gives-undefined-reference-to-function-e
