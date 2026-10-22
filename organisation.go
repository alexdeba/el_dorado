package main

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"eldorado/jeu"
)

// Organisation de l'expédition : 1 000 sols d'or, une ville de départ, un
// code secret, la couleur du pion.

func (g *Game) choisirVille(v int) {
	g.ville = v
	if !jeu.Navigable(jeu.Villes[v].Case) && !jeu.Cotiere(jeu.Villes[v].Case) {
		g.achats.EnJangada = false
	}
}

func (g *Game) consigneOrganisation() string {
	switch {
	case g.ville < 0:
		return "Cliquez sur la carte la ville d'où partira votre expédition (cerclée d'or)."
	case g.achats.Cout() > jeu.Budget:
		return "Vos achats dépassent les 1 000 sols d'or : retirez quelque chose."
	case g.hommes() == 0:
		return "Il faut au moins un homme dans l'expédition."
	}
	return "Composez votre expédition, choisissez votre code secret, puis « En route ! »."
}

func (g *Game) hommes() int {
	n := 0
	for _, v := range g.achats.Membres {
		n += v
	}
	return n
}

var articles = []struct {
	nom, aide string
	prix      int
}{
	{"Arquebusiers", "se battent contre les natifs hostiles", jeu.PrixMembre},
	{"Médecins", "évitent les retards des maladies", jeu.PrixMembre},
	{"Jésuites", "évitent les retards dus aux causes morales", jeu.PrixMembre},
	{"Guides", "évitent les retards dus à l'égarement", jeu.PrixMembre},
	{"Cadeaux", "+1 au dé d'attitude par cadeau offert", jeu.PrixCadeau},
	{"Jangadas", "radeau, construit au départ ou en matériel", jeu.PrixJangada},
}

func (g *Game) quantite(i int) *int {
	switch i {
	case 4:
		return &g.achats.Cadeaux
	case 5:
		return &g.achats.Jangadas
	}
	return &g.achats.Membres[i]
}

const orgY = 150.0

func (g *Game) dessinerOrganisation(dst *ebiten.Image) {
	dessinerCarte(dst)
	t := g.temps()
	for i, v := range jeu.Villes {
		x, y := centre(v.Case)
		r := 22 * cam.Z
		if i == g.ville {
			disque(dst, x, y, r+6, avecAlpha(coulOr2, 0x80))
			cercle(dst, x, y, r+6, 4, coulSang)
		} else {
			cercle(dst, x, y, r+2*math.Sin(t*4), 3, coulOr2)
		}
	}
	if v := jeu.VilleEn(g.survol); v >= 0 {
		x, y := centre(g.survol)
		cercle(dst, x, y, 30*cam.Z, 2, coulBlanc)
		f := face(20, policeGras)
		ecrireHalo(dst, "Partir de "+jeu.Villes[v].Nom, min(max(x-largeur("Partir de "+jeu.Villes[v].Nom, f)/2, 6), LargPlat-largeur("Partir de "+jeu.Villes[v].Nom, f)-6), y-32-30*cam.Z, f, coulParch, coulCuir, 2)
	}

	g.dessinerMini(dst)

	fondCuir(dst, PanneauX, 0, LargPanneau, HauteurEcran)
	frise(dst, PanneauX, 0, LargPanneau, 26)
	ligne(dst, PanneauX, 0, PanneauX, HauteurEcran, 3, coulOrSombre)
	ecrire(dst, "L'expédition", pnX+6, 34, face(28, policeTitre), coulOr)
	paragraphe(dst, g.consigneOrganisation(), pnX+6, 82, pnW-12, face(18, policeItalique), coulParch, 21)

	carte(dst, pnX, orgY, pnW, 400)
	titreCarte(dst, "Achats", pnX, orgY+4, pnW)
	fn := face(17, policeGras)
	fa := face(14, policeItalique)
	for i, a := range articles {
		y := orgY + 42 + float64(i)*56
		ecrire(dst, a.nom, pnX+18, y, fn, coulEncre)
		ecrire(dst, fmt.Sprintf("%d sols", a.prix), pnX+150, y+1, face(16, policeTexte), coulSepia)
		ecrire(dst, a.aide, pnX+18, y+21, fa, coulSepia)
		ecrireCentre(dst, fmt.Sprint(*g.quantite(i)), pnX+pnW-70, y+4, face(22, policeGras), coulEncre)
	}
	reste := jeu.Budget - g.achats.Cout()
	c := coulJungle
	if reste < 0 {
		c = coulSang
	}
	ecrire(dst, fmt.Sprintf("Reste : %d sols d'or sur 1 000", reste), pnX+18, orgY+376, face(18, policeGras), c)

	y := orgY + 412
	carte(dst, pnX, y, pnW, 226)
	titreCarte(dst, "Départ", pnX, y+4, pnW)
	ville := "à choisir sur la carte"
	if g.ville >= 0 {
		ville = jeu.Villes[g.ville].Nom
	}
	ecrire(dst, "Ville : "+ville, pnX+18, y+40, fn, coulEncre)
	ecrire(dst, "Mode :", pnX+18, y+76, fn, coulEncre)
	ecrire(dst, "Code secret :", pnX+18, y+124, fn, coulEncre)
	ecrire(dst, "estime la fiabilité de vos renseignements", pnX+18, y+146, fa, coulSepia)
	ecrire(dst, "Pion :", pnX+18, y+184, fn, coulEncre)
	paragraphe(dst, fmt.Sprintf("%d hommes. Les arquebusiers pèsent dans les combats (tableau 2) ; médecin, jésuite et guide évitent les retards et adoucissent les natifs ; les cadeaux achètent leur amitié. Une jangada file sur les fleuves ; abandonnée, elle est perdue.", g.hommes()),
		pnX+6, y+246, pnW-12, face(16, policeItalique), coulParch2, 19)
}

func (g *Game) boutonsOrganisation() {
	for i := range articles {
		q := g.quantite(i)
		y := orgY + 40 + float64(i)*56
		pas := 1
		g.bouton(pnX+pnW-122, y, 34, 32, "−", *q > 0, func() {
			*q -= pas
			if i == 5 && g.achats.Jangadas == 0 {
				g.achats.EnJangada = false
			}
		}).Taille = 26
		g.bouton(pnX+pnW-52, y, 34, 32, "+", g.achats.Cout()+articles[i].prix <= jeu.Budget, func() { *q += pas }).Taille = 26
	}
	y := orgY + 412
	peutJangada := g.achats.Jangadas > 0 && (g.ville < 0 || jeu.Navigable(jeu.Villes[g.ville].Case) || jeu.Cotiere(jeu.Villes[g.ville].Case))
	b := g.bouton(pnX+92, y+68, 120, 34, "À pied", true, func() { g.achats.EnJangada = false })
	b.Choisi = !g.achats.EnJangada
	b.Taille = 19
	b = g.bouton(pnX+222, y+68, 150, 34, "En jangada", peutJangada, func() { g.achats.EnJangada = true })
	b.Choisi = g.achats.EnJangada
	b.Taille = 19
	for c := 1; c <= 6; c++ {
		b := g.bouton(pnX+150+float64(c-1)*38, y+114, 32, 32, fmt.Sprint(c), true, func() { g.code = c })
		b.Choisi = g.code == c
		b.Taille = 20
	}
	for c := range 6 {
		b := g.bouton(pnX+92+float64(c)*44, y+176, 36, 36, "", true, func() { g.couleur = c })
		b.Choisi = g.couleur == c
		b.Style = BoutonDiscret
	}
	ok := g.ville >= 0 && g.achats.Cout() <= jeu.Budget && g.hommes() > 0
	b = g.bouton(pnX+pnW-200, 900, 200, 56, "En route !", ok, func() { g.commencer(*optGraine, 0) })
	b.Style = BoutonOr
	b.Taille = 24
	b = g.bouton(pnX, 900, 140, 56, "Retour", true, func() { g.ecran = EcranAccueil })
	b.Style = BoutonDiscret
	b.Taille = 22
	g.boutonsZoom()
}

// pions de couleur dessinés sur leurs boutons (après les boutons)
func (g *Game) dessinerPionsCouleur(dst *ebiten.Image) {
	y := orgY + 412
	for c := range 6 {
		x := pnX + 95 + float64(c)*44
		dessinerImage(dst, imgExpeditions[c], x, y+179, 30, 30, nil)
		if c == g.couleur {
			cadre(dst, x-5, y+174, 40, 40, 3, coulSang)
		}
	}
}
