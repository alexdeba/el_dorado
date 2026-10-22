package jeu

import (
	"math/rand/v2"
	"slices"
)

// IA : un conquistador mené par l'ordinateur. Il rejoint la ceinture des
// cases proches de la région centrale (où les renseignements sont les plus
// fiables), y cherche les tribus, entre dans la région pour les deux
// renseignements sûrs, annonce El Dorado quand sa croyance est assez forte,
// puis rapporte le trésor dans la ville la plus proche.
type IA struct {
	Audace float64 // 0 prudent … 1 téméraire
	Seuil  float64 // probabilité de zone pour annoncer
}

func NouvelleIA(audace float64) *IA {
	return &IA{Audace: audace, Seuil: 0.9 - 0.25*audace}
}

// ---- organisation ----

// PreparerIA : achats, ville et code d'une expédition de l'ordinateur.
func PreparerIA(rng *rand.Rand, villesLibres []int, audace float64) (Achats, int, int) {
	// la ville libre la plus proche du centre, à égalité au hasard
	slices.SortFunc(villesLibres, func(a, b int) int {
		return Distance(Villes[a].Case, Centre) - Distance(Villes[b].Case, Centre)
	})
	ville := villesLibres[0]
	if len(villesLibres) > 1 && rng.IntN(3) == 0 {
		ville = villesLibres[1]
	}
	var a Achats
	a.Membres = [NbMembres]int{Arquebusier: 7, Medecin: 1, Jesuite: 1, Guide: 1}
	if audace > 0.6 {
		a.Membres[Arquebusier] = 8
	}
	// en jangada depuis une ville sur un fleuve, si le fleuve mène vers le centre
	if Fleuve(Villes[ville].Case) && rng.IntN(2) == 0 {
		a.Membres[Arquebusier] -= 2
		a.Jangadas = 1
		a.EnJangada = true
	}
	a.Cadeaux = (Budget - a.Cout()) / PrixCadeau
	return a, ville, rng.IntN(6) + 1
}

// ---- trajets ----

type etatTrajet struct {
	k Case
	j bool // en jangada
}

// distances : nombre de tours pour atteindre chaque case depuis s.
func distancesDepuis(s etatTrajet, kits int) (map[etatTrajet]int, map[etatTrajet]etatTrajet, map[etatTrajet]Coup) {
	dist := map[etatTrajet]int{s: 0}
	prec := map[etatTrajet]etatTrajet{}
	coup := map[etatTrajet]Coup{}
	file := []etatTrajet{s}
	for len(file) > 0 {
		u := file[0]
		file = file[1:]
		for _, c := range CoupsDepuis(u.k, u.j, 0, nil) {
			if c.Mode == Rester {
				continue
			}
			v := etatTrajet{c.Vers, u.j && !c.Abandon}
			if _, ok := dist[v]; ok {
				continue
			}
			dist[v] = dist[u] + 1
			prec[v] = u
			coup[v] = c
			file = append(file, v)
		}
	}
	return dist, prec, coup
}

// versCible : premier coup du plus court chemin (en tours) vers une des
// cibles, la plus proche selon cout ; rester si aucune n'est atteignable.
func (ia *IA) versCible(e *Etat, j int, coups []Coup, cibles func(Case) bool, eviter func(Case) bool) Coup {
	x := e.Exp[j]
	s := etatTrajet{x.Pos, x.Jangada}
	dist, prec, coup := distancesDepuis(s, x.Kits)
	best, bd := etatTrajet{}, 1<<30
	for v, d := range dist {
		if cibles(v.k) && (d < bd || (d == bd && (less(v.k, best.k) || (v.k == best.k && !v.j && best.j)))) {
			best, bd = v, d
		}
	}
	if bd == 1<<30 || bd == 0 {
		return coups[0]
	}
	v := best
	for prec[v] != s {
		v = prec[v]
	}
	c := coup[v]
	for _, k := range coups {
		if k.Vers == c.Vers && k.Mode == c.Mode && k.Abandon == c.Abandon && (eviter == nil || !eviter(k.Vers)) {
			return k
		}
	}
	// chemin bloqué (une expédition sur la route) : un pas à pied qui rapproche
	for _, k := range coups {
		if k.Mode == APied && !k.Abandon && (eviter == nil || !eviter(k.Vers)) {
			if d, ok := dist[etatTrajet{k.Vers, false}]; ok && d < bd {
				return k
			}
		}
	}
	return coups[0]
}

func less(a, b Case) bool {
	if a.R != b.R {
		return a.R < b.R
	}
	return a.C < b.C
}

// ---- décisions ----

func (ia *IA) Deplacer(e *Etat, j int, coups []Coup) Coup {
	x := e.Exp[j]
	eviter := func(k Case) bool {
		o := e.Occupant(k, j)
		return o >= 0 && !ia.attaquer(e, j, o)
	}
	// rapporter le trésor
	if x.Tresor {
		return ia.versCible(e, j, coups, func(k Case) bool { return VilleEn(k) >= 0 }, eviter)
	}
	// reprendre le trésor
	if e.TresorAuSol {
		return ia.versCible(e, j, coups, func(k Case) bool { return k == e.CaseTresor }, eviter)
	}
	if e.DoradoTrouve {
		for o, y := range e.Exp {
			if o != j && y.Vivante && y.Tresor && ia.attaquer(e, j, o) {
				return ia.versCible(e, j, coups, func(k Case) bool { return k == y.Pos }, nil)
			}
		}
		return coups[0]
	}
	// la case est connue de tous après une annonce manquée, ou presque sûre
	if k, ok := ia.cible(e, j); ok {
		return ia.versCible(e, j, coups, func(c Case) bool { return c == k }, eviter)
	}
	// renseignements sûrs de la région centrale
	if e.FleuveCentral < 0 || e.NatureCentrale == NatInconnue {
		if d := Distance(x.Pos, Centre); d <= 2 && d >= 1 {
			return ia.errer(e, j, coups, 1, 2, eviter)
		}
		if z, p := e.ZonePresumee(j); p > 0.3 {
			return ia.versCible(e, j, coups, func(k Case) bool {
				c, ok := Centrale(k)
				return ok && c.Zone == z
			}, eviter)
		}
		return ia.versCible(e, j, coups, func(k Case) bool { _, ok := Centrale(k); return ok }, eviter)
	}
	// la zone reste incertaine : chercher des tribus près du centre
	if d := Distance(x.Pos, Centre); d >= 3 && d <= 4 {
		return ia.errer(e, j, coups, 3, 4, eviter)
	}
	return ia.versCible(e, j, coups, func(k Case) bool {
		d := Distance(k, Centre)
		return d == 3 && e.Marqueurs[k] == nil
	}, eviter)
}

// cible : la case d'El Dorado selon j, s'il en est assez sûr.
func (ia *IA) cible(e *Etat, j int) (Case, bool) {
	if e.FleuveCentral < 0 || e.NatureCentrale == NatInconnue {
		return Case{}, false
	}
	z, p := e.ZonePresumee(j)
	if z == SansZone || p < ia.Seuil {
		return Case{}, false
	}
	return CaseDorado(z, e.FleuveCentral == 1, e.NatureCentrale), true
}

// errer : une case voisine à distance du centre entre dmin et dmax, de
// préférence sans pion-natif ami (une tribu ne renseigne qu'une fois).
func (ia *IA) errer(e *Etat, j int, coups []Coup, dmin, dmax int, eviter func(Case) bool) Coup {
	var bons, moyens []Coup
	for _, c := range coups {
		if c.Mode == Rester || c.Mode == Construire || c.Abandon || (eviter != nil && eviter(c.Vers)) {
			continue
		}
		d := Distance(c.Vers, Centre)
		if d < dmin || d > dmax || c.Vers == Centre {
			continue
		}
		m := e.Marqueurs[c.Vers]
		if m == nil {
			bons = append(bons, c)
		} else if !m.Renseigne[j] && !m.Exterminee {
			moyens = append(moyens, c)
		}
	}
	switch {
	case len(bons) > 0:
		return bons[e.rng.IntN(len(bons))]
	case len(moyens) > 0:
		return moyens[e.rng.IntN(len(moyens))]
	}
	for _, c := range coups {
		if c.Mode == APied && !c.Abandon && (eviter == nil || !eviter(c.Vers)) && Distance(c.Vers, Centre) >= dmin {
			return c
		}
	}
	return coups[0]
}

func (ia *IA) Annoncer(e *Etat, j int) bool {
	k, ok := ia.cible(e, j)
	return ok && k == e.Exp[j].Pos
}

func (ia *IA) Contact(e *Etat, j, natifs int) (bool, int) {
	x := e.Exp[j]
	// sur la case du trésor abandonné, fuir ne mène à rien : on tente sa chance
	if e.TresorAuSol && e.CaseTresor == x.Pos {
		return true, x.Cadeaux
	}
	if x.Hommes() <= 2 && natifs >= 40 && x.Pos != x.Prec {
		return false, 0
	}
	if x.Tresor && x.Membres[Arquebusier] < 3 && natifs >= 30 && x.Pos != x.Prec {
		return false, 0
	}
	cadeaux := 1
	if !x.Specialiste() {
		cadeaux = 2
	}
	if DansRegion(x.Pos) || x.Tresor || (e.TresorAuSol && e.CaseTresor == x.Pos) {
		cadeaux++
	}
	return true, min(cadeaux, x.Cadeaux)
}

// Victime : on sacrifie d'abord les arquebusiers en surnombre, puis les
// spécialistes en double, en gardant un médecin, un jésuite et un guide.
func (ia *IA) Victime(e *Etat, j int) Membre {
	x := e.Exp[j]
	if x.Membres[Arquebusier] > 2 {
		return Arquebusier
	}
	for _, m := range []Membre{Jesuite, Guide, Medecin} {
		if x.Membres[m] > 1 {
			return m
		}
	}
	if x.Membres[Arquebusier] > 0 {
		return Arquebusier
	}
	for _, m := range []Membre{Jesuite, Guide, Medecin} {
		if x.Membres[m] > 0 {
			return m
		}
	}
	return Arquebusier
}

// attaquer : vaut-il la peine d'affronter l'expédition o ?
func (ia *IA) attaquer(e *Etat, j, o int) bool {
	a, b := e.Exp[j].Hommes(), e.Exp[o].Hommes()
	if e.Exp[o].Tresor {
		return float64(a) >= float64(b)*(1.3-0.5*ia.Audace)
	}
	return float64(a) >= float64(b)*(2-ia.Audace)
}

func (ia *IA) Combattre(e *Etat, j, o int) bool {
	a, b := e.Exp[j].Hommes(), e.Exp[o].Hommes()
	if e.Exp[o].Tresor || e.Exp[j].Tresor {
		return a+1 >= b
	}
	return float64(a) >= float64(b)*(1.2-0.4*ia.Audace)
}
