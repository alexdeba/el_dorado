package jeu

import "fmt"

// Déroulement d'un tour, d'après l'organigramme de la règle : déplacement,
// jet de deux dés (dizaines, unités) ; pair : événement (tableau 1) ;
// impair : rencontre d'une tribu d'autant de natifs, contact ou fuite,
// attitude (tableau 3), combat (tableaux 2 et 4), renseignement (5 et 6).

const ToursMax = 400

// JouerTour : chaque expédition joue à son tour.
func (e *Etat) JouerTour() {
	e.Tour++
	for j := range e.Exp {
		if e.Fini {
			return
		}
		if e.Exp[j].Vivante {
			e.Courant = j
			e.jouerExpedition(j)
			e.verifierFin()
		}
	}
	if e.Tour >= ToursMax && !e.Fini {
		e.Fini = true
		e.noter(-1, "Les expéditions s'épuisent : la partie s'arrête sans vainqueur.")
	}
}

func (e *Etat) verifierFin() {
	if e.Fini {
		return
	}
	if e.Vivantes() == 0 {
		e.Fini = true
		e.noter(-1, "Toutes les expéditions ont péri : la jungle garde son secret.")
	}
}

func (e *Etat) jouerExpedition(j int) {
	x := e.Exp[j]
	if x.Retard > 0 {
		x.Retard--
		if x.Construction > 0 {
			e.avancerConstruction(j)
		}
		if x.Decouragee {
			e.noter(j, "%s, profondément découragée, ne joue pas.", x.Nom)
			if x.Retard == 0 {
				x.Decouragee = false
			}
		} else {
			e.noter(j, "%s est retardée et passe son tour.", x.Nom)
		}
		e.pause()
		return
	}

	// déplacement
	coups := e.Coups(j)
	c := e.Ctrl[j].Deplacer(e, j, coups)
	if !coupValide(c, coups) {
		c = coups[0]
	}
	e.appliquer(j, c)
	e.pause()

	// combat entre expéditions
	if adv := e.Occupant(x.Pos, j); adv >= 0 {
		e.combatExpeditions(j, adv)
		if !x.Vivante || x.Pos != c.Vers {
			return
		}
	}
	if e.retourEnVille(j) {
		return
	}

	// annonce d'El Dorado
	if _, ok := Centrale(x.Pos); ok && !e.DoradoTrouve && e.Ctrl[j].Annoncer(e, j) {
		e.annoncer(j)
		return
	}

	// trésor abandonné sur la case d'un massacre
	if e.TresorAuSol && e.CaseTresor == x.Pos {
		if m := e.Marqueurs[x.Pos]; m != nil && m.Amis[j] {
			e.ramasserTresor(j)
			return
		}
		if m := e.Marqueurs[x.Pos]; m != nil && !m.Exterminee && m.Absents < 0 {
			e.noter(j, "%s retrouve la case du massacre : la tribu n° %d y veille sur le trésor.", x.Nom, m.Numero)
			e.rencontre(j, m.Natifs)
			return
		}
		e.ramasserTresor(j)
		return
	}

	// jet des deux dés
	d1, d2 := e.De(), e.De()
	n := 10*d1 + d2
	e.montrerDes(j, "jet du tour", d1, d2)
	if n%2 == 0 {
		e.evenement(j, n)
	} else {
		e.rencontre(j, n)
	}
}

func coupValide(c Coup, coups []Coup) bool {
	for _, k := range coups {
		if k.Vers == c.Vers && k.Mode == c.Mode && k.Abandon == c.Abandon {
			return true
		}
	}
	return false
}

// quitter : l'expédition laisse sa case ; les natifs qui s'y cachaient de
// sa présence réapparaissent.
func (e *Etat) quitter(j int) {
	if m := e.Marqueurs[e.Exp[j].Pos]; m != nil && m.Absents == j {
		m.Absents = -1
	}
}

func (e *Etat) appliquer(j int, c Coup) {
	x := e.Exp[j]
	x.Prec = x.Pos
	switch c.Mode {
	case Rester:
		x.Construction = 0
		e.noter(j, "%s reste sur place.", x.Nom)
	case Construire:
		e.avancerConstruction(j)
	case APied, EnJangada:
		if c.Abandon && x.Jangada {
			x.Jangada = false
			e.noter(j, "%s abandonne sa jangada.", x.Nom)
		}
		x.Construction = 0
		e.quitter(j)
		x.Pos = c.Vers
		mode := "à pied"
		if c.Mode == EnJangada {
			mode = "en jangada"
		}
		e.noter(j, "%s avance %s (%s).", x.Nom, mode, TerrainDe(x.Pos))
		if Cascade(x.Pos) && c.Mode == EnJangada {
			e.noter(j, "Cascade : %s s'arrête.", x.Nom)
		}
	}
}

func (e *Etat) avancerConstruction(j int) {
	x := e.Exp[j]
	if x.Kits == 0 || x.Jangada {
		return
	}
	x.Construction++
	if x.Construction >= ToursConstruction {
		x.Construction = 0
		x.Kits--
		x.Jangada = true
		e.noter(j, "%s a fini de construire sa jangada.", x.Nom)
	} else {
		e.noter(j, "%s construit une jangada (%d/%d).", x.Nom, x.Construction, ToursConstruction)
	}
}

// retourEnVille : le trésor rapporté dans une ville termine la partie.
func (e *Etat) retourEnVille(j int) bool {
	x := e.Exp[j]
	if v := VilleEn(x.Pos); v >= 0 && x.Tresor {
		e.Fini = true
		e.Gagnant = j
		e.noter(j, "%s rapporte le trésor d'El Dorado à %s et gagne la partie !", x.Nom, Villes[v].Nom)
		e.pause()
		return true
	}
	return false
}

// ---- événements ----

func (e *Etat) evenement(j int, n int) {
	x := e.Exp[j]
	nom, eff := Evenement(n)
	switch eff {
	case EffMort:
		e.noter(j, "%d : %s. Un mort.", n, nom)
		e.pause()
		e.tuer(j, 1)
	default:
		protege := [...]Membre{EffRetardMedecin: Medecin, EffRetardGuide: Guide, EffRetardJesuite: Jesuite}[eff]
		if x.Membres[protege] > 0 {
			e.noter(j, "%d : %s. Le %s évite le retard.", n, nom, protege)
		} else {
			x.Retard++
			e.noter(j, "%d : %s. Faute de %s, l'expédition perd un tour.", n, nom, protege)
		}
		e.pause()
	}
}

// tuer : n morts, choisis par le joueur.
func (e *Etat) tuer(j, n int) {
	x := e.Exp[j]
	for range n {
		if x.Hommes() == 0 {
			break
		}
		v := e.Ctrl[j].Victime(e, j)
		if v < 0 || v >= NbMembres || x.Membres[v] == 0 {
			for m := range NbMembres {
				if x.Membres[m] > 0 {
					v = m
					break
				}
			}
		}
		x.Membres[v]--
		e.noter(j, "%s perd un %s.", x.Nom, v)
	}
	if x.Hommes() == 0 {
		e.massacre(j)
	}
}

func (e *Etat) massacre(j int) {
	x := e.Exp[j]
	x.Vivante = false
	e.quitter(j)
	e.noter(j, "L'expédition %s est anéantie.", x.Nom)
	if x.Tresor {
		x.Tresor = false
		e.TresorAuSol = true
		e.CaseTresor = x.Pos
		e.noter(j, "Le trésor d'El Dorado gît sur la case du massacre.")
	}
	e.pause()
}

// ---- rencontres ----

func (e *Etat) marqueur(k Case, natifs, j int) *Marqueur {
	m := e.Marqueurs[k]
	if m == nil {
		n := len(e.Exp)
		m = &Marqueur{Numero: e.ProchainNumero, Natifs: natifs, Amis: make([]bool, n), Renseigne: make([]bool, n), Absents: -1, Pose: j}
		e.ProchainNumero++
		e.Marqueurs[k] = m
	}
	return m
}

func (e *Etat) rencontre(j, natifs int) {
	x := e.Exp[j]
	m := e.Marqueurs[x.Pos]
	if m != nil {
		switch {
		case m.Exterminee:
			e.noter(j, "Plus aucun natif ne vit ici.")
			e.pause()
			return
		case m.Absents >= 0:
			e.noter(j, "Les natifs de la tribu n° %d restent cachés.", m.Numero)
			e.pause()
			return
		case m.Amis[j]:
			e.noter(j, "La tribu n° %d, amie de l'expédition, l'accueille.", m.Numero)
			e.pause()
			return
		}
		natifs = m.Natifs
		e.noter(j, "Rencontre de la tribu n° %d : %d natifs.", m.Numero, natifs)
	} else {
		e.noter(j, "Rencontre d'une tribu de %d natifs.", natifs)
	}
	e.pause()
	contact, cadeaux := e.Ctrl[j].Contact(e, j, natifs)
	if !contact {
		e.noter(j, "%s fuit le contact et retourne sur ses pas.", x.Nom)
		e.quitter(j)
		x.Pos = x.Prec
		e.pause()
		return
	}
	cadeaux = min(max(cadeaux, 0), x.Cadeaux)
	x.Cadeaux -= cadeaux
	d := e.De()
	mod := cadeaux
	if !x.Specialiste() {
		mod--
	}
	e.montrerDes(j, "attitude", d)
	att := AttitudeNatifs(d + mod)
	txt := ""
	if cadeaux > 0 {
		txt = fmt.Sprintf(", %d cadeau%s", cadeaux, pluriel(cadeaux))
	}
	if !x.Specialiste() {
		txt += ", sans médecin, jésuite ni guide"
	}
	e.noter(j, "Dé %d%s : les natifs sont %s.", d, txt, [...]string{"hostiles", "neutres", "amicaux"}[att])
	e.pause()
	switch att {
	case Hostile:
		e.combatNatifs(j, natifs)
	case Neutre:
		m = e.marqueur(x.Pos, natifs, j)
		e.noter(j, "Ils se cachent : pion-natif n° %d posé.", m.Numero)
		e.pause()
	case Amicale:
		m = e.marqueur(x.Pos, natifs, j)
		m.Amis[j] = true
		e.noter(j, "Ils accueillent l'expédition : pion-natif n° %d posé.", m.Numero)
		e.pause()
		e.renseignement(j, m)
	}
}

// Compte : « 3 arquebusiers », « 1 médecin », « aucun guide ».
func Compte(n int, m Membre) string {
	switch n {
	case 0:
		return "aucun " + m.String()
	case 1:
		return "1 " + m.String()
	}
	return fmt.Sprintf("%d %s", n, m.Pluriel())
}

func pluriel(n int) string {
	if n > 1 {
		return "x"
	}
	return ""
}

func (e *Etat) combatNatifs(j, natifs int) {
	x := e.Exp[j]
	mod := RapportForces(x.Membres[Arquebusier], natifs)
	d := e.De()
	e.montrerDes(j, "combat", d)
	v := min(max(d+mod, 1), 6)
	issue := ResultatCombat(v)
	e.noter(j, "Combat : %s contre %d natifs, dé %d %+d = %d.", Compte(x.Membres[Arquebusier], Arquebusier), natifs, d, mod, v)
	e.pause()
	switch issue {
	case TroisMorts, DeuxMorts, UnMort:
		n := 3 - int(issue)
		e.noter(j, "%d mort%s dans l'expédition.", n, map[bool]string{true: "s"}[n > 1])
		e.pause()
		e.tuer(j, n)
		m := e.marqueur(x.Pos, natifs, j)
		if x.Vivante {
			m.Exterminee = true
			e.noter(j, "La tribu est exterminée : plus de natifs sur cette case.")
		} else {
			e.noter(j, "La tribu n° %d reste maîtresse de la case.", m.Numero)
		}
	case FuiteNatifs:
		m := e.marqueur(x.Pos, natifs, j)
		m.Absents = j
		e.noter(j, "Les natifs s'enfuient et se cachent : pion-natif n° %d.", m.Numero)
	case Prisonniers:
		m := e.marqueur(x.Pos, natifs, j)
		m.Absents = j
		e.noter(j, "Les natifs sont faits prisonniers : pion-natif n° %d.", m.Numero)
		e.pause()
		e.renseignement(j, m)
	}
	e.pause()
}

// ---- renseignements ----

func (e *Etat) renseignement(j int, m *Marqueur) {
	x := e.Exp[j]
	if e.TresorAuSol && e.CaseTresor == x.Pos {
		e.ramasserTresor(j)
		return
	}
	if e.DoradoTrouve {
		return // le secret est éventé : plus rien à apprendre
	}
	if m.Renseigne[j] {
		e.noter(j, "Cette tribu a déjà renseigné l'expédition.")
		e.pause()
		return
	}
	m.Renseigne[j] = true
	dist := Distance(x.Pos, Centre)
	d := e.De()
	r := Renseignement{Joueur: j, Tour: e.Tour, Distance: dist, De: d, Fiab: 100}
	if DansRegion(x.Pos) {
		switch {
		case e.FleuveCentral < 0:
			r.Type = RensFleuve
			r.Fleuve = FleuveDe(d)
			e.FleuveCentral = map[bool]int{false: 0, true: 1}[r.Fleuve]
		case e.NatureCentrale == NatInconnue:
			r.Type = RensNature
			r.Nature = NatureDe(d)
			e.NatureCentrale = r.Nature
		default:
			e.noter(j, "Les natifs n'ont plus rien à apprendre sur la région centrale.")
			e.pause()
			return
		}
		e.montrerDes(j, "renseignement", d)
		e.Feuille = append(e.Feuille, r)
		e.noter(j, "Renseignement sûr (région centrale) : El Dorado — %s.", r.Texte())
		e.pause()
		return
	}
	r.Type = RensZone
	r.Zone = ZoneDe(d)
	f := e.De()
	r.De = f
	r.Fiab = Fiabilite(DeCode(f, x.Code), dist)
	e.montrerDes(j, "renseignement", d, f)
	e.Feuille = append(e.Feuille, r)
	if e.ZoneVraie == SansZone {
		k := [2]int{j, int(r.Zone)}
		e.sommes[k] += r.Fiab
		if e.sommes[k] >= 100 {
			e.ZoneVraie = r.Zone
		}
	}
	e.noter(j, "Renseignement : El Dorado serait dans la zone %s (%d cases du centre, dé de fiabilité %d).", r.Zone, dist, f)
	e.pause()
}

// ---- El Dorado ----

func (e *Etat) annoncer(j int) {
	x := e.Exp[j]
	e.noter(j, "%s annonce « El Dorado » ! Les codes secrets sont dévoilés.", x.Nom)
	e.CodesReveles = true
	e.pause()
	if k, ok := e.Dorado(); ok && k == x.Pos {
		e.DoradoTrouve = true
		x.Tresor = true
		e.noter(j, "C'est bien la cité d'or ! %s s'empare du trésor et doit le rapporter dans une ville.", x.Nom)
	} else {
		x.Retard = 2
		x.Decouragee = true
		if ok {
			e.noter(j, "Ce n'est pas El Dorado. L'expédition, profondément découragée, ne jouera pas pendant deux tours. La bonne case est désormais connue.")
		} else {
			e.noter(j, "Ce n'est pas El Dorado : les renseignements ne désignent encore aucune case. L'expédition, profondément découragée, ne jouera pas pendant deux tours.")
		}
	}
	e.pause()
}

func (e *Etat) ramasserTresor(j int) {
	x := e.Exp[j]
	e.TresorAuSol = false
	x.Tresor = true
	e.noter(j, "%s récupère le trésor d'El Dorado !", x.Nom)
	e.pause()
}

// ---- combats entre expéditions ----

// combatExpeditions : a vient d'entrer sur la case de d ; a est l'attaquant
// et jette le dé le premier. Le plus fort tue un homme adverse, l'égalité
// coûte un homme à l'attaquant. À chaque passe, chacun peut fuir.
func (e *Etat) combatExpeditions(a, d int) {
	xa, xd := e.Exp[a], e.Exp[d]
	e.noter(a, "%s (%d hommes) attaque %s (%d hommes).", xa.Nom, xa.Hommes(), xd.Nom, xd.Hommes())
	e.pause()
	surDorado := false
	if k, ok := e.Dorado(); ok && e.DoradoTrouve && k == xa.Pos {
		surDorado = true
	}
	fuir := func(f, autre int) {
		xf := e.Exp[f]
		e.noter(f, "%s prend la fuite.", xf.Nom)
		if xf.Tresor && surDorado {
			xf.Tresor = false
			e.Exp[autre].Tresor = true
			e.noter(autre, "%s garde le trésor sur la case d'El Dorado.", e.Exp[autre].Nom)
		}
		e.quitter(f)
		xf.Pos = xf.Prec
		e.pause()
	}
	for xa.Vivante && xd.Vivante {
		if !e.Ctrl[a].Combattre(e, a, d) {
			fuir(a, d)
			return
		}
		if xd.Prec != xd.Pos && !e.Ctrl[d].Combattre(e, d, a) {
			fuir(d, a)
			return
		}
		da, dd := e.De(), e.De()
		e.montrerDes(a, "combat", da, dd)
		perdant := a
		if da > dd {
			perdant = d
		}
		e.noter(a, "Dés %d contre %d : %s perd un homme.", da, dd, e.Exp[perdant].Nom)
		e.pause()
		e.tuer(perdant, 1)
	}
	gagnant, perdant := a, d
	if !xa.Vivante {
		gagnant, perdant = d, a
	}
	if e.TresorAuSol && e.CaseTresor == e.Exp[gagnant].Pos && !e.Exp[perdant].Vivante && e.Exp[perdant].Pos == e.CaseTresor {
		// le trésor du vaincu revient au vainqueur
		e.TresorAuSol = false
		e.Exp[gagnant].Tresor = true
		e.noter(gagnant, "%s s'empare du trésor d'El Dorado !", e.Exp[gagnant].Nom)
		e.pause()
	}
}
