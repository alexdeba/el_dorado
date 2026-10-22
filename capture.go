package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"

	"eldorado/jeu"
)

// Options de test : capture d'écran et mise en place directe d'une partie.
var (
	optCapture = flag.String("shot", "", "enregistre l'écran dans ce fichier PNG puis quitte")
	optImages  = flag.Int("frames", 40, "nombre d'images avant la capture")
	optEcran   = flag.String("ecran", "", "organisation ou jeu : saute l'accueil")
	optTours   = flag.Int("tours", 0, "tours joués automatiquement avant de rendre la main")
	optGraine  = flag.Uint64("graine", 0, "graine du hasard (0 = aléatoire)")
	optAdv     = flag.Int("adv", 3, "nombre d'adversaires (1 à 5)")
	optVille   = flag.Int("ville", 1, "ville de départ (0 Quito … 5 Buenos Aires)")
	optJangada = flag.Bool("jangada", false, "partir en jangada")
	optRegle   = flag.Int("regle", 0, "afficher cette page de la règle (1 à 8)")
	optFeuille = flag.Bool("feuille", false, "afficher la feuille des renseignements")
	optRapide  = flag.Bool("rapide", false, "ordinateur accéléré")
	optSurvol  = flag.String("survol", "", "captures : souris sur cette case (c,r)")
	optZoom    = flag.Float64("zoom", 1, "captures : zoom du plateau (1 à 4), centré sur -survol ou sur votre pion")
	optArret   = flag.String("arret", "", "captures : l'ordinateur joue pour vous jusqu'à cette demande (deplacer, annoncer, contact, victime, combat)")
)

var arretCapture TypeReq

func (g *Game) preparer() {
	g.nbAdv = min(max(*optAdv, 1), 5)
	if *optRegle > 0 {
		g.regles, g.pageRgl = true, *optRegle-1
	}
	switch *optEcran {
	case "organisation":
		g.ecran = EcranOrganisation
		if *optVille >= 0 && *optVille < len(jeu.Villes) && *optTours > 0 {
			g.ville = *optVille
		}
	case "jeu":
		arretCapture = map[string]TypeReq{"deplacer": ReqDeplacer, "annoncer": ReqAnnoncer, "contact": ReqContact,
			"victime": ReqVictime, "combat": ReqCombat}[*optArret]
		g.ville = min(max(*optVille, 0), len(jeu.Villes)-1)
		g.code = 3
		if *optJangada {
			g.achats.Membres[jeu.Arquebusier] = 5
			g.achats.Jangadas = 1
			g.achats.EnJangada = true
			g.achats.Cadeaux = 7
		}
		g.commencer(*optGraine, *optTours)
		g.p.Rapide = *optRapide
		g.feuille = *optFeuille
	}
	if *optZoom > 1 {
		cam.Z = min(*optZoom, ZoomMax)
		k := jeu.Centre
		if g.p != nil {
			g.p.mu.Lock()
			k = g.p.E.Exp[g.p.Humain].Pos
			g.p.mu.Unlock()
		}
		fmt.Sscanf(*optSurvol, "%d,%d", &k.C, &k.R)
		cam.centrer(centreMonde(k))
	}
}

func (g *Game) capturer(dst *ebiten.Image) {
	if *optCapture == "" {
		return
	}
	g.images++
	if g.images < *optImages {
		return
	}
	// on attend que la demande voulue soit arrivée
	if g.p != nil && arretCapture != ReqAucune && g.p.Req.Type != arretCapture && g.images < *optImages+2000 {
		return
	}
	b := dst.Bounds()
	img := image.NewRGBA(b)
	dst.ReadPixels(img.Pix)
	f, err := os.Create(*optCapture)
	if err != nil {
		log.Fatal(err)
	}
	png.Encode(f, img)
	f.Close()
	os.Exit(0)
}

// survolCapture : option -survol, la souris est posée sur une case.
func (g *Game) survolCapture() {
	if *optSurvol == "" {
		return
	}
	var k jeu.Case
	fmt.Sscanf(*optSurvol, "%d,%d", &k.C, &k.R)
	if k.Valide() {
		g.survol = k
		x, y := centre(k)
		g.mx, g.my = int(x), int(y)
	}
}
