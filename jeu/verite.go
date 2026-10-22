package jeu

// La zone d'El Dorado n'est fixée qu'en cours de partie : c'est celle du
// premier joueur dont les fiabilités réelles, cumulées sur une même zone,
// atteignent 100 %. Chacun ne connaît que son propre code secret ; la
// feuille commune donne le dé brut et la distance de chaque renseignement.
// Croyance énumère les codes possibles des autres joueurs (équiprobables)
// pour estimer la probabilité de chaque zone.

// Croyance : probabilités [indéterminée, Ouest, Est, Sud] vues par j.
func (e *Etat) Croyance(j int) [4]float64 {
	var res [4]float64
	if e.CodesReveles {
		res[e.ZoneVraie] = 1
		return res
	}
	// joueurs dont le code est inconnu de j et qui ont des renseignements de zone
	var inconnus []int
	vu := map[int]bool{}
	for _, r := range e.Feuille {
		if r.Type == RensZone && r.Joueur != j && !vu[r.Joueur] {
			vu[r.Joueur] = true
			inconnus = append(inconnus, r.Joueur)
		}
	}
	codes := make([]int, len(e.Exp))
	codes[j] = e.Exp[j].Code
	total := 1
	for range inconnus {
		total *= 6
	}
	somme := map[[2]int]int{}
	for n := range total {
		v := n
		for _, p := range inconnus {
			codes[p] = v%6 + 1
			v /= 6
		}
		clear(somme)
		z := SansZone
		for _, r := range e.Feuille {
			if r.Type != RensZone {
				continue
			}
			k := [2]int{r.Joueur, int(r.Zone)}
			somme[k] += Fiabilite(DeCode(r.De, codes[r.Joueur]), r.Distance)
			if somme[k] >= 100 {
				z = r.Zone
				break
			}
		}
		res[z]++
	}
	for i := range res {
		res[i] /= float64(total)
	}
	return res
}

// ZonePresumee : la zone la plus probable pour j et sa probabilité.
func (e *Etat) ZonePresumee(j int) (Zone, float64) {
	c := e.Croyance(j)
	z, p := SansZone, 0.0
	for i := Ouest; i <= Sud; i++ {
		if c[i] > p {
			z, p = i, c[i]
		}
	}
	return z, p
}
