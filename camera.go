package main

import (
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"eldorado/jeu"
)

// La caméra du plateau : zoom et cadrage. X, Y : coin haut gauche de la vue,
// en coordonnées du plateau entier affiché à l'échelle 1 (LargPlat × HautPlat).

type Camera struct{ Z, X, Y float64 }

const (
	ZoomMax    = 4.0
	ZoomEspace = 2.5 // zoom de la barre d'espace
)

var cam = Camera{Z: 1}

func (c *Camera) borner() {
	c.Z = min(max(c.Z, 1), ZoomMax)
	c.X = min(max(c.X, 0), LargPlat-LargPlat/c.Z)
	c.Y = min(max(c.Y, 0), HautPlat-HautPlat/c.Z)
}

// ecran : un point du plateau vers l'écran.
func (c *Camera) ecran(x, y float64) (float64, float64) { return (x - c.X) * c.Z, (y - c.Y) * c.Z }

// monde : un point de l'écran vers le plateau.
func (c *Camera) monde(x, y float64) (float64, float64) { return c.X + x/c.Z, c.Y + y/c.Z }

// zoomer : multiplie le zoom par f, le point (sx, sy) de l'écran restant fixe.
func (c *Camera) zoomer(f, sx, sy float64) {
	wx, wy := c.monde(sx, sy)
	c.Z = min(max(c.Z*f, 1), ZoomMax)
	c.X, c.Y = wx-sx/c.Z, wy-sy/c.Z
	c.borner()
}

// centrer : la vue centrée sur le point (x, y) du plateau.
func (c *Camera) centrer(x, y float64) {
	c.X, c.Y = x-LargPlat/c.Z/2, y-HautPlat/c.Z/2
	c.borner()
}

// cameraActive : le plateau est visible et rien ne le couvre.
func (g *Game) cameraActive() bool {
	switch {
	case g.regles || g.ecran == EcranAccueil:
		return false
	case g.ecran == EcranJeu:
		return !g.feuille && !g.p.Termine && !g.p.E.Fini
	}
	return true
}

// gererCamera : molette, glisser du bouton droit (ou du milieu), flèches,
// + et −, espace (zoom ou carte entière) ; clic ou glisser sur la mini-carte.
// Vrai si la souris est sur la mini-carte : le plateau dessous ne doit pas
// réagir.
func (g *Game) gererCamera() bool {
	mx, my := float64(g.mx), float64(g.my)
	surPlateau := mx >= 0 && mx < LargPlat && my >= 0 && my < HautPlat
	if _, dy := ebiten.Wheel(); dy != 0 && surPlateau {
		cam.zoomer(math.Pow(1.2, dy), mx, my)
	}
	for _, r := range ebiten.AppendInputChars(nil) {
		switch r {
		case '+':
			cam.zoomer(1.5, LargPlat/2, HautPlat/2)
		case '-':
			cam.zoomer(1/1.5, LargPlat/2, HautPlat/2)
		case '0':
			cam = Camera{Z: 1}
		}
	}
	// espace : zoom autour du curseur (ou de votre pion), ou retour à la carte entière
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		switch {
		case cam.Z > 1:
			cam = Camera{Z: 1}
		case surPlateau:
			cam.zoomer(ZoomEspace, mx, my)
		default:
			cam.Z = ZoomEspace
			if g.p != nil {
				cam.centrer(centreMonde(g.p.E.Exp[g.p.Humain].Pos))
			} else {
				cam.centrer(LargPlat/2, HautPlat/2)
			}
		}
	}
	pas := 14 / cam.Z
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		cam.X -= pas
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		cam.X += pas
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		cam.Y -= pas
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		cam.Y += pas
	}
	cam.borner()

	// glisser le plateau
	tire := ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) || ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle)
	if surPlateau && (inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonMiddle)) {
		g.glisse = true
	} else if !tire {
		g.glisse = false
	}
	if g.glisse {
		cam.X -= (mx - g.gx) / cam.Z
		cam.Y -= (my - g.gy) / cam.Z
		cam.borner()
	}
	g.gx, g.gy = mx, my

	// mini-carte
	if !miniVisible() {
		g.miniTire = false
		return false
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && dansMini(mx, my) {
		g.miniTire = true
	} else if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		g.miniTire = false
	}
	if g.miniTire {
		cam.centrer((mx-miniX)/miniW*LargPlat, (my-miniY)/miniH*HautPlat)
	}
	return g.miniTire || dansMini(mx, my)
}

// boutonsZoom : en haut à gauche du plateau.
func (g *Game) boutonsZoom() {
	b := g.bouton(14, 14, 44, 44, "+", cam.Z < ZoomMax, func() { cam.zoomer(1.5, LargPlat/2, HautPlat/2) })
	b.Taille = 28
	b = g.bouton(64, 14, 44, 44, "−", cam.Z > 1, func() { cam.zoomer(1/1.5, LargPlat/2, HautPlat/2) })
	b.Taille = 28
	b = g.bouton(114, 14, 104, 44, "Tout voir", cam.Z > 1, func() { cam = Camera{Z: 1} })
	b.Taille = 19
}

// ---------------------------------------------------------------- mini-carte

const (
	miniW = 250.0
	miniH = miniW * HautPlat / LargPlat
	miniX = 14.0
	miniY = HautPlat - miniH - 14
)

func miniVisible() bool { return cam.Z > 1.001 }

func dansMini(x, y float64) bool {
	return x >= miniX && x < miniX+miniW && y >= miniY && y < miniY+miniH
}

// versMini : un point du plateau sur la mini-carte.
func versMini(x, y float64) (float64, float64) {
	return miniX + x/LargPlat*miniW, miniY + y/HautPlat*miniH
}

// dessinerMini : le plateau entier en réduction, la vue courante encadrée,
// les expéditions (et les villes pendant l'organisation).
func (g *Game) dessinerMini(dst *ebiten.Image) {
	if !miniVisible() {
		return
	}
	rect(dst, miniX-4+3, miniY-4+4, miniW+8, miniH+8, avecAlpha(coulNoir, 0x90))
	rect(dst, miniX-4, miniY-4, miniW+8, miniH+8, coulCuir)
	dessinerImage(dst, imgPlateau, miniX, miniY, miniW, miniH, nil)
	cadre(dst, miniX-4, miniY-4, miniW+8, miniH+8, 2, coulOrSombre)

	// hors de la vue : assombri
	vx, vy := versMini(cam.X, cam.Y)
	vw, vh := miniW/cam.Z, miniH/cam.Z
	voile := avecAlpha(coulCuir, 0x78)
	rect(dst, miniX, miniY, miniW, vy-miniY, voile)
	rect(dst, miniX, vy+vh, miniW, miniY+miniH-vy-vh, voile)
	rect(dst, miniX, vy, vx-miniX, vh, voile)
	rect(dst, vx+vw, vy, miniX+miniW-vx-vw, vh, voile)

	if g.ecran == EcranOrganisation {
		for i, v := range jeu.Villes {
			x, y := versMini(centreMonde(v.Case))
			c := coulOr2
			if i == g.ville {
				c = coulSang
			}
			disque(dst, x, y, 4.5, coulCuir)
			disque(dst, x, y, 3, c)
		}
	} else if g.p != nil {
		e := g.p.E
		if e.TresorAuSol {
			x, y := versMini(centreMonde(e.CaseTresor))
			disque(dst, x, y, 4, coulOr2)
		}
		for i, x := range e.Exp {
			if !x.Vivante {
				continue
			}
			px, py := versMini(centreMonde(x.Pos))
			if i == g.p.Humain {
				disque(dst, px, py, 6.5, coulBlanc)
			}
			disque(dst, px, py, 5, coulCuir)
			disque(dst, px, py, 3.8, coulExpeditions[x.Couleur])
		}
	}
	sub := dst.SubImage(image.Rect(int(miniX), int(miniY), int(miniX+miniW), int(miniY+miniH))).(*ebiten.Image)
	cadre(sub, vx-1, vy-1, vw+2, vh+2, 3.5, avecAlpha(coulCuir, 0xc0))
	cadre(sub, vx, vy, vw, vh, 2, coulBlanc)
}
