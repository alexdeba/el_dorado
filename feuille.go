package main

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"eldorado/jeu"
)

// La feuille commune des renseignements (touche F) : tout le monde voit le
// renseignement, la distance et le dé brut ; chacun ne connaît que sa propre
// fiabilité, jusqu'à ce qu'une annonce dévoile les codes secrets.

func (g *Game) dessinerFeuille(dst *ebiten.Image) {
	e := g.p.E
	rect(dst, 0, 0, LargPlat, HautPlat, avecAlpha(coulCuir, 0xb0))
	x, y, w, h := 40.0, 40.0, LargPlat-80, 870.0
	carte(dst, x, y, w, h)
	ecrire(dst, "Feuille commune des renseignements", x+30, y+18, face(30, policeTitre), coulSang)
	sous := "Vous seul connaissez la fiabilité de vos renseignements."
	if e.CodesReveles {
		sous = "Les codes secrets ont été dévoilés : toutes les fiabilités sont connues."
	}
	ecrire(dst, sous, x+30, y+64, face(18, policeItalique), coulSepia)
	cols := []float64{x + 30, x + 100, x + 300, x + 600, x + 730, x + 830}
	ty := y + 104
	rect(dst, x+20, ty-4, w-40, 28, coulJungle)
	for i, t := range []string{"Tour", "Expédition", "Renseignement", "Distance", "Dé", "Fiabilité"} {
		ecrire(dst, t, cols[i], ty, face(17, policeGras), coulParch)
	}
	fn := face(17, policeTexte)
	ty += 30
	debut := max(0, len(e.Feuille)-26)
	if debut > 0 {
		ecrire(dst, fmt.Sprintf("… %d renseignements plus anciens", debut), cols[0], ty, face(15, policeItalique), coulSepia)
		ty += 24
	}
	for _, r := range e.Feuille[debut:] {
		nom := e.Exp[r.Joueur].Nom
		if r.Joueur == g.p.Humain {
			nom = "Vous"
		}
		c := coulEncre
		if r.Joueur == g.p.Humain {
			c = coulJungle
		}
		rect(dst, cols[1]-14, ty+4, 8, 14, coulExpeditions[e.Exp[r.Joueur].Couleur])
		ecrire(dst, fmt.Sprint(r.Tour), cols[0], ty, fn, c)
		ecrire(dst, nom, cols[1], ty, fn, c)
		txt := r.Texte()
		if r.Type != jeu.RensZone {
			txt = "région centrale : " + txt
		}
		ecrire(dst, txt, cols[2], ty, fn, c)
		ecrire(dst, fmt.Sprint(r.Distance), cols[3], ty, fn, c)
		ecrire(dst, fmt.Sprint(r.De), cols[4], ty, fn, c)
		fiab := "?"
		if f, ok := e.FiabiliteConnue(g.p.Humain, r); ok {
			fiab = fmt.Sprintf("%d %%", f)
		}
		ecrire(dst, fiab, cols[5], ty, face(17, policeGras), c)
		ligne(dst, x+20, ty+24, x+w-20, ty+24, 1, coulParch2)
		ty += 26
	}
	if len(e.Feuille) == 0 {
		ecrire(dst, "Aucun renseignement pour l'instant : les tribus amies ou prisonnières en donnent.", cols[0], ty+6, face(18, policeItalique), coulSepia)
	}
	// synthèse
	sy := y + h - 66
	ligne(dst, x+20, sy-10, x+w-20, sy-10, 1.5, coulOrSombre)
	ecrire(dst, g.resumeRenseignements(), x+30, sy, face(18, policeGras), coulEncre)
	ecrire(dst, g.resumeCentral(), x+30, sy+26, face(18, policeGras), coulEncre)
	ecrireDroite(dst, g.estimation(), x+w-30, sy, face(18, policeGras), coulJungle)
	if k, ok := e.Dorado(); ok && e.CodesReveles {
		ecrireDroite(dst, fmt.Sprintf("El Dorado : case %d,%d (zone %s)", k.C, k.R, e.ZoneVraie), x+w-30, sy+26, face(18, policeGras), coulSang)
	}
}

// estimation : probabilité de chaque zone d'après la feuille, vue du joueur
// (les codes secrets des autres sont inconnus).
func (g *Game) estimation() string {
	c := g.p.E.Croyance(g.p.Humain)
	if c[jeu.SansZone] == 1 {
		return "Aucune zone n'est encore établie"
	}
	pc := func(v float64) int { return int(v*100 + .5) }
	s := fmt.Sprintf("Zone établie : Ouest %d %%, Est %d %%, Sud %d %%", pc(c[jeu.Ouest]), pc(c[jeu.Est]), pc(c[jeu.Sud]))
	if c[jeu.SansZone] > 0.005 {
		s += fmt.Sprintf(", pas encore %d %%", pc(c[jeu.SansZone]))
	}
	return s
}
