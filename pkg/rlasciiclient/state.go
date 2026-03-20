package rlasciiclient

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/mlge/client"
	"github.com/mechanical-lich/mlge/transport"
	"golang.org/x/image/font/basicfont"
)

// DefaultCellW and DefaultCellH are the pixel dimensions of one character cell
// when using the built-in basicfont (7×13 pixels per glyph).
const (
	DefaultCellW = 7
	DefaultCellH = 13
)

// defaultFace wraps basicfont.Face7x13 for use with Ebiten's text/v2 package.
var defaultFace = text.NewGoXFace(basicfont.Face7x13)

// AsciiClientState implements client.ClientState.
//
// It renders a Cols×Rows viewport of the AsciiWorld as a grid of colored
// characters. The viewport's top-left tile is (CameraX, CameraY) on layer CameraZ.
//
// Games typically embed or wrap this state and update the camera each frame:
//
//	type MyState struct {
//	    ascii *rlasciiclient.AsciiClientState
//	}
//
//	func (s *MyState) Update(snap *transport.Snapshot) client.ClientState {
//	    s.ascii.CameraX = playerX - s.ascii.Cols/2
//	    s.ascii.CameraY = playerY - s.ascii.Rows/2
//	    return nil
//	}
//
//	func (s *MyState) Draw(screen *ebiten.Image) {
//	    s.ascii.DrawViewport(screen)
//	}
type AsciiClientState struct {
	World *AsciiWorld

	// Camera: tile coordinate of the viewport's top-left corner.
	CameraX int
	CameraY int
	CameraZ int

	// Viewport dimensions in character cells.
	Cols int
	Rows int

	// CellW and CellH are the pixel size of each character cell.
	// Defaults to DefaultCellW / DefaultCellH (basicfont 7×13).
	CellW int
	CellH int

	// Background is drawn for cells with no entity. Defaults to black.
	Background color.Color

	// Face is the text face used to render glyphs.
	// Defaults to basicfont.Face7x13 via text/v2.GoXFace.
	// Replace with any text/v2.Face to use a different font.
	Face text.Face
}

// Compile-time assertion.
var _ client.ClientState = (*AsciiClientState)(nil)

// NewAsciiClientState creates a ready-to-use AsciiClientState with sensible defaults.
func NewAsciiClientState(world *AsciiWorld, cols, rows int) *AsciiClientState {
	return &AsciiClientState{
		World:      world,
		Cols:       cols,
		Rows:       rows,
		CellW:      DefaultCellW,
		CellH:      DefaultCellH,
		Background: color.Black,
		Face:       defaultFace,
	}
}

// Done always returns false. Override by wrapping this state.
func (s *AsciiClientState) Done() bool { return false }

// Update satisfies client.ClientState. Does nothing by default — camera and
// game logic should be handled by the wrapping state.
func (s *AsciiClientState) Update(_ *transport.Snapshot) client.ClientState { return nil }

// Draw satisfies client.ClientState and calls DrawViewport.
func (s *AsciiClientState) Draw(screen *ebiten.Image) {
	s.DrawViewport(screen)
}

// DrawViewport renders the current viewport to screen. Call this from a
// wrapping state's Draw method if you need to layer additional UI on top.
func (s *AsciiClientState) DrawViewport(screen *ebiten.Image) {
	cellW := s.CellW
	cellH := s.CellH
	if cellW <= 0 {
		cellW = DefaultCellW
	}
	if cellH <= 0 {
		cellH = DefaultCellH
	}

	face := s.Face
	if face == nil {
		face = defaultFace
	}

	bg := s.Background
	if bg == nil {
		bg = color.Black
	}

	// Build a spatial index for the visible region: tile (x,y) → appearance.
	// Only the first entity at each position is rendered (highest-priority).
	type cell struct {
		char string
		fg   color.RGBA
	}
	cells := make(map[[2]int]cell, len(s.World.Entities))
	for _, e := range s.World.Entities {
		if !e.HasComponent(rlcomponents.Position) || !e.HasComponent(rlcomponents.AsciiAppearance) {
			continue
		}
		pc := e.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
		if pc.Z != s.CameraZ {
			continue
		}
		// Cull to viewport.
		col := pc.X - s.CameraX
		row := pc.Y - s.CameraY
		if col < 0 || col >= s.Cols || row < 0 || row >= s.Rows {
			continue
		}
		key := [2]int{col, row}
		if _, occupied := cells[key]; occupied {
			continue
		}
		ac := e.GetComponent(rlcomponents.AsciiAppearance).(*rlcomponents.AsciiAppearanceComponent)
		cells[key] = cell{
			char: ac.Character,
			fg:   color.RGBA{R: ac.R, G: ac.G, B: ac.B, A: 255},
		}
	}

	// Render each cell.
	for row := 0; row < s.Rows; row++ {
		for col := 0; col < s.Cols; col++ {
			px := float64(col * cellW)
			py := float64(row * cellH)

			// Background.
			ebitenutil.DrawRect(screen, px, py, float64(cellW), float64(cellH), bg)

			// Character (if any entity is here).
			if c, ok := cells[[2]int{col, row}]; ok {
				op := &text.DrawOptions{}
				op.GeoM.Translate(px, py)
				op.ColorScale.ScaleWithColor(c.fg)
				text.Draw(screen, string(c.char), face, op)
			}
		}
	}
}
