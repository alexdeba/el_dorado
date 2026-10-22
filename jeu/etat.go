package jeu

import (
	"fmt"
	"math/rand/v2"
)

// Membre de l'expédition.
type Membre int

const (
	Arquebusier Membre = iota
	Medecin
	Jesuite
	Guide
	NbMembres
)

func (m Membre) String() string {
	return [...]string{"arquebusier", "médecin", "jésuite", "guide"}[m]
}

func (m Membre) Pluriel() string {
	return [...]string{"arquebusiers", "médecins", "jésuites", "guides"}[m]
}

// Prix en sols d'or (organisation de l'expédition).
const (
	Budget            = 1000
	PrixMembre        = 80
	PrixCadeau        = 20
	PrixJangada       = 160
	ToursConstruction = 4
)

// Achats : composition de l'expédition au départ.
type Achats struct {
	Membres   [NbMembres]int
	Cadeaux   int
	Jangadas  int  // jangadas ou matériel pour en construire
	EnJangada bool // partir en jangada (une des jangadas est construite au départ)
}

func (a Achats) Cout() int {
	n := 0
	for _, v := range a.Membres {
		n += v * PrixMembre
	}
	return n + a.Cadeaux*PrixCadeau + a.Jangadas*PrixJangada
}

type Expedition struct {
	Nom     string
	Couleur int // pion (0..5)
	Ville   int // ville de départ

	Membres      [NbMembres]int
	Cadeaux      int
	Jangada      bool // l'expédition voyage en jangada
	Kits         int  // matériel pour construire des jangadas
	Construction int  // tours passés à construire
	Code         int  // code secret, 1 à 6

	Pos, Prec  Case
	Retard     int // tours à passer
	Vivante    bool
	Tresor     bool
	Decouragee bool // a annoncé El Dorado à tort
}

func (x *Expedition) Hommes() int {
	n := 0
	for _, v := range x.Membres {
		n += v
	}
	return n
}

// Specialiste : médecin, jésuite ou guide présent.
func (x *Expedition) Specialiste() bool {
	return x.Membres[Medecin]+x.Membres[Jesuite]+x.Membres[Guide] > 0
}

// Marqueur : pion-natif posé sur une case.
type Marqueur struct {
	Numero     int
	Natifs     int
	Exterminee bool   // plus de natifs sur cette case
	Amis       []bool // tribu amicale envers ce joueur (pour toute la partie)
	Renseigne  []bool // la tribu a déjà renseigné ce joueur
	Absents    int    // natifs enfuis ou relâchés : absents tant que ce joueur reste (-1 : présents)
	Pose       int    // joueur qui l'a posé
}

type TypeRens int

const (
	RensZone TypeRens = iota
	RensFleuve
	RensNature
)

// Renseignement : une ligne de la feuille commune.
type Renseignement struct {
	Joueur   int
	Tour     int
	Type     TypeRens
	Zone     Zone
	Fleuve   bool
	Nature   Nature
	Distance int
	De       int // dé non modifié par le code secret
	Fiab     int // fiabilité réelle (secrète tant que les codes ne sont pas révélés)
}

func (r Renseignement) Texte() string {
	switch r.Type {
	case RensFleuve:
		if r.Fleuve {
			return "il y a un fleuve"
		}
		return "pas de fleuve"
	case RensNature:
		return r.Nature.String()
	}
	return "zone " + r.Zone.String()
}

// Entree : une ligne du journal de la partie.
type Entree struct {
	Tour   int
	Joueur int // -1 : général
	Texte  string
}

// Des : le dernier jet, pour l'affichage.
type Des struct {
	Joueur  int
	Valeurs []int
	Sens    string // « événement », « attitude »...
	Num     int    // compteur, pour repérer un nouveau jet
}

type Etat struct {
	Exp  []*Expedition
	Ctrl []Controleur

	Tour    int
	Courant int
	Humain  int // pour les messages (-1 : aucun)

	Marqueurs      map[Case]*Marqueur
	ProchainNumero int
	Feuille        []Renseignement

	// vérité, établie au fil des renseignements
	ZoneVraie      Zone
	FleuveCentral  int // -1 inconnu, 0 non, 1 oui
	NatureCentrale Nature
	sommes         map[[2]int]int // (joueur, zone) -> fiabilités cumulées

	DoradoTrouve bool
	TresorAuSol  bool
	CaseTresor   Case
	CodesReveles bool

	Fini    bool
	Gagnant int // -1 : personne

	Journal []Entree
	Des     Des

	// Pause : laisse l'interface montrer ce qui vient de se passer.
	Pause func()

	rng *rand.Rand
}

// Nouveau : les expéditions sont déjà organisées et placées sur leur ville.
func Nouveau(exps []*Expedition, ctrls []Controleur, graine uint64) *Etat {
	e := &Etat{
		Exp: exps, Ctrl: ctrls, Humain: -1,
		Marqueurs: map[Case]*Marqueur{}, ProchainNumero: 1,
		FleuveCentral: -1, sommes: map[[2]int]int{}, Gagnant: -1,
		rng: rand.New(rand.NewPCG(graine, graine^0x5eed)),
	}
	return e
}

// NouvelleExpedition : expédition organisée selon les achats, posée sur sa ville.
func NouvelleExpedition(nom string, couleur int, a Achats, ville, code int) *Expedition {
	x := &Expedition{Nom: nom, Couleur: couleur, Ville: ville, Membres: a.Membres,
		Cadeaux: a.Cadeaux, Kits: a.Jangadas, Code: code, Vivante: true}
	if a.EnJangada && a.Jangadas > 0 {
		x.Jangada = true
		x.Kits--
	}
	x.Pos = Villes[ville].Case
	x.Prec = x.Pos
	return x
}

func (e *Etat) De() int { return e.rng.IntN(6) + 1 }

func (e *Etat) Rng() *rand.Rand { return e.rng }

func (e *Etat) noter(j int, format string, args ...any) {
	e.Journal = append(e.Journal, Entree{e.Tour, j, fmt.Sprintf(format, args...)})
}

func (e *Etat) pause() {
	if e.Pause != nil {
		e.Pause()
	}
}

func (e *Etat) montrerDes(j int, sens string, v ...int) {
	e.Des = Des{Joueur: j, Valeurs: v, Sens: sens, Num: e.Des.Num + 1}
}

// Occupant : expédition vivante sur la case (autre que sauf), -1 sinon.
func (e *Etat) Occupant(k Case, sauf int) int {
	for i, x := range e.Exp {
		if i != sauf && x.Vivante && x.Pos == k {
			return i
		}
	}
	return -1
}

// Vivantes : nombre d'expéditions encore en vie.
func (e *Etat) Vivantes() int {
	n := 0
	for _, x := range e.Exp {
		if x.Vivante {
			n++
		}
	}
	return n
}

// Dorado : case d'El Dorado si les trois renseignements sont établis.
func (e *Etat) Dorado() (Case, bool) {
	if e.ZoneVraie == SansZone || e.FleuveCentral < 0 || e.NatureCentrale == NatInconnue {
		return Case{}, false
	}
	return CaseDorado(e.ZoneVraie, e.FleuveCentral == 1, e.NatureCentrale), true
}

// FiabiliteConnue : fiabilité d'un renseignement telle que la connaît le joueur j.
func (e *Etat) FiabiliteConnue(j int, r Renseignement) (int, bool) {
	if r.Type != RensZone || r.Joueur == j || e.CodesReveles {
		return r.Fiab, true
	}
	return 0, false
}
