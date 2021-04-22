// SPDX-License-Identifier: Apache-2.0
// Copyright 2020,2021 Philipp Naumann, Marcus Soll
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"math/rand"
	"sync"
)

func init() {
	err := RegisterAI("BadRandomAI", func() AI { return new(BadRandomAI) })
	if err != nil {
		panic(err)
	}
}

// BadRandomAI is an AI that performs random actions. It explicitly does not try to avoid crashes in others, it only avoids crashes in existing filled cells.
type BadRandomAI struct {
	l sync.Mutex
	i chan string
}

// GetChannel receives the answer channel.
func (r *BadRandomAI) GetChannel(c chan string) {
	r.l.Lock()
	defer r.l.Unlock()
	r.i = c
}

// GetState gets the game state and computes an answer.
func (r *BadRandomAI) GetState(g *Game) {
	r.l.Lock()
	defer r.l.Unlock()

	if r.i == nil {
		return
	}

	if g.Running {
		// actions
		actions := []string{ActionTurnLeft, ActionTurnRight, ActionSlower, ActionFaster, ActionNOOP}
		rand.Shuffle(len(actions), func(i, j int) { actions[i], actions[j] = actions[j], actions[i] })

		// test actions
		for i := range actions {
			// do action
			switch actions[i] {
			case ActionTurnLeft:
				switch g.Players[g.You].Direction {
				case DirectionLeft:
					g.Players[g.You].Direction = DirectionDown
				case DirectionRight:
					g.Players[g.You].Direction = DirectionUp
				case DirectionUp:
					g.Players[g.You].Direction = DirectionLeft
				case DirectionDown:
					g.Players[g.You].Direction = DirectionRight
				}
			case ActionTurnRight:
				switch g.Players[g.You].Direction {
				case DirectionLeft:
					g.Players[g.You].Direction = DirectionUp
				case DirectionRight:
					g.Players[g.You].Direction = DirectionDown
				case DirectionUp:
					g.Players[g.You].Direction = DirectionRight
				case DirectionDown:
					g.Players[g.You].Direction = DirectionLeft
				}
			case ActionFaster:
				g.Players[g.You].Speed++
				if g.Players[g.You].Speed > MaxSpeed {
					g.Players[g.You].Speed--
					continue
				}
			case ActionSlower:
				g.Players[g.You].Speed--
				if g.Players[g.You].Speed < 1 {
					g.Players[g.You].Speed++
					continue
				}
			case ActionNOOP:
				// Do nothing
			default:
				log.Println("bad random ai:", "unknown action", actions[i])
			}

			// test
			if !r.willCrash(g) {
				select {
				case r.i <- actions[i]:
				default:
				}
				return
			}

			// undo action
			switch actions[i] {
			case ActionTurnLeft:
				switch g.Players[g.You].Direction {
				case DirectionLeft:
					g.Players[g.You].Direction = DirectionUp
				case DirectionRight:
					g.Players[g.You].Direction = DirectionDown
				case DirectionUp:
					g.Players[g.You].Direction = DirectionRight
				case DirectionDown:
					g.Players[g.You].Direction = DirectionLeft
				}
			case ActionTurnRight:
				switch g.Players[g.You].Direction {
				case DirectionLeft:
					g.Players[g.You].Direction = DirectionDown
				case DirectionRight:
					g.Players[g.You].Direction = DirectionUp
				case DirectionUp:
					g.Players[g.You].Direction = DirectionLeft
				case DirectionDown:
					g.Players[g.You].Direction = DirectionRight
				}
			case ActionFaster:
				g.Players[g.You].Speed--
			case ActionSlower:
				g.Players[g.You].Speed++
			case ActionNOOP:
				// Do nothing
			}
		}

		// no valid actions - pick random
		select {
		case r.i <- actions[0]:
		default:
		}
		return
	}
}

// willCrash computes whether the given game state will result in a (possible) crash.
// Not safe for concurrent use on the same game.
func (r *BadRandomAI) willCrash(g *Game) bool {
	oldX, oldY := g.Players[g.You].X, g.Players[g.You].Y
	defer func() {
		g.Players[g.You].X, g.Players[g.You].Y = oldX, oldY
	}()

	var dostep func(x, y int) (int, int)
	switch g.Players[g.You].Direction {
	case DirectionUp:
		dostep = func(x, y int) (int, int) { return x, y - 1 }
	case DirectionDown:
		dostep = func(x, y int) (int, int) { return x, y + 1 }
	case DirectionLeft:
		dostep = func(x, y int) (int, int) { return x - 1, y }
	case DirectionRight:
		dostep = func(x, y int) (int, int) { return x + 1, y }
	}

	for s := 0; s < g.Players[g.You].Speed; s++ {
		g.Players[g.You].X, g.Players[g.You].Y = dostep(g.Players[g.You].X, g.Players[g.You].Y)
		if g.Players[g.You].X < 0 || g.Players[g.You].X >= g.Width || g.Players[g.You].Y < 0 || g.Players[g.You].Y >= g.Height {
			return true
		}
		if g.Players[g.You].Speed >= HoleSpeed && (g.Players[g.You].stepCounter+1)%HolesEachStep == 0 && s != 0 && s != g.Players[g.You].Speed-1 {
			continue
		}
		if g.Cells[g.Players[g.You].Y][g.Players[g.You].X] != 0 {
			return true
		}
	}

	return false
}

// Name returns the name of the AI.
func (r *BadRandomAI) Name() string {
	return "BadRandomAI"
}
