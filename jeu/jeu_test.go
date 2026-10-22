package jeu

import (
	"math/rand/v2"
	"testing"
)

func TestTableaux(t *testing.T) {
	if _, e := Evenement(36); e != EffMort {
		t.Error("36 doit tuer")
	}
	if n, e := Evenement(44); e != EffRetardMedecin || n != "fruits empoisonnés, diarrhée" {
		t.Error("44", n, e)
	}
	if _, e := Evenement(56); e != EffRetardGuide {
		t.Error("56")
	}
	if _, e := Evenement(62); e != EffRetardJesuite {
		t.Error("62")
	}
	cas := []struct{ arq, nat, v int }{
		{1, 10, -1}, {7, 15, 1}, {25, 65, 5}, {4, 33, -2}, {12, 47, 0}, {17, 60, 1}, {0, 20, -3}, {30, 10, 5},
	}
	for _, c := range cas {
		if v := RapportForces(c.arq, c.nat); v != c.v {
			t.Errorf("rapport %d arquebusiers / %d natifs = %d, attendu %d", c.arq, c.nat, v, c.v)
		}
	}
	if Fiabilite(DeCode(3, 2), 7) != 71 || Fiabilite(DeCode(4, 3), 3) != 33 || Fiabilite(DeCode(5, 4), 3) != 100 || Fiabilite(6, 20) != 54 {
		t.Error("tableau 6")
	}
	if AttitudeNatifs(0) != Hostile || AttitudeNatifs(4) != Neutre || AttitudeNatifs(7) != Amicale {
		t.Error("tableau 3")
	}
	if ResultatCombat(-2) != TroisMorts || ResultatCombat(4) != FuiteNatifs || ResultatCombat(9) != Prisonniers {
		t.Error("tableau 4")
	}
}

func TestRegionCentrale(t *testing.T) {
	if len(regionCentrale) != 18 {
		t.Fatal(len(regionCentrale))
	}
	vu := map[CaseCentrale]bool{}
	par := map[Zone]int{}
	for k, c := range regionCentrale {
		if d := Distance(k, Centre); d < 1 || d > 2 {
			t.Errorf("%v à %d du centre", k, d)
		}
		if vu[c] {
			t.Errorf("combinaison en double %+v", c)
		}
		vu[c] = true
		par[c.Zone]++
		if CaseDorado(c.Zone, c.Fleuve, c.Nature) != k {
			t.Error("CaseDorado", k)
		}
		if !Terre(k) {
			t.Error("mer", k)
		}
	}
	for z := Ouest; z <= Sud; z++ {
		if par[z] != 6 {
			t.Errorf("zone %v : %d cases", z, par[z])
		}
	}
	n := 0
	for _, k := range Cases() {
		if DansRegion(k) {
			n++
		}
	}
	if n != 19 {
		t.Errorf("%d cases dans la région centrale", n)
	}
	if !Fleuve(Case{13, 10}) || Fleuve(Case{12, 10}) || TerrainDe(Case{10, 12}) != Marais {
		t.Error("relevé de la région centrale")
	}
}

func TestPlateau(t *testing.T) {
	if len(Villes) != 6 {
		t.Fatal("villes")
	}
	for _, v := range Villes {
		if !Accessible(v.Case) {
			t.Errorf("%s n'est pas sur la terre ferme", v.Nom)
		}
		// chaque ville rejoint le centre à pied
		vu := map[Case]bool{v.Case: true}
		file := []Case{v.Case}
		for len(file) > 0 {
			k := file[0]
			file = file[1:]
			for _, w := range Voisins(k) {
				if Accessible(w) && !vu[w] {
					vu[w] = true
					file = append(file, w)
				}
			}
		}
		if !vu[Centre] {
			t.Errorf("%s ne rejoint pas le centre à pied", v.Nom)
		}
	}
	for _, tr := range tronconsPlateau {
		if !Voisines(tr.Amont, tr.Aval) || !Terre(tr.Amont) || !Terre(tr.Aval) {
			t.Errorf("tronçon %+v", tr)
		}
	}
	for _, k := range cascadesPlateau {
		if !Fleuve(k) {
			t.Errorf("cascade %v hors fleuve", k)
		}
	}
	// chaque case : au plus un aval et un amont par fleuve
	for k, ls := range liensFleuve {
		n := map[[2]int]int{}
		for _, l := range ls {
			b := 0
			if l.aval {
				b = 1
			}
			n[[2]int{l.fleuve, b}]++
		}
		for f, c := range n {
			if c > 1 {
				t.Errorf("case %v : fleuve %d bifurque (%d)", k, f[0], c)
			}
		}
	}
	if !Fleuve(Villes[0].Case) || !Fleuve(Villes[2].Case) || !Fleuve(Villes[5].Case) {
		t.Error("Quito, Belem et Buenos Aires sont sur un fleuve")
	}
	if !Cotiere(Villes[3].Case) || !Cotiere(Villes[4].Case) {
		t.Error("Bahia et Rio sont sur la côte")
	}
}

func destinations(cs []Coup, m Mode) map[Case]Coup {
	r := map[Case]Coup{}
	for _, c := range cs {
		if c.Mode == m {
			r[c.Vers] = c
		}
	}
	return r
}

func TestDeplacements(t *testing.T) {
	// à pied : jamais dans le marais
	for _, k := range Cases() {
		for _, c := range CoupsDepuis(k, false, 0, nil) {
			if c.Mode == APied && TerrainDe(c.Vers) == Marais {
				t.Fatalf("à pied dans le marais %v", c.Vers)
			}
			if c.Mode == EnJangada {
				t.Fatal("jangada sans jangada")
			}
		}
	}
	// en jangada : le chemin suit un seul fleuve, 3 cases en descendant, 2 en
	// remontant, arrêt sur cascade et marais ; destinations navigables
	for _, k := range Cases() {
		for _, c := range CoupsDepuis(k, true, 0, nil) {
			if c.Mode != EnJangada {
				continue
			}
			if !Navigable(c.Vers) && !Cotiere(c.Vers) {
				t.Errorf("%v → %v : ni fleuve, ni marais, ni côte", k, c.Vers)
			}
			if len(c.Chemin) > 3 || (len(c.Chemin) == 3 && !c.Descente) {
				t.Errorf("%v → %v : chemin trop long %v", k, c.Vers, c.Chemin)
			}
			for _, p := range c.Chemin[:len(c.Chemin)-1] {
				if Cascade(p) || TerrainDe(p) == Marais {
					t.Errorf("%v → %v : traverse %v sans s'arrêter", k, c.Vers, p)
				}
			}
		}
	}
	// descente de l'Amazone depuis Quito : trois cases
	q := Villes[0].Case
	n := 0
	for _, c := range CoupsDepuis(q, true, 0, nil) {
		if c.Mode == EnJangada && len(c.Chemin) == 3 {
			n++
		}
	}
	if n == 0 {
		t.Error("de Quito, aucune descente de trois cases")
	}
	// une cascade arrête la descente
	casc := Case{14, 8}
	for _, l := range liensFleuve[casc] {
		if l.aval {
			continue
		}
		amont := l.vers
		for _, c := range CoupsDepuis(amont, true, 0, nil) {
			for i, p := range c.Chemin {
				if p == casc && i != len(c.Chemin)-1 {
					t.Error("la cascade n'arrête pas")
				}
			}
		}
	}
	// une expédition sur le chemin arrête aussi
	occ := func(k Case) bool { return k == (Case{5, 4}) }
	for _, c := range CoupsDepuis(q, true, 0, occ) {
		for i, p := range c.Chemin {
			if p == (Case{5, 4}) && i != len(c.Chemin)-1 {
				t.Error("passage à travers une expédition")
			}
		}
	}
	// construction : seulement avec du matériel, sur une case navigable
	if len(destinations(CoupsDepuis(q, false, 1, nil), Construire)) != 1 {
		t.Error("construire à Quito")
	}
	if len(destinations(CoupsDepuis(Case{14, 12}, false, 1, nil), Construire)) != 0 {
		t.Error("construire hors fleuve")
	}
}

// controleur scripté pour les tests
type script struct {
	IA
	contact bool
	annonce bool
}

func (s *script) Contact(e *Etat, j, n int) (bool, int)  { return s.contact, 0 }
func (s *script) Annoncer(e *Etat, j int) bool           { return s.annonce }
func (s *script) Deplacer(e *Etat, j int, c []Coup) Coup { return c[0] }

func partieTest(n int) *Etat {
	var exps []*Expedition
	var ctrls []Controleur
	for i := range n {
		a := Achats{Membres: [NbMembres]int{5, 1, 1, 1}, Cadeaux: 5}
		exps = append(exps, NouvelleExpedition("E"+string(rune('A'+i)), i, a, i, i+1))
		ctrls = append(ctrls, &script{contact: true})
	}
	return Nouveau(exps, ctrls, 7)
}

func TestVerite(t *testing.T) {
	e := partieTest(3)
	// joueur 1 (code 2) : dé 4 à 3 cases → DeCode 6 → 100 % : la zone est fixée
	e.Feuille = append(e.Feuille, Renseignement{Joueur: 1, Type: RensZone, Zone: Est, Distance: 3, De: 4, Fiab: 100})
	c := e.Croyance(0)
	// vu de 0 : le code de 1 est inconnu ; dé 4 + code c ≥ 3 sur 1..6 → 100 % pour c ∉ {5}
	// (4+5 = 9 → 3 → 100 aussi), seul c = 3 (→ 1, 33 %) et c = 4 (→ 2, 66 %) échouent
	if c[Est] < 0.66 || c[Est] > 0.67 || c[SansZone] < 0.33 {
		t.Errorf("croyance %v", c)
	}
	if c1 := e.Croyance(1); c1[Est] != 1 {
		t.Errorf("le joueur 1 connaît son code : %v", c1)
	}
	e.CodesReveles = true
	e.ZoneVraie = Est
	if c := e.Croyance(2); c[Est] != 1 {
		t.Error("codes révélés")
	}
}

func TestRenseignementsCentraux(t *testing.T) {
	e := partieTest(2)
	x := e.Exp[0]
	x.Pos = Case{11, 11}
	m := e.marqueur(x.Pos, 21, 0)
	e.renseignement(0, m)
	if e.FleuveCentral < 0 || e.Feuille[0].Type != RensFleuve || e.Feuille[0].Fiab != 100 {
		t.Fatal("premier renseignement central : le fleuve")
	}
	m2 := e.marqueur(Case{11, 12}, 31, 1)
	e.Exp[1].Pos = Case{11, 12}
	e.renseignement(1, m2)
	if e.NatureCentrale == NatInconnue {
		t.Fatal("second : la nature du terrain")
	}
	e.renseignement(1, m2)
	if len(e.Feuille) != 2 {
		t.Error("une tribu ne renseigne qu'une fois un même joueur")
	}
}

func TestAnnonceManquee(t *testing.T) {
	e := partieTest(2)
	x := e.Exp[0]
	x.Pos = Case{11, 11}
	e.annoncer(0)
	if !x.Decouragee || x.Retard != 2 || !e.CodesReveles || e.DoradoTrouve {
		t.Error("annonce sans renseignements")
	}
	e.ZoneVraie, e.FleuveCentral, e.NatureCentrale = Ouest, 1, NatJungle
	y := e.Exp[1]
	y.Pos = Case{11, 11}
	e.annoncer(1)
	if !y.Tresor || !e.DoradoTrouve {
		t.Error("bonne annonce")
	}
	// retour en ville
	y.Pos = Villes[1].Case
	if !e.retourEnVille(1) || e.Gagnant != 1 {
		t.Error("victoire")
	}
}

func TestMassacreEtTresor(t *testing.T) {
	e := partieTest(2)
	x := e.Exp[0]
	x.Tresor = true
	x.Pos = Case{9, 9}
	x.Membres = [NbMembres]int{1, 0, 0, 0}
	e.marqueur(x.Pos, 40, 0)
	e.tuer(0, 1)
	if x.Vivante || !e.TresorAuSol || e.CaseTresor != (Case{9, 9}) {
		t.Fatal("massacre")
	}
	y := e.Exp[1]
	y.Pos = Case{9, 9}
	e.Marqueurs[y.Pos].Amis[1] = true
	e.renseignement(1, e.Marqueurs[y.Pos])
	if !y.Tresor || e.TresorAuSol {
		t.Error("tribu amie : le trésor revient")
	}
}

func TestParties(t *testing.T) {
	tours, gagnees, n := 0, 0, 40
	for g := range n {
		rng := rand.New(rand.NewPCG(uint64(g), 1))
		var exps []*Expedition
		var ctrls []Controleur
		libres := []int{0, 1, 2, 3, 4, 5}
		nb := 2 + g%5
		for i := range nb {
			au := float64(i) / float64(nb)
			a, v, code := PreparerIA(rng, libres, au)
			if a.Cout() > Budget {
				t.Fatalf("budget %d", a.Cout())
			}
			for k, l := range libres {
				if l == v {
					libres = append(libres[:k], libres[k+1:]...)
					break
				}
			}
			exps = append(exps, NouvelleExpedition(string(rune('A'+i)), i, a, v, code))
			ctrls = append(ctrls, NouvelleIA(au))
		}
		e := Nouveau(exps, ctrls, uint64(g))
		for !e.Fini {
			e.JouerTour()
		}
		if e.Tour >= ToursMax {
			t.Errorf("partie %d : pas de fin en %d tours", g, ToursMax)
		}
		tours += e.Tour
		if e.Gagnant >= 0 {
			gagnees++
		}
	}
	t.Logf("%d parties : %.1f tours en moyenne, %d gagnées", n, float64(tours)/float64(n), gagnees)
}
