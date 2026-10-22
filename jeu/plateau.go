package jeu

// Le plateau : cases hexagonales (pointe en haut) de la carte de l'encart,
// rangées horizontales, les rangées impaires décalées d'une demi-case vers
// la droite. Terrain, fleuves, cascades et villes sont relevés sur la photo
// par outils/decoupe (plateau_donnees.go) ; la région centrale est décrite
// ici à la main.

const NbRang, NbCol = 25, 27

type Case struct{ C, R int }

type Terrain byte

const (
	Mer      Terrain = 'm'
	Marais   Terrain = 'a'
	Jungle   Terrain = 'j'
	Plaine   Terrain = 'p'
	Plateau  Terrain = 't'
	Montagne Terrain = 'n'
)

func (t Terrain) String() string {
	switch t {
	case Mer:
		return "mer"
	case Marais:
		return "marais"
	case Jungle:
		return "jungle"
	case Plaine:
		return "plaine"
	case Plateau:
		return "plateau"
	case Montagne:
		return "montagne"
	}
	return "?"
}

// Troncon : lien de fleuve entre deux cases voisines, de l'amont vers l'aval.
type Troncon struct {
	Amont, Aval Case
	Fleuve      int
}

type Ville struct {
	Nom  string
	Case Case
}

// lienFleuve : vu d'une case, un tronçon qui mène à vers.
type lienFleuve struct {
	vers   Case
	fleuve int
	aval   bool // vers est en aval
}

var (
	liensFleuve = map[Case][]lienFleuve{}
	cascades    = map[Case]bool{}
	Villes      = villesPlateau
)

func init() {
	for _, t := range tronconsPlateau {
		liensFleuve[t.Amont] = append(liensFleuve[t.Amont], lienFleuve{t.Aval, t.Fleuve, true})
		liensFleuve[t.Aval] = append(liensFleuve[t.Aval], lienFleuve{t.Amont, t.Fleuve, false})
	}
	for _, k := range cascadesPlateau {
		cascades[k] = true
	}
}

func (k Case) Valide() bool { return k.C >= 0 && k.R >= 0 && k.C < NbCol && k.R < NbRang }

func TerrainDe(k Case) Terrain {
	if !k.Valide() {
		return Mer
	}
	return Terrain(rangeesPlateau[k.R][k.C])
}

func Terre(k Case) bool { return TerrainDe(k) != Mer }

// Voisins : les six cases voisines (y compris hors plateau ou en mer).
func Voisins(k Case) [6]Case {
	if k.R%2 != 0 {
		return [6]Case{{k.C, k.R - 1}, {k.C + 1, k.R - 1}, {k.C - 1, k.R}, {k.C + 1, k.R}, {k.C, k.R + 1}, {k.C + 1, k.R + 1}}
	}
	return [6]Case{{k.C - 1, k.R - 1}, {k.C, k.R - 1}, {k.C - 1, k.R}, {k.C + 1, k.R}, {k.C - 1, k.R + 1}, {k.C, k.R + 1}}
}

func Voisines(a, b Case) bool {
	for _, v := range Voisins(a) {
		if v == b {
			return true
		}
	}
	return false
}

// cube : coordonnées cubiques (rangées impaires décalées à droite).
func cube(k Case) (int, int, int) {
	x := k.C - (k.R-(k.R&1))/2
	z := k.R
	return x, -x - z, z
}

// Distance : nombre de pas d'une case à l'autre.
func Distance(a, b Case) int {
	ax, ay, az := cube(a)
	bx, by, bz := cube(b)
	return max(abs(ax-bx), abs(ay-by), abs(az-bz))
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// Fleuve : la case porte un fleuve navigable.
func Fleuve(k Case) bool { return len(liensFleuve[k]) > 0 }

func Cascade(k Case) bool { return cascades[k] }

// Cotiere : case de terre qui touche la mer.
func Cotiere(k Case) bool {
	if !Terre(k) {
		return false
	}
	for _, v := range Voisins(k) {
		if v.Valide() && !Terre(v) {
			return true
		}
	}
	return false
}

// VilleEn : numéro de la ville de la case, -1 sinon.
func VilleEn(k Case) int {
	for i, v := range Villes {
		if v.Case == k {
			return i
		}
	}
	return -1
}

// Cases : toutes les cases de terre.
func Cases() []Case {
	var l []Case
	for r := range NbRang {
		for c := range NbCol {
			if k := (Case{c, r}); Terre(k) {
				l = append(l, k)
			}
		}
	}
	return l
}

// ---- région centrale ----

// Centre : la « Montagne sacrée », au cœur des 19 cases de la région centrale.
var Centre = Case{12, 12}

type Zone int

const (
	SansZone Zone = iota
	Ouest
	Est
	Sud
)

func (z Zone) String() string {
	return [...]string{"?", "Ouest", "Est", "Sud"}[z]
}

// Nature du terrain d'une case de la région centrale (tableau 5).
type Nature int

const (
	NatInconnue Nature = iota
	NatJungle
	NatMarais
	NatSec // plaine ou plateau
)

func (n Nature) String() string {
	return [...]string{"?", "jungle", "marais", "plaine ou plateau"}[n]
}

type CaseCentrale struct {
	Zone   Zone
	Fleuve bool
	Nature Nature
}

// regionCentrale : les 18 cases où peut se trouver El Dorado. Chaque zone en
// compte six, une par combinaison (fleuve oui/non) × (jungle, marais, sec),
// comme l'annonce la règle. Sur la carte, la zone Est n'a pas de marais bien
// net : ses deux cases d'un turquoise plus bleu (12,11 et 13,12) en tiennent
// lieu ; 11,14, longée par le fleuve, compte comme case à fleuve.
var regionCentrale = map[Case]CaseCentrale{
	{11, 10}: {Ouest, false, NatJungle}, {11, 11}: {Ouest, true, NatJungle},
	{11, 12}: {Ouest, false, NatSec}, {10, 11}: {Ouest, true, NatSec},
	{10, 13}: {Ouest, false, NatMarais}, {10, 12}: {Ouest, true, NatMarais},

	{12, 10}: {Est, false, NatJungle}, {13, 10}: {Est, true, NatJungle},
	{14, 12}: {Est, false, NatSec}, {13, 11}: {Est, true, NatSec},
	{13, 12}: {Est, false, NatMarais}, {12, 11}: {Est, true, NatMarais},

	{11, 13}: {Sud, false, NatJungle}, {12, 13}: {Sud, true, NatJungle},
	{12, 14}: {Sud, false, NatSec}, {11, 14}: {Sud, true, NatSec},
	{13, 14}: {Sud, false, NatMarais}, {13, 13}: {Sud, true, NatMarais},
}

// DansRegion : la case fait partie des 19 cases de la région centrale.
func DansRegion(k Case) bool { return Distance(k, Centre) <= 2 }

// Centrale : description d'une des 18 cases candidates.
func Centrale(k Case) (CaseCentrale, bool) {
	c, ok := regionCentrale[k]
	return c, ok
}

// CaseDorado : la seule case qui réunit les trois renseignements.
func CaseDorado(z Zone, fleuve bool, n Nature) Case {
	for k, c := range regionCentrale {
		if c.Zone == z && c.Fleuve == fleuve && c.Nature == n {
			return k
		}
	}
	return Case{-1, -1}
}
