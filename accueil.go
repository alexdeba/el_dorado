package main

import (
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

// L'accueil : la couverture, et le choix du nombre d'adversaires.

func (g *Game) dessinerAccueil(dst *ebiten.Image) {
	b := imgCouv.Bounds()
	h := float64(HauteurEcran)
	w := h * float64(b.Dx()) / float64(b.Dy())
	dessinerImage(dst, imgCouv, (LargeurEcran-w)/2, 0, w, h, nil)
	// bandeau du bas
	y0 := 790.0
	for i := range 40 {
		a := uint8(min(255, i*9))
		rect(dst, 0, y0-80+float64(i)*2, LargeurEcran, 2, avecAlpha(coulCuir, a))
	}
	rect(dst, 0, y0, LargeurEcran, HauteurEcran-y0, coulCuir)
	frise(dst, 0, y0, LargeurEcran, 26)
	ecrireCentre(dst, "Vers 1550, une légende parcourt les comptoirs du Nouveau Monde : au cœur du continent, un cacique couvert de poudre d'or…",
		LargeurEcran/2, y0+36, face(21, policeItalique), coulParch2)
	ecrire(dst, "Expéditions rivales :", 330, y0+92, face(26, policeGothique), coulOr)
	ecrireCentre(dst, "D'après « El Dorado », jeu en encart d'un magazine, signé Strate et Pop",
		LargeurEcran/2, HauteurEcran-30, face(15, policeTexte), coulParch3)
}

func (g *Game) boutonsAccueil() {
	y := 790.0 + 84
	for n := 1; n <= 5; n++ {
		b := g.bouton(560+float64(n-1)*58, y, 50, 46, fmt.Sprint(n), true, func() { g.nbAdv = n })
		b.Choisi = g.nbAdv == n
		b.Taille = 26
	}
	b := g.bouton(890, y-4, 300, 54, "Organiser l'expédition", true, func() { g.ecran = EcranOrganisation })
	b.Style = BoutonOr
	b.Taille = 25
	g.bouton(1210, y, 130, 46, "Règle", true, func() { g.regles = true }).Taille = 23
	b = g.bouton(1350, y, 120, 46, "Quitter", true, func() { os.Exit(0) })
	b.Style = BoutonDanger
	b.Taille = 23
}
