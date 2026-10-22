package jeu

// Déplacements : à pied, une case par tour sur tout terrain sauf le marais ;
// en jangada, sur les fleuves (3 cases en descendant, 2 en remontant, un seul
// fleuve par tour, arrêt sur les cascades), dans les marais (l'entrée termine
// le déplacement) et le long des côtes (2 cases par tour).

type Mode int

const (
	Rester Mode = iota
	APied
	EnJangada
	Construire
)

type Coup struct {
	Vers     Case
	Mode     Mode
	Abandon  bool   // abandonner la jangada pour continuer à pied
	Chemin   []Case // cases traversées, arrivée comprise
	Descente bool   // en jangada sur un fleuve : vers l'aval
}

// Accessible : on peut y entrer à pied.
func Accessible(k Case) bool {
	t := TerrainDe(k)
	return t != Mer && t != Marais
}

// Navigable : on peut y entrer en jangada.
func Navigable(k Case) bool {
	return Fleuve(k) || TerrainDe(k) == Marais
}

// CoupsDepuis : déplacements possibles depuis k, en jangada ou non. occupe
// dit si une autre expédition se trouve sur une case (on s'y arrête).
func CoupsDepuis(k Case, jangada bool, kits int, occupe func(Case) bool) []Coup {
	coups := []Coup{{Vers: k, Mode: Rester}}
	if !jangada && kits > 0 && Navigable(k) {
		coups = append(coups, Coup{Vers: k, Mode: Construire})
	}
	for _, v := range Voisins(k) {
		if Accessible(v) {
			coups = append(coups, Coup{Vers: v, Mode: APied, Abandon: jangada, Chemin: []Case{v}})
		}
	}
	if jangada {
		coups = append(coups, coupsJangada(k, occupe)...)
	}
	return coups
}

func coupsJangada(k Case, occupe func(Case) bool) []Coup {
	vu := map[Case]bool{k: true}
	var coups []Coup
	ajoute := func(c Coup) {
		if !vu[c.Vers] {
			vu[c.Vers] = true
			coups = append(coups, c)
		}
	}
	arret := func(v Case) bool {
		return Cascade(v) || TerrainDe(v) == Marais || (occupe != nil && occupe(v))
	}
	// le long d'un fleuve, dans un seul sens
	for _, l := range liensFleuve[k] {
		maxi := 2
		if l.aval {
			maxi = 3
		}
		chemin := []Case{}
		cur, suiv := k, l.vers
		for pas := 1; ; pas++ {
			chemin = append(chemin, suiv)
			ajoute(Coup{Vers: suiv, Mode: EnJangada, Chemin: append([]Case(nil), chemin...), Descente: l.aval})
			if pas == maxi || arret(suiv) {
				break
			}
			cur, suiv = suiv, Case{-1, -1}
			for _, m := range liensFleuve[cur] {
				if m.fleuve == l.fleuve && m.aval == l.aval {
					suiv = m.vers
				}
			}
			if !suiv.Valide() {
				break
			}
		}
	}
	// entrer dans un marais voisin, ou en sortir vers une case à fleuve
	marais := TerrainDe(k) == Marais
	for _, v := range Voisins(k) {
		if !Terre(v) {
			continue
		}
		vm := TerrainDe(v) == Marais
		if (vm && !marais) || (marais && !vm && Fleuve(v)) {
			ajoute(Coup{Vers: v, Mode: EnJangada, Chemin: []Case{v}})
		}
	}
	// le long de la côte, deux cases
	if Cotiere(k) {
		for _, v := range Voisins(k) {
			if !Cotiere(v) {
				continue
			}
			ajoute(Coup{Vers: v, Mode: EnJangada, Chemin: []Case{v}})
			if arret(v) {
				continue
			}
			for _, w := range Voisins(v) {
				if w != k && Cotiere(w) {
					ajoute(Coup{Vers: w, Mode: EnJangada, Chemin: []Case{v, w}})
				}
			}
		}
	}
	return coups
}

// Coups : déplacements possibles pour l'expédition j.
func (e *Etat) Coups(j int) []Coup {
	x := e.Exp[j]
	return CoupsDepuis(x.Pos, x.Jangada, x.Kits, func(k Case) bool { return e.Occupant(k, j) >= 0 })
}
