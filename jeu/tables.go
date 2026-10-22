package jeu

// Les six tableaux imprimés sur la carte.

// Effet d'un événement du tableau 1.
type Effet int

const (
	EffMort          Effet = iota // 12 à 36 : un mort
	EffRetardMedecin              // 42 à 46 : un retard, sauf si médecin
	EffRetardGuide                // 52 à 56 : sauf si guide
	EffRetardJesuite              // 62 à 66 : sauf si jésuite
)

var nomsEvenements = map[int]string{
	12: "fièvre de l'or, rixe", 14: "plante carnivore", 16: "piranhas, morsure fatale",
	22: "fosse avec bambous empoisonnés", 24: "boa constrictor", 26: "ocelot",
	32: "champignons vénéneux", 34: "couguar", 36: "mygale",
	42: "épidémie de malaria", 44: "fruits empoisonnés, diarrhée", 46: "eau croupie, fièvre",
	52: "boussole déréglée", 54: "parcours inextricable", 56: "mirages, vapeurs fétides",
	62: "superstition", 64: "découragement", 66: "angoisse",
}

// Evenement : tableau 1, pour un nombre pair de 12 à 66.
func Evenement(n int) (string, Effet) {
	e := EffMort
	switch {
	case n >= 62:
		e = EffRetardJesuite
	case n >= 52:
		e = EffRetardGuide
	case n >= 42:
		e = EffRetardMedecin
	}
	return nomsEvenements[n], e
}

// Tableau 2 : rapport des forces. Colonnes : nombre d'arquebusiers ;
// rangées : nombre de natifs (10 à 60).
var (
	colonnesArquebusiers = [13]int{1, 3, 5, 7, 9, 11, 13, 15, 17, 19, 21, 23, 25}
	rapportForces        = [6][13]int{
		{-1, 0, 0, 1, 2, 3, 4, 5, 5, 5, 5, 5, 5},
		{-2, -1, 0, 0, 1, 2, 3, 4, 5, 5, 5, 5, 5},
		{-3, -2, -1, 0, 0, 1, 2, 3, 4, 5, 5, 5, 5},
		{-4, -3, -2, -1, 0, 0, 1, 2, 3, 4, 5, 5, 5},
		{-5, -4, -3, -2, -1, 0, 0, 1, 2, 3, 4, 5, 5},
		{-5, -5, -4, -3, -2, -1, 0, 0, 1, 2, 3, 4, 5},
	}
)

// RapportForces : modificateur du dé de combat. On lit la colonne du plus
// grand nombre d'arquebusiers imprimé qui ne dépasse pas le leur, et la
// rangée des dizaines de natifs ; sans arquebusier, on retire encore un point
// à la première colonne.
func RapportForces(arquebusiers, natifs int) int {
	r := min(max(natifs/10, 1), 6) - 1
	if arquebusiers <= 0 {
		return rapportForces[r][0] - 1
	}
	c := 0
	for i, n := range colonnesArquebusiers {
		if arquebusiers >= n {
			c = i
		}
	}
	return rapportForces[r][c]
}

type Attitude int

const (
	Hostile Attitude = iota
	Neutre
	Amicale
)

func (a Attitude) String() string { return [...]string{"hostile", "neutre", "amicale"}[a] }

// AttitudeNatifs : tableau 3 (dé modifié, borné à 1..6).
func AttitudeNatifs(d int) Attitude {
	switch {
	case d <= 2:
		return Hostile
	case d <= 4:
		return Neutre
	}
	return Amicale
}

type IssueCombat int

const (
	TroisMorts IssueCombat = iota
	DeuxMorts
	UnMort
	FuiteNatifs
	Prisonniers
)

// ResultatCombat : tableau 4 (dé modifié, borné à 1..6).
func ResultatCombat(d int) IssueCombat {
	switch {
	case d <= 1:
		return TroisMorts
	case d == 2:
		return DeuxMorts
	case d == 3:
		return UnMort
	case d <= 5:
		return FuiteNatifs
	}
	return Prisonniers
}

// Tableau 5 : localisation d'El Dorado.
func ZoneDe(d int) Zone {
	switch {
	case d <= 2:
		return Ouest
	case d <= 4:
		return Est
	}
	return Sud
}

func FleuveDe(d int) bool { return d >= 4 }

func NatureDe(d int) Nature {
	switch {
	case d <= 2:
		return NatJungle
	case d <= 4:
		return NatMarais
	}
	return NatSec
}

// Tableau 6 : fiabilité des renseignements, rangées « dé + code secret »
// (1..6), colonnes : distance à la case centrale (3 à 11).
var fiabilites = [6][9]int{
	{33, 25, 20, 16, 14, 12, 11, 10, 9},
	{66, 50, 40, 33, 28, 25, 22, 20, 18},
	{100, 75, 60, 50, 42, 37, 33, 30, 27},
	{100, 100, 80, 66, 57, 50, 44, 40, 36},
	{100, 100, 100, 84, 71, 62, 55, 50, 45},
	{100, 100, 100, 100, 85, 75, 66, 60, 54},
}

// DeCode : le dé plus le code secret, ramené à 1..6.
func DeCode(d, code int) int {
	v := d + code
	for v > 6 {
		v -= 6
	}
	return v
}

// Fiabilite : tableau 6 ; au-delà de 11 cases on lit la dernière colonne.
func Fiabilite(deCode, distance int) int {
	c := min(max(distance, 3), 11) - 3
	return fiabilites[deCode-1][c]
}
