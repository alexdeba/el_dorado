package main

import (
	"os"
	"testing"
	"time"

	"eldorado/jeu"
)

// partieSansFenetre : une partie complète pilotée comme par l'interface,
// l'ordinateur jouant pour l'humain jusqu'à la demande voulue.
func partieSansFenetre(graine uint64, nbAdv, ville int, jangada bool, arret TypeReq) *Partie {
	g := nouveauJeu()
	g.ville = ville
	g.code = 3
	if jangada {
		g.achats.Membres[jeu.Arquebusier] = 5
		g.achats.Jangadas = 1
		g.achats.EnJangada = true
		g.achats.Cadeaux = 7
	}
	p := nouvellePartie(g.achats, g.ville, g.code, 0, nbAdv, graine, 0)
	p.muet = true
	p.arret = arret
	return p
}

// atteindre : joue la partie jusqu'à la demande voulue (ou la fin).
func atteindre(p *Partie, t TypeReq) bool {
	p.lancer()
	limite := time.Now().Add(5 * time.Second)
	for time.Now().Before(limite) {
		p.mu.Lock()
		r, fin := p.Req.Type, p.Termine
		p.mu.Unlock()
		if r == t {
			return true
		}
		if fin {
			return false
		}
		time.Sleep(time.Millisecond)
	}
	return false
}

func TestDemandes(t *testing.T) {
	// chaque type de décision doit se présenter dans quelques parties
	for _, d := range []struct {
		t   TypeReq
		nom string
		adv int
	}{{ReqDeplacer, "deplacer", 3}, {ReqContact, "contact", 3}, {ReqVictime, "victime", 3}, {ReqAnnoncer, "annoncer", 3}, {ReqCombat, "combat", 5}} {
		trouve := uint64(0)
		for gr := uint64(1); gr <= 80 && trouve == 0; gr++ {
			if atteindre(partieSansFenetre(gr, d.adv, 1, false, d.t), d.t) {
				trouve = gr
			}
		}
		if trouve == 0 {
			t.Errorf("demande %s jamais atteinte", d.nom)
		} else if os.Getenv("GRAINES") != "" {
			t.Logf("-arret %s -adv %d : graine %d", d.nom, d.adv, trouve)
		}
	}
}

func TestReponses(t *testing.T) {
	// le joueur humain répond à une rencontre en fuyant, puis rien ne bloque
	p := partieSansFenetre(3, 2, 1, false, ReqContact)
	if !atteindre(p, ReqContact) {
		t.Skip("pas de rencontre")
	}
	p.mu.Lock()
	moi := p.E.Exp[p.Humain]
	avant := moi.Prec
	p.arret = ReqDeplacer
	p.repondre(cmdNon{})
	p.mu.Unlock()
	for range 2000 {
		p.mu.Lock()
		r := p.Req.Type
		pos := moi.Pos
		p.mu.Unlock()
		if r == ReqDeplacer {
			if pos != avant {
				t.Errorf("après la fuite : %v, attendu %v", pos, avant)
			}
			return
		}
		time.Sleep(time.Millisecond)
	}
}

func TestGeometrie(t *testing.T) {
	// le centre de chaque case retombe sur la case
	for _, k := range jeu.Cases() {
		x, y := centre(k)
		if caseSous(x, y) != k || caseSous(x+RayonHexX*0.7, y) != k || caseSous(x, y-RayonHexY*0.7) != k {
			t.Fatalf("case %v mal repérée", k)
		}
	}
}

func TestVictoires(t *testing.T) {
	if os.Getenv("GRAINES") == "" {
		t.Skip()
	}
	n, m := 0, 0
	for gr := uint64(1); gr <= 60; gr++ {
		g := nouveauJeu()
		p := nouvellePartie(g.achats, 1, 3, 0, 3, gr, jeu.ToursMax)
		p.muet = true
		for !p.E.Fini {
			p.E.JouerTour()
		}
		if p.E.Gagnant == 0 && n < 3 {
			t.Logf("victoire du joueur : graine %d, tour %d", gr, p.E.Tour)
			n++
		}
		if p.E.Gagnant > 0 && m < 3 {
			m++
			t.Logf("victoire de %s : graine %d, tour %d", p.E.Exp[p.E.Gagnant].Nom, gr, p.E.Tour)
		}
	}
}
