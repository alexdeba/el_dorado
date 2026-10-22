package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"eldorado/jeu"
)

// Le plateau : la photo redressée de la carte, mise à l'échelle de l'écran ;
// les cases sont repérées par la grille relevée par outils/decoupe.

const (
	HautPlat  = HauteurEcran
	Echelle   = float64(HautPlat) / jeu.HautImage
	LargPlat  = jeu.LargImage * Echelle
	RayonHexX = jeu.GrilleDX / 2 * Echelle
	RayonHexY = jeu.GrilleDY * 2 / 3 * Echelle
)

// centreMonde : centre d'une case sur le plateau entier à l'échelle 1.
func centreMonde(k jeu.Case) (float64, float64) {
	x := jeu.GrilleX0 + jeu.GrilleDX*float64(k.C)
	if k.R%2 != 0 {
		x += jeu.GrilleDX / 2
	}
	return x * Echelle, (jeu.GrilleY0 + jeu.GrilleDY*float64(k.R)) * Echelle
}

// centre : position à l'écran du centre d'une case, selon la caméra.
func centre(k jeu.Case) (float64, float64) { return cam.ecran(centreMonde(k)) }

// caseSous : la case sous le point (x, y) de l'écran, invalide hors plateau.
func caseSous(x, y float64) jeu.Case {
	if x < 0 || x >= LargPlat || y < 0 || y >= HautPlat {
		return jeu.Case{C: -1, R: -1}
	}
	x, y = cam.monde(x, y)
	r0 := int(math.Round((y/Echelle - jeu.GrilleY0) / jeu.GrilleDY))
	best, bd := jeu.Case{C: -1, R: -1}, math.Inf(1)
	for r := r0 - 1; r <= r0+1; r++ {
		for c := -1; c <= jeu.NbCol; c++ {
			k := jeu.Case{C: c, R: r}
			cx, cy := centreMonde(k)
			if d := math.Hypot(cx-x, cy-y); d < bd {
				best, bd = k, d
			}
		}
	}
	if !best.Valide() {
		return jeu.Case{C: -1, R: -1}
	}
	return best
}

func sommets(k jeu.Case, f float64) [][2]float64 {
	x, y := centre(k)
	rx, ry := RayonHexX*f*cam.Z, RayonHexY*f*cam.Z
	return [][2]float64{{x, y - ry}, {x + rx, y - ry/2}, {x + rx, y + ry/2}, {x, y + ry}, {x - rx, y + ry/2}, {x - rx, y - ry/2}}
}

func dessinerCarte(dst *ebiten.Image) {
	sub := dst.SubImage(image.Rect(0, 0, int(math.Ceil(LargPlat)), HautPlat)).(*ebiten.Image)
	x, y := cam.ecran(0, 0)
	dessinerImage(sub, imgPlateau, x, y, LargPlat*cam.Z, HautPlat*cam.Z, nil)
}

// coupsParCase : la destination de chaque coup ; en jangada, on préfère
// rester sur l'eau quand les deux sont possibles.
func coupsParCase(coups []jeu.Coup) map[jeu.Case]jeu.Coup {
	m := map[jeu.Case]jeu.Coup{}
	for _, c := range coups {
		if c.Mode == jeu.Rester || c.Mode == jeu.Construire {
			continue
		}
		if old, ok := m[c.Vers]; ok && old.Mode == jeu.EnJangada {
			continue
		}
		m[c.Vers] = c
	}
	return m
}

func couleurCoup(c jeu.Coup) color.RGBA {
	switch {
	case c.Mode == jeu.EnJangada:
		return coulEau
	case c.Abandon:
		return coulOrange
	}
	return coulOr2
}

// dessinerPlateau : carte, région centrale, marqueurs, expéditions, coups.
func (g *Game) dessinerPlateau(dst *ebiten.Image) {
	dessinerCarte(dst)
	p := g.p
	e := p.E
	moi := e.Exp[p.Humain]
	t := g.temps()

	// cases candidates d'après les renseignements sûrs
	if !e.DoradoTrouve {
		for _, k := range jeu.Cases() {
			c, ok := jeu.Centrale(k)
			if !ok {
				continue
			}
			if (e.FleuveCentral >= 0 && c.Fleuve != (e.FleuveCentral == 1)) || (e.NatureCentrale != jeu.NatInconnue && c.Nature != e.NatureCentrale) {
				polygone(dst, sommets(k, 0.96), avecAlpha(coulCuir, 0x70))
			} else if e.FleuveCentral >= 0 {
				contour(dst, sommets(k, 0.86), 1.5, avecAlpha(coulOr2, 0xc0))
			}
		}
	}
	if k, ok := e.Dorado(); ok && e.CodesReveles {
		a := 0.6 + 0.4*math.Sin(t*3)
		contour(dst, sommets(k, 0.9), 4, avecAlpha(coulOr2, uint8(255*a)))
		if !e.DoradoTrouve {
			ecrireHalo(dst, "El Dorado", centreX(k)-28, centreY(k)+10, face(13, policeGras), coulOr2, coulCuir, 1.5)
		}
	}

	// coups possibles
	if p.Req.Type == ReqDeplacer {
		parCase := coupsParCase(p.Req.Coups)
		a := uint8(0x50 + 0x28*math.Sin(t*4))
		for k, c := range parCase {
			col := couleurCoup(c)
			polygone(dst, sommets(k, 0.92), avecAlpha(col, a))
			contour(dst, sommets(k, 0.86), 6, avecAlpha(coulCuir, 0xd0))
			contour(dst, sommets(k, 0.86), 3, col)
			x, y := centre(k)
			disque(dst, x, y, 6, coulCuir)
			disque(dst, x, y, 4, col)
		}
		contour(dst, sommets(moi.Pos, 0.98), 3, coulBlanc)
		if c, ok := parCase[g.survol]; ok {
			x0, y0 := centre(moi.Pos)
			for _, k := range c.Chemin {
				x1, y1 := centre(k)
				ligne(dst, x0, y0, x1, y1, 3, coulSang)
				disque(dst, x1, y1, 4, coulSang)
				x0, y0 = x1, y1
			}
		}
	}

	// trésor abandonné
	if e.TresorAuSol {
		x, y := centre(e.CaseTresor)
		coffre(dst, x+10*cam.Z, y-14*cam.Z, cam.Z)
	}

	// pions-natifs
	for k, m := range e.Marqueurs {
		g.pionNatif(dst, k, m)
	}

	// expéditions
	for i, x := range e.Exp {
		if !x.Vivante {
			continue
		}
		g.pionExpedition(dst, i, x, t)
	}

	g.dessinerMini(dst)
	if p.Req.Type == ReqDeplacer && g.survol.Valide() {
		g.infobulle(dst)
	} else if g.survol.Valide() && !g.regles && !g.feuille {
		g.infobulle(dst)
	}
}

func centreX(k jeu.Case) float64 { x, _ := centre(k); return x }
func centreY(k jeu.Case) float64 { _, y := centre(k); return y }

// pionNatif : le pion-natif de l'encart, numéroté ; vert s'il est ami,
// barré s'il n'y a plus de natifs.
func (g *Game) pionNatif(dst *ebiten.Image, k jeu.Case, m *jeu.Marqueur) {
	x, y := centre(k)
	s := 24 * cam.Z
	px, py := x-s/2, y-s/2
	if g.p.E.Occupant(k, -1) >= 0 {
		px, py = x+6*cam.Z, y-s-6*cam.Z // à côté du pion-expédition
	}
	rect(dst, px+2, py+2, s, s, avecAlpha(coulNoir, 0x80))
	cs := &ebiten.ColorScale{}
	if m.Exterminee {
		cs.Scale(0.55, 0.5, 0.45, 1)
	}
	dessinerImage(dst, imgNatif, px, py, s, s, cs)
	bord := coulCuir
	if m.Amis[g.p.Humain] {
		bord = coulTurquoise
	}
	cadre(dst, px, py, s, s, 2, bord)
	// numéro (cache celui imprimé sur le pion)
	n := fmt.Sprint(m.Numero)
	f := face(math.Round(11*min(cam.Z, 2)), policeGras)
	w := largeur(n, f) + 5
	hn := 13 * min(cam.Z, 2)
	rect(dst, px+s-w+1, py+s-hn+1, w, hn, coulParch)
	ecrire(dst, n, px+s-w+3.5, py+s-hn, f, coulEncre)
	if m.Exterminee {
		ligne(dst, px+2, py+2, px+s-2, py+s-2, 2.5, coulSang)
		ligne(dst, px+s-2, py+2, px+2, py+s-2, 2.5, coulSang)
	}
	if m.Absents >= 0 {
		rect(dst, px, py, s, s, avecAlpha(coulParch, 0x90))
	}
}

// pionExpedition : le conquistador de l'encart, aux couleurs de l'expédition.
func (g *Game) pionExpedition(dst *ebiten.Image, i int, x *jeu.Expedition, t float64) {
	a, ok := g.affiche[i]
	if !ok {
		a[0], a[1] = g.positionAffichee(i)
	}
	px, py := cam.ecran(a[0], a[1])
	s := 34 * cam.Z
	if i == g.p.E.Courant && !g.p.E.Fini {
		cercle(dst, px, py, s*0.78+2*math.Sin(t*5), 3, coulBlanc)
	}
	if x.Tresor {
		disque(dst, px, py, s*0.8, avecAlpha(coulOr2, uint8(0x90+0x50*math.Sin(t*4))))
	}
	rect(dst, px-s/2+2, py-s/2+3, s, s, avecAlpha(coulNoir, 0x90))
	dessinerImage(dst, imgExpeditions[x.Couleur], px-s/2, py-s/2, s, s, nil)
	bord := coulCuir
	if i == g.p.Humain {
		bord = coulBlanc
	}
	cadre(dst, px-s/2, py-s/2, s, s, 2, bord)
	if x.Jangada {
		radeau(dst, px, py+s/2+2, cam.Z)
	}
	if x.Tresor {
		coffre(dst, px+s/2-2, py-s/2-4, 0.8*cam.Z)
	}
	if x.Retard > 0 {
		ecrireHalo(dst, "zzz", px+s/2-6, py-s/2-14, face(12, policeItalique), coulParch, coulCuir, 1.2)
	}
}

// radeau : quelques troncs liés, sous le pion.
func radeau(dst *ebiten.Image, x, y, s float64) {
	rect(dst, x-19*s, y-s, 38*s, 9*s, coulCuir)
	for i := range 5 {
		rect(dst, x+(-18+float64(i)*7.4)*s, y, 6.6*s, 7*s, color.RGBA{0xa8, 0x7a, 0x42, 0xff})
	}
	ligne(dst, x-19*s, y+3.5*s, x+19*s, y+3.5*s, 1.2*s, coulCuir)
}

// coffre : le trésor d'El Dorado.
func coffre(dst *ebiten.Image, x, y, s float64) {
	w, h := 18*s, 13*s
	rect(dst, x-w/2+1, y-h/2+2, w, h, avecAlpha(coulNoir, 0x90))
	rect(dst, x-w/2, y-h/2, w, h, coulOr)
	rect(dst, x-w/2, y-h/2, w, h*0.38, coulOr2)
	cadre(dst, x-w/2, y-h/2, w, h, 1.5, coulOrSombre)
	rect(dst, x-1.5*s, y-h*0.15, 3*s, 4*s, coulOrSombre)
}

// infobulle : ce que l'on sait de la case survolée.
func (g *Game) infobulle(dst *ebiten.Image) {
	e := g.p.E
	k := g.survol
	if !jeu.Terre(k) {
		return
	}
	var l []string
	titre := strings.ToUpper(jeu.TerrainDe(k).String()[:1]) + jeu.TerrainDe(k).String()[1:]
	if v := jeu.VilleEn(k); v >= 0 {
		titre = jeu.Villes[v].Nom
		l = append(l, "Ville de départ ("+jeu.TerrainDe(k).String()+")")
	}
	if k == jeu.Centre {
		titre = "La Montagne sacrée"
	}
	if c, ok := jeu.Centrale(k); ok {
		f := "sans fleuve"
		if c.Fleuve {
			f = "avec fleuve"
		}
		l = append(l, fmt.Sprintf("Région centrale, zone %s : %s, %s", c.Zone, c.Nature, f))
	} else if jeu.Fleuve(k) {
		s := "Fleuve navigable"
		if jeu.Cascade(k) {
			s += ", cascade"
		}
		l = append(l, s)
	}
	if !jeu.DansRegion(k) {
		l = append(l, fmt.Sprintf("%d cases de la case centrale", jeu.Distance(k, jeu.Centre)))
	}
	if m := e.Marqueurs[k]; m != nil {
		s := fmt.Sprintf("Pion-natif n° %d : %d natifs", m.Numero, m.Natifs)
		switch {
		case m.Exterminee:
			s = fmt.Sprintf("Pion-natif n° %d : tribu exterminée", m.Numero)
		case m.Amis[g.p.Humain]:
			s += ", amis"
		}
		l = append(l, s)
	}
	for i, x := range e.Exp {
		if x.Vivante && x.Pos == k {
			s := x.Nom
			if i == g.p.Humain {
				s = "Vous"
			}
			if x.Tresor {
				s += " (porte le trésor)"
			}
			l = append(l, s)
		}
	}
	if e.TresorAuSol && e.CaseTresor == k {
		l = append(l, "Le trésor d'El Dorado gît ici")
	}
	if g.p.Req.Type == ReqDeplacer {
		if c, ok := coupsParCase(g.p.Req.Coups)[k]; ok {
			switch {
			case c.Mode == jeu.EnJangada:
				l = append(l, fmt.Sprintf("Clic : en jangada (%d case%s)", len(c.Chemin), pl(len(c.Chemin))))
			case c.Abandon:
				l = append(l, "Clic : à pied, la jangada est abandonnée")
			default:
				l = append(l, "Clic : à pied")
			}
		}
	}
	ft := face(17, policeGras)
	fn := face(15, policeTexte)
	w := largeur(titre, ft)
	for _, s := range l {
		w = max(w, largeur(s, fn))
	}
	w += 22
	h := 30 + 19*float64(len(l))
	x, y := centre(k)
	x += 26
	y -= h / 2
	if x+w > LargPlat-4 {
		x -= w + 52
	}
	y = min(max(y, 4), HautPlat-h-4)
	rect(dst, x+3, y+3, w, h, avecAlpha(coulNoir, 0x70))
	rect(dst, x, y, w, h, coulParch)
	cadre(dst, x, y, w, h, 1.5, coulOrSombre)
	ecrire(dst, titre, x+11, y+4, ft, coulSang)
	for i, s := range l {
		ecrire(dst, s, x+11, y+26+19*float64(i), fn, coulEncre)
	}
}

func pl(n int) string {
	if n > 1 {
		return "s"
	}
	return ""
}
