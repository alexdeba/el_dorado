package main

import (
	"sync"
	"time"

	"eldorado/jeu"
)

// Le moteur joue la partie dans sa propre goroutine, en tenant le verrou de
// la Partie ; il le relâche pendant les pauses d'animation et quand il attend
// une décision du joueur humain. L'interface (Update/Draw) prend ce verrou
// pour lire l'état et poster les réponses.

type TypeReq int

const (
	ReqAucune   TypeReq = iota // l'ordinateur joue
	ReqDeplacer                // choisir sa destination
	ReqAnnoncer                // sur une case candidate : annoncer El Dorado ?
	ReqContact                 // rencontre : contact ou fuite, cadeaux
	ReqVictime                 // qui meurt
	ReqCombat                  // combat entre expéditions : poursuivre ou fuir
)

type Requete struct {
	Type   TypeReq
	Coups  []jeu.Coup
	Natifs int
	Adv    int
	Num    int // numéro de demande (pour réinitialiser la saisie)
}

// commandes de l'interface
type (
	cmdCoup    struct{ c jeu.Coup }
	cmdOui     struct{}
	cmdNon     struct{}
	cmdContact struct{ cadeaux int }
	cmdMembre  struct{ m jeu.Membre }
)

type Partie struct {
	mu      sync.Mutex
	E       *jeu.Etat
	Humain  int
	Req     Requete
	rep     chan any
	Rapide  bool
	Termine bool
	auto    int     // tours joués par l'ordinateur pour le joueur humain (captures)
	arret   TypeReq // captures : l'ordinateur joue pour l'humain jusqu'à cette demande
	muet    bool    // pas de pause (tests)
	nreq    int
}

var nomsConquistadors = []string{"Orellana", "Quesada", "Belalcázar", "Federmann", "Aguirre", "Hutten"}

// nouvellePartie : l'expédition du joueur est organisée ; l'ordinateur
// organise les autres sur les villes restantes.
func nouvellePartie(a jeu.Achats, ville, code, couleur, nbAdv int, graine uint64, toursAuto int) *Partie {
	p := &Partie{Humain: 0, rep: make(chan any, 1), auto: toursAuto}
	moi := jeu.NouvelleExpedition("Votre expédition", couleur, a, ville, code)
	exps := []*jeu.Expedition{moi}
	ctrls := []jeu.Controleur{&Humain{p: p, ia: jeu.NouvelleIA(0.5)}}
	e := jeu.Nouveau(nil, nil, graine)
	var libres []int
	for v := range jeu.Villes {
		if v != ville {
			libres = append(libres, v)
		}
	}
	coul := 0
	for i := range nbAdv {
		au := 0.2 + 0.6*float64(i)/float64(max(nbAdv-1, 1))
		ach, v, c := jeu.PreparerIA(e.Rng(), libres, au)
		for k, l := range libres {
			if l == v {
				libres = append(libres[:k], libres[k+1:]...)
				break
			}
		}
		if coul == couleur {
			coul++
		}
		exps = append(exps, jeu.NouvelleExpedition(nomsConquistadors[i], coul, ach, v, c))
		ctrls = append(ctrls, jeu.NouvelleIA(au))
		coul++
	}
	e.Exp, e.Ctrl = exps, ctrls
	p.E = e
	e.Humain = p.Humain
	e.Pause = p.pause
	return p
}

// lancer : la partie se joue en arrière-plan.
func (p *Partie) lancer() {
	go func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		for !p.E.Fini {
			p.E.JouerTour()
		}
		p.Termine = true
	}()
}

// jouerAuto : premiers tours joués d'un coup (options de capture).
func (p *Partie) jouerAuto() {
	for p.E.Tour < p.auto && !p.E.Fini {
		p.E.JouerTour()
	}
}

func (p *Partie) pause() {
	if p.E.Tour <= p.auto || p.muet || p.arret != ReqAucune {
		return
	}
	d := 420 * time.Millisecond
	if p.E.Courant == p.Humain {
		d = 650 * time.Millisecond
	}
	if p.Rapide {
		d /= 6
	}
	p.mu.Unlock()
	time.Sleep(d)
	p.mu.Lock()
}

// attendre : publie une demande et attend la réponse de l'interface.
func (p *Partie) attendre(r Requete) any {
	p.nreq++
	r.Num = p.nreq
	p.Req = r
	p.mu.Unlock()
	v := <-p.rep
	p.mu.Lock()
	p.Req = Requete{}
	return v
}

// repondre : appelé par l'interface, verrou tenu.
func (p *Partie) repondre(v any) {
	if p.Req.Type == ReqAucune {
		return
	}
	p.Req = Requete{}
	select {
	case p.rep <- v:
	default:
	}
}

// Humain : contrôleur du joueur humain ; délègue à l'IA pendant les tours
// automatiques et, pour les captures, jusqu'à la demande voulue.
type Humain struct {
	p  *Partie
	ia *jeu.IA
}

func (h *Humain) auto(e *jeu.Etat, t TypeReq) bool {
	if e.Tour <= h.p.auto {
		return true
	}
	if h.p.arret == ReqAucune {
		return false
	}
	if h.p.arret == t {
		h.p.arret = ReqAucune
		return false
	}
	return true
}

func (h *Humain) Deplacer(e *jeu.Etat, j int, coups []jeu.Coup) jeu.Coup {
	if h.auto(e, ReqDeplacer) {
		return h.ia.Deplacer(e, j, coups)
	}
	for {
		if c, ok := h.p.attendre(Requete{Type: ReqDeplacer, Coups: coups}).(cmdCoup); ok {
			return c.c
		}
	}
}

func (h *Humain) Annoncer(e *jeu.Etat, j int) bool {
	if h.auto(e, ReqAnnoncer) {
		return h.ia.Annoncer(e, j)
	}
	_, ok := h.p.attendre(Requete{Type: ReqAnnoncer}).(cmdOui)
	return ok
}

func (h *Humain) Contact(e *jeu.Etat, j, natifs int) (bool, int) {
	if h.auto(e, ReqContact) {
		return h.ia.Contact(e, j, natifs)
	}
	for {
		switch c := h.p.attendre(Requete{Type: ReqContact, Natifs: natifs}).(type) {
		case cmdContact:
			return true, c.cadeaux
		case cmdNon:
			return false, 0
		}
	}
}

func (h *Humain) Victime(e *jeu.Etat, j int) jeu.Membre {
	if h.auto(e, ReqVictime) {
		return h.ia.Victime(e, j)
	}
	x := e.Exp[j]
	// un seul choix possible : inutile de demander
	n, m := 0, jeu.Arquebusier
	for k := range jeu.NbMembres {
		if x.Membres[k] > 0 {
			n++
			m = k
		}
	}
	if n <= 1 {
		return m
	}
	for {
		if c, ok := h.p.attendre(Requete{Type: ReqVictime}).(cmdMembre); ok && x.Membres[c.m] > 0 {
			return c.m
		}
	}
}

func (h *Humain) Combattre(e *jeu.Etat, j, adv int) bool {
	if h.auto(e, ReqCombat) {
		return h.ia.Combattre(e, j, adv)
	}
	_, ok := h.p.attendre(Requete{Type: ReqCombat, Adv: adv}).(cmdOui)
	return ok
}
