// El Dorado — d'après le jeu paru en encart dans un magazine, signé Strate
// et Pop (assets/Regles1.jpg, Regles2.jpg, Encart.jpg). Un joueur humain contre une à cinq expéditions menées par l'ordinateur, à la
// recherche de la cité d'or au cœur du Mato Grosso.
package main

import (
	"flag"
	"log"
	"math"
	"math/rand/v2"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"eldorado/jeu"
)

const (
	LargeurEcran = 1600
	HauteurEcran = 1000
	PanneauX     = 1180.0
	LargPanneau  = LargeurEcran - PanneauX
)

type Ecran int

const (
	EcranAccueil Ecran = iota
	EcranOrganisation
	EcranJeu
)

type Game struct {
	ecran Ecran
	nbAdv int
	p     *Partie

	// organisation de l'expédition
	achats  jeu.Achats
	ville   int
	code    int
	couleur int

	// saisie
	survol    jeu.Case
	cadeaux   int // cadeaux proposés lors d'une rencontre
	numReq    int
	regles    bool
	feuille   bool
	pageRgl   int
	lignesRgl []string
	defilLog  int

	boutons []*Bouton
	mx, my  int
	images  int
	t0      time.Time

	// pions affichés : glissement vers leur case
	affiche map[int][2]float64
	dernier time.Time

	// caméra : glisser du bouton droit, glisser sur la mini-carte
	glisse, miniTire bool
	gx, gy           float64
}

func nouveauJeu() *Game {
	g := &Game{nbAdv: 3, t0: time.Now(), ville: -1, code: rand.IntN(6) + 1, affiche: map[int][2]float64{}}
	g.achats = jeu.Achats{Membres: [jeu.NbMembres]int{jeu.Arquebusier: 7, jeu.Medecin: 1, jeu.Jesuite: 1, jeu.Guide: 1}, Cadeaux: 10}
	return g
}

func (g *Game) Layout(int, int) (int, int) { return LargeurEcran, HauteurEcran }

func (g *Game) temps() float64 { return time.Since(g.t0).Seconds() }

// commencer : l'expédition est prête, la partie démarre.
func (g *Game) commencer(graine uint64, toursAuto int) {
	if graine == 0 {
		graine = rand.Uint64()
	}
	g.p = nouvellePartie(g.achats, g.ville, g.code, g.couleur, g.nbAdv, graine, toursAuto)
	g.p.arret = arretCapture
	g.p.jouerAuto()
	g.p.lancer()
	g.ecran = EcranJeu
	g.affiche = map[int][2]float64{}
}

// positionAffichee : les pions glissent vers leur case ; deux pions sur la
// même case s'écartent. Coordonnées du plateau, hors caméra.
func (g *Game) positionAffichee(i int) (float64, float64) {
	e := g.p.E
	x := e.Exp[i]
	cx, cy := centreMonde(x.Pos)
	n, rang := 0, 0
	for j, y := range e.Exp {
		if y.Vivante && y.Pos == x.Pos {
			if j < i {
				rang++
			}
			n++
		}
	}
	if n > 1 {
		a := float64(rang)/float64(n)*2*math.Pi - math.Pi/2
		cx += 14 * math.Cos(a)
		cy += 10 * math.Sin(a)
	}
	return cx, cy
}

func (g *Game) glisser() {
	dt := time.Since(g.dernier).Seconds()
	g.dernier = time.Now()
	if dt > 0.2 {
		dt = 0.2
	}
	for i := range g.p.E.Exp {
		cx, cy := g.positionAffichee(i)
		a, ok := g.affiche[i]
		if !ok || math.Hypot(a[0]-cx, a[1]-cy) > 400 {
			g.affiche[i] = [2]float64{cx, cy}
			continue
		}
		k := 1 - math.Exp(-dt*9)
		g.affiche[i] = [2]float64{a[0] + (cx-a[0])*k, a[1] + (cy-a[1])*k}
	}
}

func (g *Game) Update() error {
	g.mx, g.my = ebiten.CursorPosition()
	clic := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	if inpututil.IsKeyJustPressed(ebiten.KeyF11) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}
	if g.p != nil {
		g.p.mu.Lock()
		defer g.p.mu.Unlock()
		if g.p.Req.Num != g.numReq {
			g.numReq = g.p.Req.Num
			g.nouvelleDemande()
		}
		g.glisser()
	}
	g.construireBoutons()
	if clic {
		for _, b := range g.boutons {
			if b.Actif && b.Contient(g.mx, g.my) {
				b.Action()
				return nil
			}
		}
	}
	if g.regles {
		g.touchesRegle()
		return nil
	}
	surMini := g.cameraActive() && g.gererCamera()
	g.survol = caseSous(float64(g.mx), float64(g.my))
	if surMini {
		g.survol = jeu.Case{C: -1, R: -1}
	}
	g.survolCapture()
	switch g.ecran {
	case EcranAccueil:
		return nil
	case EcranOrganisation:
		if clic && g.survol.Valide() {
			if v := jeu.VilleEn(g.survol); v >= 0 {
				g.choisirVille(v)
			}
		}
		return nil
	}
	if g.feuille {
		if inpututil.IsKeyJustPressed(ebiten.KeyF) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			g.feuille = false
		}
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.regles = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		g.feuille = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyA) {
		g.p.Rapide = !g.p.Rapide
	}
	if _, dy := ebiten.Wheel(); dy != 0 && float64(g.mx) > PanneauX {
		g.defilLog = max(g.defilLog+int(dy), 0)
	}
	if clic && g.survol.Valide() && g.p.Req.Type == ReqDeplacer {
		moi := g.p.E.Exp[g.p.Humain]
		if g.survol == moi.Pos {
			g.p.repondre(cmdCoup{jeu.Coup{Vers: moi.Pos, Mode: jeu.Rester}})
		} else if c, ok := coupsParCase(g.p.Req.Coups)[g.survol]; ok {
			g.p.repondre(cmdCoup{c})
		}
	}
	return nil
}

// nouvelleDemande : réglages par défaut d'une nouvelle décision.
func (g *Game) nouvelleDemande() {
	e := g.p.E
	moi := e.Exp[g.p.Humain]
	if g.p.Req.Type == ReqContact {
		g.cadeaux = 0
		if moi.Cadeaux > 0 {
			g.cadeaux = 1
		}
	}
	g.defilLog = 0
}

func (g *Game) Draw(dst *ebiten.Image) {
	if g.p != nil {
		g.p.mu.Lock()
		defer g.p.mu.Unlock()
	}
	dst.Fill(coulCuir)
	switch g.ecran {
	case EcranAccueil:
		g.dessinerAccueil(dst)
	case EcranOrganisation:
		g.dessinerOrganisation(dst)
	case EcranJeu:
		g.dessinerPlateau(dst)
		g.dessinerPanneau(dst)
		if g.feuille {
			g.dessinerFeuille(dst)
		} else if g.p.Termine || g.p.E.Fini {
			g.dessinerFin(dst)
		}
	}
	if g.regles {
		g.dessinerRegles(dst)
	}
	for _, b := range g.boutons {
		dessinerBouton(dst, b, b.Actif && b.Contient(g.mx, g.my))
	}
	if g.ecran == EcranOrganisation && !g.regles {
		g.dessinerPionsCouleur(dst)
	}
	g.capturer(dst)
}

func (g *Game) bouton(x, y, w, h float64, t string, actif bool, f func()) *Bouton {
	b := &Bouton{X: x, Y: y, W: w, H: h, Texte: t, Actif: actif, Action: f}
	g.boutons = append(g.boutons, b)
	return b
}

func (g *Game) construireBoutons() {
	g.boutons = g.boutons[:0]
	if g.regles {
		g.boutonsRegles()
		return
	}
	switch g.ecran {
	case EcranAccueil:
		g.boutonsAccueil()
	case EcranOrganisation:
		g.boutonsOrganisation()
	case EcranJeu:
		if g.feuille {
			b := g.bouton(PanneauX/2-90, 930, 180, 44, "Fermer", true, func() { g.feuille = false })
			b.Style = BoutonOr
			return
		}
		if g.p.Termine || g.p.E.Fini {
			b := g.bouton(LargPlat/2-160, 600, 320, 56, "Nouvelle expédition", true, func() {
				n := g.nbAdv
				*g = *nouveauJeu()
				g.nbAdv = n
				g.ecran = EcranOrganisation
			})
			b.Style = BoutonOr
			b.Taille = 24
			return
		}
		g.boutonsJeu()
	}
}

var optPleinEcran = flag.Bool("fullscreen", false, "démarre en plein écran (F11 bascule)")

func main() {
	flag.Parse()
	chargerRessources()
	g := nouveauJeu()
	g.preparer()
	ebiten.SetWindowSize(LargeurEcran, HauteurEcran)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("El Dorado")
	ebiten.SetFullscreen(*optPleinEcran)
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
