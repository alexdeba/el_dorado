package main

import (
	"fmt"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"eldorado/jeu"
)

// Le panneau de droite : en-tête, expéditions, consigne et décisions,
// journal de route, dés et chronique.

const (
	pnX   = PanneauX + 10
	pnW   = LargPanneau - 20
	yExp  = 82.0
	yCons = 138.0
	hCons = 268.0
	yJour = yCons + hCons + 10
	hJour = 190.0
	yChro = yJour + hJour + 10
	hChro = 938 - yChro
)

func (g *Game) dessinerPanneau(dst *ebiten.Image) {
	e := g.p.E
	fondCuir(dst, PanneauX, 0, LargPanneau, HauteurEcran)
	frise(dst, PanneauX, 0, LargPanneau, 26)
	ligne(dst, PanneauX, 0, PanneauX, HauteurEcran, 3, coulOrSombre)

	// en-tête
	ecrire(dst, fmt.Sprintf("Tour %d", max(e.Tour, 1)), pnX+6, 32, face(26, policeTitre), coulOr)
	if !e.Fini {
		x := e.Exp[e.Courant]
		s := x.Nom + " joue"
		if e.Courant == g.p.Humain {
			s = "À vous de jouer"
		}
		ecrireDroite(dst, s, pnX+pnW-4, 38, face(24, policeGothique), coulParch)
	}

	// expéditions
	n := len(e.Exp)
	w := pnW / float64(n)
	for i, x := range e.Exp {
		cx := pnX + w*float64(i) + w/2
		s := 30.0
		if i == e.Courant && !e.Fini {
			rect(dst, cx-w/2+3, yExp-4, w-6, 52, avecAlpha(coulOr, 0x50))
		}
		cs := &ebiten.ColorScale{}
		if !x.Vivante {
			cs.Scale(0.35, 0.3, 0.3, 1)
		}
		dessinerImage(dst, imgExpeditions[x.Couleur], cx-s/2, yExp, s, s, cs)
		cadre(dst, cx-s/2, yExp, s, s, 1.5, coulOrSombre)
		nom := x.Nom
		if i == g.p.Humain {
			nom = "Vous"
		}
		ecrireCentre(dst, nom, cx, yExp+30, face(13, policeTexte), coulParch2)
		if !x.Vivante {
			ligne(dst, cx-s/2, yExp, cx+s/2, yExp+s, 2.5, coulSang)
			ligne(dst, cx+s/2, yExp, cx-s/2, yExp+s, 2.5, coulSang)
		}
		if x.Tresor {
			coffre(dst, cx+s/2, yExp+2, 0.7)
		}
	}

	g.dessinerConsigne(dst)
	g.dessinerJournal(dst)
	g.dessinerChronique(dst)
}

// ---------------------------------------------------------------- consigne

func (g *Game) consigne() (string, string) {
	p := g.p
	e := p.E
	moi := e.Exp[p.Humain]
	if e.Fini {
		switch {
		case e.Gagnant == p.Humain:
			return "Fin de la partie", "Vous avez gagné : le trésor d'El Dorado est en lieu sûr."
		case e.Gagnant >= 0:
			return "Fin de la partie", fmt.Sprintf("L'expédition %s a gagné la partie.", e.Exp[e.Gagnant].Nom)
		}
		return "Fin de la partie", "Personne n'a rapporté le trésor."
	}
	switch p.Req.Type {
	case ReqDeplacer:
		var b strings.Builder
		if moi.Tresor {
			b.WriteString("Rapportez le trésor dans l'une des six villes ! ")
		}
		if moi.Jangada {
			b.WriteString("Cliquez une case en surbrillance : bleu clair en jangada (3 cases en descendant un fleuve, 2 en le remontant, 2 le long de la côte ; arrêt aux cascades et dans les marais), orange à pied en abandonnant la jangada.")
		} else {
			b.WriteString("Cliquez une case dorée pour avancer d'une case à pied (le marais est infranchissable à pied).")
		}
		b.WriteString(" Cliquez votre propre case pour rester sur place.")
		if moi.Kits > 0 && !moi.Jangada {
			if jeu.Navigable(moi.Pos) {
				fmt.Fprintf(&b, " Construire une jangada demande %d tours sur cette case.", jeu.ToursConstruction)
			} else {
				b.WriteString(" On ne construit une jangada que sur une case de fleuve ou de marais.")
			}
		}
		return "Déplacement", b.String()
	case ReqAnnoncer:
		c, _ := jeu.Centrale(moi.Pos)
		f := "sans fleuve"
		if c.Fleuve {
			f = "avec fleuve"
		}
		est := "La feuille ne permet pas encore de désigner une case."
		if e.FleuveCentral >= 0 && e.NatureCentrale != jeu.NatInconnue {
			if (c.Fleuve == (e.FleuveCentral == 1)) && c.Nature == e.NatureCentrale {
				est = fmt.Sprintf("Fleuve et terrain concordent ; zone %s probable à %d %% d'après la feuille.", c.Zone, int(e.Croyance(p.Humain)[c.Zone]*100+.5))
			} else {
				est = "Fleuve ou terrain ne concordent pas : ce n'est pas ici."
			}
		}
		return "Annoncer El Dorado ?", fmt.Sprintf("Vous êtes dans la région centrale : zone %s, %s, %s. %s Si c'est la cité d'or, vous prenez le trésor ; sinon votre expédition, découragée, ne jouera pas pendant deux tours. Les codes secrets seront dévoilés.", c.Zone, c.Nature, f, est)
	case ReqContact:
		var b strings.Builder
		fmt.Fprintf(&b, "Une tribu de %d natifs ! Prenez contact (dé d'attitude : 1-2 hostiles, 3-4 neutres, 5-6 amicaux) ou fuyez vers la case occupée au tour précédent.", p.Req.Natifs)
		if !moi.Specialiste() {
			b.WriteString(" Sans médecin, jésuite ni guide : −1 au dé.")
		}
		if moi.Cadeaux > 0 {
			b.WriteString(" Chaque cadeau offert ajoute 1.")
		}
		fmt.Fprintf(&b, " En cas de combat, rapport des forces : %+d.", jeu.RapportForces(moi.Membres[jeu.Arquebusier], p.Req.Natifs))
		return "Rencontre", b.String()
	case ReqVictime:
		return "Un homme meurt", "Choisissez qui, dans votre expédition, a péri. Les arquebusiers pèsent dans les combats contre les natifs ; le médecin, le guide et le jésuite évitent chacun un type de retard, et sans aucun des trois les natifs sont moins bien disposés (−1)."
	case ReqCombat:
		adv := e.Exp[p.Req.Adv]
		return "Combat entre expéditions", fmt.Sprintf("%s compte %d hommes, vous %d. À chaque passe, un dé chacun : le plus fort tue un homme adverse, l'égalité coûte un homme à l'attaquant. Poursuivre, ou fuir vers votre case précédente ?", adv.Nom, adv.Hommes(), moi.Hommes())
	}
	if e.Courant != p.Humain {
		return "Patience", fmt.Sprintf("L'expédition %s joue. Suivez ses aventures dans la chronique. (A : accélérer)", e.Exp[e.Courant].Nom)
	}
	if moi.Retard > 0 {
		return "Retard", "Votre expédition passe son tour."
	}
	return "Votre tour", "Les dés roulent…"
}

func (g *Game) dessinerConsigne(dst *ebiten.Image) {
	carte(dst, pnX, yCons, pnW, hCons)
	titre, texte := g.consigne()
	titreCarte(dst, titre, pnX, yCons+4, pnW)
	paragraphe(dst, texte, pnX+18, yCons+40, pnW-36, face(17, policeTexte), coulEncre, 20.5)
	if g.p.Req.Type == ReqContact {
		moi := g.p.E.Exp[g.p.Humain]
		if moi.Cadeaux > 0 {
			y := yCons + hCons - 104
			ecrire(dst, fmt.Sprintf("Cadeaux offerts : %d", g.cadeaux), pnX+18, y+6, face(18, policeGras), coulEncre)
			ecrire(dst, fmt.Sprintf("(sur %d)", moi.Cadeaux), pnX+210, y+7, face(16, policeItalique), coulSepia)
		}
	}
}

func (g *Game) boutonsJeu() {
	p := g.p
	e := p.E
	moi := e.Exp[p.Humain]
	yb := yCons + hCons - 56
	bw := (pnW - 46) / 2
	switch p.Req.Type {
	case ReqDeplacer:
		g.bouton(pnX+16, yb, bw, 42, "Rester sur place", true, func() {
			p.repondre(cmdCoup{jeu.Coup{Vers: moi.Pos, Mode: jeu.Rester}})
		})
		for _, c := range p.Req.Coups {
			if c.Mode == jeu.Construire {
				c := c
				b := g.bouton(pnX+30+bw, yb, bw, 42, fmt.Sprintf("Construire (%d/%d)", moi.Construction, jeu.ToursConstruction), true, func() {
					p.repondre(cmdCoup{c})
				})
				b.Style = BoutonOr
			}
		}
	case ReqAnnoncer:
		b := g.bouton(pnX+16, yb, bw, 42, "« El Dorado ! »", true, func() { p.repondre(cmdOui{}) })
		b.Style = BoutonOr
		g.bouton(pnX+30+bw, yb, bw, 42, "Ne rien annoncer", true, func() { p.repondre(cmdNon{}) })
	case ReqContact:
		if moi.Cadeaux > 0 {
			y := yCons + hCons - 102
			g.bouton(pnX+pnW-110, y, 40, 34, "−", g.cadeaux > 0, func() { g.cadeaux-- }).Taille = 24
			g.bouton(pnX+pnW-62, y, 40, 34, "+", g.cadeaux < moi.Cadeaux, func() { g.cadeaux++ }).Taille = 24
		}
		b := g.bouton(pnX+16, yb, bw, 42, "Prendre contact", true, func() { p.repondre(cmdContact{g.cadeaux}) })
		b.Style = BoutonOr
		b2 := g.bouton(pnX+30+bw, yb, bw, 42, "Fuir", moi.Prec != moi.Pos, func() { p.repondre(cmdNon{}) })
		b2.Style = BoutonDanger
	case ReqVictime:
		var ms []jeu.Membre
		for m := range jeu.NbMembres {
			if moi.Membres[m] > 0 {
				ms = append(ms, m)
			}
		}
		// largeurs à proportion des noms : « Arquebusier » ne rapetisse pas
		noms := make([]string, len(ms))
		tot := 0.0
		for i, m := range ms {
			noms[i] = strings.ToUpper(m.String()[:1]) + m.String()[1:]
			tot += largeur(noms[i], face(18, policeGras)) + 16
		}
		k := (pnW - 32 - 8*float64(len(ms)-1)) / tot
		x := pnX + 16
		for i, m := range ms {
			m := m
			w := (largeur(noms[i], face(18, policeGras)) + 16) * k
			b := g.bouton(x, yb-8, w, 50, fmt.Sprintf("%s\n(%d)", noms[i], moi.Membres[m]), true, func() {
				p.repondre(cmdMembre{m})
			})
			b.Taille = 18
			b.Style = BoutonDanger
			x += w + 8
		}
	case ReqCombat:
		b := g.bouton(pnX+16, yb, bw, 42, "Poursuivre", true, func() { p.repondre(cmdOui{}) })
		b.Style = BoutonDanger
		g.bouton(pnX+30+bw, yb, bw, 42, "Fuir", true, func() { p.repondre(cmdNon{}) })
	}
	// barre du bas
	y := 948.0
	w := (pnW - 16) / 3
	g.bouton(pnX, y, w, 40, "Règle", true, func() { g.regles = true }).Touche = "R"
	g.bouton(pnX+w+8, y, w, 40, "Feuille", true, func() { g.feuille = true }).Touche = "F"
	b := g.bouton(pnX+2*w+16, y, w, 40, "Accélérer", true, func() { p.Rapide = !p.Rapide })
	b.Touche = "A"
	b.Choisi = p.Rapide
	for _, b := range g.boutons[len(g.boutons)-3:] {
		b.Style = BoutonDiscret
		b.Taille = 21
	}
	g.boutonsZoom()
}

// ---------------------------------------------------------------- journal de route

func (g *Game) dessinerJournal(dst *ebiten.Image) {
	e := g.p.E
	moi := e.Exp[g.p.Humain]
	carte(dst, pnX, yJour, pnW, hJour)
	titreCarte(dst, "Journal de route", pnX, yJour+4, pnW)
	dessinerImage(dst, imgExpeditions[moi.Couleur], pnX+pnW-56, yJour+40, 38, 38, nil)
	cadre(dst, pnX+pnW-56, yJour+40, 38, 38, 1.5, coulOrSombre)
	fn := face(17, policeTexte)
	fg := face(17, policeGras)
	x0, y := pnX+18, yJour+40
	if !moi.Vivante {
		ecrire(dst, "Votre expédition a péri.", x0, y, face(20, policeGras), coulSang)
		return
	}
	col := []float64{x0, x0 + 172}
	for m := range jeu.NbMembres {
		x := col[int(m)%2]
		yy := y + float64(int(m)/2)*22
		ecrire(dst, fmt.Sprint(moi.Membres[m]), x, yy, fg, coulEncre)
		nom := m.Pluriel()
		if moi.Membres[m] == 1 {
			nom = m.String()
		}
		ecrire(dst, nom, x+24, yy, fn, coulEncre)
	}
	y += 46
	ecrire(dst, fmt.Sprint(moi.Cadeaux), x0, y, fg, coulEncre)
	ecrire(dst, "cadeau"+map[bool]string{true: "x"}[moi.Cadeaux > 1], x0+24, y, fn, coulEncre)
	ecrire(dst, fmt.Sprintf("Code secret : %d", moi.Code), col[1], y, fg, coulSang)
	y += 22
	s := "À pied"
	if moi.Jangada {
		s = "En jangada"
	}
	if moi.Kits > 0 {
		s += fmt.Sprintf(", matériel pour %d jangada%s", moi.Kits, pl(moi.Kits))
		if moi.Construction > 0 {
			s += fmt.Sprintf(" (%d/%d)", moi.Construction, jeu.ToursConstruction)
		}
	}
	ecrire(dst, s, x0, y, fn, coulEncre)
	y += 22
	switch {
	case moi.Tresor:
		ecrire(dst, "Vous portez le trésor d'El Dorado !", x0, y, fg, coulOrSombre)
	case moi.Retard > 0:
		ecrire(dst, fmt.Sprintf("Retard : %d tour%s à passer", moi.Retard, pl(moi.Retard)), x0, y, fg, coulSang)
	default:
		ecrire(dst, g.resumeRenseignements(), x0, y, face(16, policeItalique), coulSepia)
	}
	y += 22
	ecrire(dst, g.resumeCentral(), x0, y, face(16, policeItalique), coulSepia)
}

// resumeRenseignements : fiabilités cumulées de vos renseignements par zone.
func (g *Game) resumeRenseignements() string {
	e := g.p.E
	var s [4]int
	for _, r := range e.Feuille {
		if r.Joueur == g.p.Humain && r.Type == jeu.RensZone {
			s[r.Zone] += r.Fiab
		}
	}
	return fmt.Sprintf("Vos indices : Ouest %d %%, Est %d %%, Sud %d %%", s[jeu.Ouest], s[jeu.Est], s[jeu.Sud])
}

func (g *Game) resumeCentral() string {
	e := g.p.E
	f := "?"
	if e.FleuveCentral == 1 {
		f = "oui"
	} else if e.FleuveCentral == 0 {
		f = "non"
	}
	return fmt.Sprintf("Région centrale : fleuve %s, terrain %s", f, e.NatureCentrale)
}

// ---------------------------------------------------------------- dés et chronique

func (g *Game) dessinerChronique(dst *ebiten.Image) {
	e := g.p.E
	carte(dst, pnX, yChro, pnW, hChro)
	titreCarte(dst, "Chronique", pnX, yChro+4, pnW)
	// derniers dés
	d := e.Des
	if len(d.Valeurs) > 0 {
		x := pnX + pnW - 18
		for i := len(d.Valeurs) - 1; i >= 0; i-- {
			x -= 40
			fond := coulParch
			if d.Joueur == g.p.Humain {
				fond = coulOr2
			}
			de(dst, x, yChro+40, 34, d.Valeurs[i], fond, coulEncre)
		}
		lib := d.Sens
		if d.Sens == "jet du tour" && len(d.Valeurs) == 2 {
			lib = fmt.Sprintf("jet du tour : %d%d", d.Valeurs[0], d.Valeurs[1])
		}
		nom := e.Exp[d.Joueur].Nom
		if d.Joueur == g.p.Humain {
			nom = "Vous"
		}
		ecrire(dst, nom, pnX+18, yChro+38, face(17, policeGras), coulEncre)
		ecrire(dst, lib, pnX+18, yChro+57, face(16, policeItalique), coulSepia)
	}
	// journal, du plus récent (en bas) au plus ancien
	y0 := yChro + 88
	ligne(dst, pnX+14, y0-4, pnX+pnW-14, y0-4, 1, coulParch3)
	fn := face(15.5, policeTexte)
	inter := 18.0
	yMax := yChro + hChro - 12
	type bloc struct {
		j      int
		lignes []string
		tour   int
	}
	var blocs []bloc
	h := 0.0
	fin := len(e.Journal) - g.defilLog
	if fin < 1 {
		fin = min(1, len(e.Journal))
		g.defilLog = len(e.Journal) - fin
	}
	for i := fin - 1; i >= 0 && h < yMax-y0; i-- {
		en := e.Journal[i]
		l := couper(en.Texte, fn, pnW-48)
		blocs = append(blocs, bloc{en.Joueur, l, en.Tour})
		h += float64(len(l))*inter + 3
	}
	y := yMax
	for _, b := range blocs {
		hb := float64(len(b.lignes))*inter + 3
		if y-hb < y0 {
			break
		}
		y -= hb
		c := coulSepia
		if b.j >= 0 {
			c = coulExpeditions[e.Exp[b.j].Couleur]
		}
		rect(dst, pnX+16, y+3, 6, hb-6, c)
		col := coulEncre
		if b.j == g.p.Humain {
			col = coulJungle
		}
		for i, s := range b.lignes {
			ecrire(dst, s, pnX+30, y+float64(i)*inter, fn, col)
		}
	}
	if g.defilLog > 0 {
		ecrireDroite(dst, "molette : plus récent ↓", pnX+pnW-16, yChro+hChro-26, face(13, policeItalique), coulSepia)
	}
}

// ---------------------------------------------------------------- fin

func (g *Game) dessinerFin(dst *ebiten.Image) {
	e := g.p.E
	rect(dst, 0, 0, LargPlat, HautPlat, avecAlpha(coulCuir, 0x90))
	w, h := 640.0, 300.0
	x, y := LargPlat/2-w/2, 330.0
	carte(dst, x, y, w, h)
	titre, texte := "La jungle garde son secret", "Aucune expédition n'a rapporté le trésor."
	switch {
	case e.Gagnant == g.p.Humain:
		titre = "Victoire !"
		texte = fmt.Sprintf("Votre expédition rapporte le trésor d'El Dorado à %s. Gloire et fortune vous attendent à la cour d'Espagne.", jeu.Villes[jeu.VilleEn(e.Exp[e.Gagnant].Pos)].Nom)
	case e.Gagnant >= 0:
		titre = "Défaite"
		texte = fmt.Sprintf("L'expédition %s rapporte le trésor d'El Dorado à %s.", e.Exp[e.Gagnant].Nom, jeu.Villes[jeu.VilleEn(e.Exp[e.Gagnant].Pos)].Nom)
	case !e.Exp[g.p.Humain].Vivante && e.Vivantes() == 0:
		texte = "Toutes les expéditions ont péri dans la forêt."
	}
	ecrireCentre(dst, titre, x+w/2, y+34, face(32, policeTitre), coulSang)
	losange(dst, x+w/2, y+92, 5, coulOrSombre)
	ligne(dst, x+w/2-120, y+92, x+w/2-12, y+92, 1, coulOrSombre)
	ligne(dst, x+w/2+12, y+92, x+w/2+120, y+92, 1, coulOrSombre)
	paragraphe(dst, texte, x+40, y+110, w-80, face(21, policeTexte), coulEncre, 26)
	ecrireCentre(dst, fmt.Sprintf("Partie terminée au tour %d.", e.Tour), x+w/2, y+200, face(17, policeItalique), coulSepia)
}
