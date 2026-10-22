package main

import (
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// La règle consultable en jeu (touche R) : la règle de l'encart (signée
// Strate et Pop), résumée fidèlement, et les choix faits là où elle est muette.
// Format : « # » titre, « - » puce, « | » tableau, « > » exemple.

var pagesRegle = []struct{ onglet, texte string }{
	{"En bref", `# Le but du jeu
Une cité d'or, El Dorado, se cache dans l'une des 18 cases de la grande région centrale de la carte (bordée de dents noires). Votre expédition doit découvrir laquelle, s'y rendre, prendre le trésor, et le rapporter dans l'une des six villes de la côte. La première expédition qui y parvient gagne.
# Personne ne sait où est la cité
Au début, même le jeu l'ignore : l'emplacement se décide peu à peu, grâce aux renseignements que donnent les natifs. Il faut réunir trois indices :
- la zone : Ouest, Est ou Sud (renseignements hors de la région centrale, plus ou moins fiables) ;
- y a-t-il un fleuve ? (premier renseignement donné dans la région centrale, toujours sûr) ;
- le terrain : jungle, marais, ou plaine et plateau (second renseignement dans la région centrale, toujours sûr).
Zone, fleuve et terrain ne désignent qu'une seule des 18 cases.
===
# Une partie, en cinq temps
- 1. Organiser l'expédition : 1 000 sols d'or pour recruter des hommes et acheter cadeaux et radeaux ; choisir une ville de départ.
- 2. Marcher vers le centre. Chaque tour, les dés décident : un événement (souvent un mort ou un retard) ou une rencontre avec une tribu.
- 3. Gagner l'amitié des tribus (ou les faire prisonnières) : elles donnent des renseignements. Plus vous êtes près du centre, plus ils sont fiables.
- 4. Entrer dans la région centrale pour les deux renseignements sûrs, puis se placer sur la bonne case et annoncer « El Dorado ! ».
- 5. Revenir dans une ville avec le trésor, en survivant aux natifs et aux expéditions rivales qui veulent vous le prendre.
> Le danger est permanent : un jet sur quatre tue un homme. Une expédition sans homme est perdue. Il suffit d'un survivant pour rapporter le trésor.`},
	{"L'expédition", `# Recruter et acheter
| Achat | Prix | À quoi ça sert
| Arquebusier | 80 | gagner les combats contre les natifs
| Médecin | 80 | éviter les retards dus aux maladies
| Jésuite | 80 | éviter les retards dus au moral
| Guide | 80 | éviter les retards dus à l'égarement
| Cadeau | 20 | +1 au dé d'attitude des natifs
| Jangada | 160 | radeau pour les fleuves
Sans médecin, ni jésuite, ni guide, les natifs sont moins bien disposés (−1 au dé d'attitude).
# Le départ
Chaque expédition part d'une ville différente : Quito, Lima, Belem, Salvador de Bahia, Rio de Janeiro ou Santa Maria de los Buenos Aires. On part à pied, ou en jangada si la ville est sur un fleuve ou sur la côte.
===
# Le code secret
Chaque joueur choisit en secret un chiffre de 1 à 6. Il sert à calculer la fiabilité de ses propres renseignements (page « Indices »). Comme les autres ne le connaissent pas, ils ne savent pas si vos renseignements sont fiables : c'est tout le sel du jeu.
# Le journal de route
L'ordinateur tient pour vous le journal de route (panneau de droite) : vos hommes, vos cadeaux, votre jangada, votre code secret, et le cumul de vos indices.
> Conseil : 7 arquebusiers, un médecin, un jésuite, un guide et 10 cadeaux font une expédition équilibrée.`},
	{"Un tour", `# Chaque tour, chaque expédition
- 1. Se déplace d'une case (ou plus en jangada), ou reste sur place.
- 2. Lance deux dés : le premier donne les dizaines, le second les unités (de 11 à 66).
- 3. Nombre pair : un événement (tableau 1, ci-contre).
- 4. Nombre impair : une tribu d'autant de natifs (53 = 53 natifs) : voir « Natifs ».
# Les retards
Un retard fait passer un tour entier. Le bon spécialiste l'évite : le médecin pour les maladies, le guide pour l'égarement, le jésuite pour le moral.
# Les morts
Quand un homme meurt, vous choisissez lequel. Gardez si possible au moins un spécialiste.
===
# Tableau 1 : les événements
| Dés | Événement | Conséquence
| 12 à 36 | fièvre de l'or, piranhas, boa… | 1 mort
| 42 à 46 | malaria, eau croupie… | retard sauf si médecin
| 52 à 56 | boussole déréglée, mirages… | retard sauf si guide
| 62 à 66 | superstition, angoisse… | retard sauf si jésuite
> Exemple : les dés donnent 4 puis 4 : 44, « fruits empoisonnés ». Avec un médecin, rien ne se passe ; sans médecin, l'expédition perd son prochain tour.`},
	{"Se déplacer", `# À pied
Une case par tour, sur tous les terrains sauf le marais, infranchissable à pied.
# En jangada (radeau)
Seulement sur les cases de fleuve ou de marais, et le long de la côte.
- en descendant un fleuve : 3 cases par tour ; en le remontant : 2 ; le long de la côte : 2 ;
- on s'arrête sur une cascade (barres blanches sur le fleuve), en entrant dans un marais, et sur la case d'une autre expédition ;
- pour passer d'un fleuve à un autre, on s'arrête à leur confluent et l'on repart au tour suivant.
# Abandonner, construire
On peut abandonner sa jangada pour continuer à pied : elle est perdue. Avec du matériel (une jangada achetée en plus), on en construit une en restant 4 tours sur une case de fleuve ou de marais.
===
# À l'écran
- cases dorées : un pas à pied ;
- cases bleu clair : en jangada ; survolez-les pour voir le trajet ;
- cases orange : à pied, en abandonnant la jangada ;
- cliquez votre propre case (cerclée de blanc) pour rester sur place ;
- l'infobulle d'une case donne son terrain, son fleuve, sa distance au centre, ses natifs ;
- pion-natif cerclé de turquoise : tribu amie ; voilé : natifs cachés de vous ; barré : tribu exterminée ;
- dans la région centrale, les cases écartées par les indices sûrs s'assombrissent.
# Les touches
- R : la règle ; F : la feuille commune des renseignements ;
- A : accélérer l'ordinateur ; molette sur la chronique : la remonter ;
- F11 : plein écran.`},
	{"Natifs", `# Une tribu en vue
Vous choisissez : prendre contact, ou fuir vers la case d'où vous venez (le tour s'arrête là). Au contact, offrez des cadeaux (+1 chacun), puis un dé donne l'attitude (tableau 3) :
| Dé | Attitude | Conséquence
| 1, 2 | hostile | ils attaquent : combat
| 3, 4 | neutre | ils se cachent
| 5, 6 | amicale | ils vous renseignent
La tribu laisse un pion-natif sur la case. Une tribu amie le reste toute la partie, mais ne renseigne qu'une fois chaque joueur.
# Le combat
On lance un dé, auquel on ajoute le rapport des forces (tableau 2, ci-contre), puis on lit le tableau 4 :
| Dé modifié | Résultat
| 1, 2, 3 | 3, 2 ou 1 mort chez vous ; la tribu est exterminée
| 4, 5 | les natifs s'enfuient
| 6 | prisonniers : ils vous renseignent
===
# Tableau 2 : rapport des forces
| Natifs | 1 | 3 | 5 | 7 | 9 | 11 | 13 | 15 | 17 | 19 | 21 | 23 | 25
| 10 | −1 | 0 | 0 | 1 | 2 | 3 | 4 | 5 | 5 | 5 | 5 | 5 | 5
| 20 | −2 | −1 | 0 | 0 | 1 | 2 | 3 | 4 | 5 | 5 | 5 | 5 | 5
| 30 | −3 | −2 | −1 | 0 | 0 | 1 | 2 | 3 | 4 | 5 | 5 | 5 | 5
| 40 | −4 | −3 | −2 | −1 | 0 | 0 | 1 | 2 | 3 | 4 | 5 | 5 | 5
| 50 | −5 | −4 | −3 | −2 | −1 | 0 | 0 | 1 | 2 | 3 | 4 | 5 | 5
| 60 | −5 | −5 | −4 | −3 | −2 | −1 | 0 | 0 | 1 | 2 | 3 | 4 | 5
En colonnes, vos arquebusiers ; en rangées, les natifs (par dizaines).
> Exemple : 7 arquebusiers contre 53 natifs : −2. Un dé de 5 donne 3 : un mort chez vous, la tribu est exterminée.
Le panneau affiche ce rapport avant que vous décidiez de prendre contact ou de fuir.`},
	{"Indices", `# D'où viennent les indices
Une tribu amie, ou faite prisonnière, vous donne un renseignement. Hors de la région centrale, il porte sur la zone (tableau 5) : un dé, 1-2 Ouest (côté Lima), 3-4 Est (côté Belem), 5-6 Sud (côté Buenos Aires).
# Est-il fiable ?
Un second dé, plus votre code secret (au-delà de 6, on retire 6), croisé avec votre distance au centre (tableau 6), donne sa fiabilité : de 9 % loin du centre à 100 % tout près.
| Dé + code | 3 cases | 5 | 7 | 9 | 11
| 1 | 33 % | 20 | 14 | 11 | 9
| 3 | 100 % | 60 | 42 | 33 | 27
| 6 | 100 % | 100 | 85 | 66 | 54
> Exemple : à 5 cases du centre, la tribu dit « Est ». Dé 2 + code 3 = 5 : 100 %.
===
# Qui décide de la zone ?
Les fiabilités d'un même joueur pour une même zone s'additionnent (20 % + 33 % + 50 % = 103 %). Le premier joueur qui atteint 100 % fixe la bonne zone, pour tout le monde.
Tout est noté sur la feuille commune (touche F) : qui, quelle zone, à quelle distance, avec quel dé. Mais chacun ne connaît que son propre code secret : vous savez où vous en êtes, pas les autres. Le jeu vous donne une estimation de chaque zone d'après la feuille.
# Dans la région centrale
Les renseignements y sont sûrs et il n'y en a que deux pour toute la partie : le premier dit s'il y a un fleuve (dé 1-3 non, 4-6 oui), le second donne le terrain (1-2 jungle, 3-4 marais, 5-6 plaine ou plateau). Ils servent à tous.`},
	{"El Dorado", `# Annoncer
Sur l'une des 18 cases, vous pouvez annoncer « El Dorado ! ». Les codes secrets sont alors dévoilés et la feuille dit la vérité :
- si c'est la bonne case, vous prenez le trésor ;
- sinon, votre expédition, découragée, ne joue pas pendant deux tours… et tout le monde connaît désormais la bonne case.
N'annoncez donc que si les trois indices concordent.
# Le retour
Rapportez le trésor dans n'importe quelle ville. En chemin, les natifs restent dangereux. Si votre expédition est massacrée, le trésor reste sur place : une autre expédition peut le récupérer en obtenant l'amitié des natifs ou en les faisant prisonniers.
===
# Les combats entre expéditions
Deux expéditions sur la même case se battent : celle qui arrive attaque. Chacune lance un dé ; la plus forte tue un homme de l'autre ; en cas d'égalité, l'attaquant perd un homme. On continue jusqu'à ce qu'une expédition soit exterminée, sauf si l'une fuit vers sa case précédente. Médecins, jésuites et guides se battent aussi.
Le vainqueur prend le trésor du vaincu. Sur la case d'El Dorado, celle qui fuit laisse le trésor à l'autre.
# Gagner
La première expédition qui rapporte le trésor dans une ville gagne la partie. Si toutes périssent, la jungle garde son secret.`},
	{"Adaptation", `# Choix de l'adaptation
La règle laisse quelques points dans l'ombre ; voici ce que fait le jeu.
- Région centrale : la zone Est n'a pas de marais bien net sur la carte ; ses deux cases d'un turquoise plus bleu (12,11 et 13,12) en tiennent lieu, pour que chaque zone ait ses six combinaisons. 11,14, longée par le fleuve, compte comme case à fleuve. L'infobulle d'une case candidate rappelle sa description.
- Tableau 2 : on lit la colonne du plus grand nombre d'arquebusiers imprimé qui ne dépasse pas le sien, et la rangée des dizaines de natifs. Sans arquebusier : colonne 1, et encore −1.
- Tableau 6 : au-delà de 11 cases de distance, on lit la colonne 11.
- Une tribu exterminée laisse un pion-natif barré : plus de rencontre sur cette case (la règle dit qu'on note « qu'il n'y a plus de natifs »).
- Une tribu ne donne qu'un renseignement par joueur, qu'elle soit amie ou prisonnière.
- Jangada : tout départ d'une ville côtière peut suivre la côte (2 cases par tour), comme à Bahia ; elle se construit sur une case de fleuve ou de marais ; un tour de retard passé sur place compte pour la construction.
- On annonce « El Dorado » juste après s'être déplacé (ou être resté) sur l'une des 18 cases. Si les renseignements ne désignent encore aucune case, l'annonce est manquée.
- Combat entre expéditions : celle qui n'a pas bougé ce tour-ci ne peut pas fuir ; le vainqueur prend le trésor du vaincu ; une fuite ne déclenche pas de nouveau combat.
- Un retard ou un découragement : l'expédition ne fait rien de son tour.`},
}

const (
	rgX = 120.0
	rgY = 60.0
	rgW = float64(LargeurEcran - 240)
	rgH = float64(HauteurEcran - 120)
)

func (g *Game) touchesRegle() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.regles = false
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		g.pageRgl = (g.pageRgl + 1) % len(pagesRegle)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		g.pageRgl = (g.pageRgl + len(pagesRegle) - 1) % len(pagesRegle)
	}
}

func (g *Game) boutonsRegles() {
	n := len(pagesRegle)
	w := (rgW - 60) / float64(n)
	for i, pg := range pagesRegle {
		b := g.bouton(rgX+30+float64(i)*w, rgY+96, w-8, 36, pg.onglet, true, func() { g.pageRgl = i })
		b.Choisi = i == g.pageRgl
		b.Taille = 19
	}
	b := g.bouton(rgX+rgW-150, rgY+rgH-58, 120, 40, "Fermer", true, func() { g.regles = false })
	b.Style = BoutonOr
}

func (g *Game) dessinerRegles(dst *ebiten.Image) {
	rect(dst, 0, 0, LargeurEcran, HauteurEcran, avecAlpha(coulCuir, 0xc0))
	frise(dst, rgX-10, rgY-34, rgW+20, 26)
	carte(dst, rgX, rgY, rgW, rgH)
	ecrire(dst, "El Dorado", rgX+30, rgY+20, face(40, policeTitre), coulSang)
	ecrire(dst, "la règle du jeu en encart, signée Strate et Pop", rgX+330, rgY+40, face(22, policeItalique), coulSepia)
	ecrireDroite(dst, "flèches ← → : pages · Échap : fermer", rgX+rgW-30, rgY+44, face(16, policeTexte), coulSepia)
	x0, y0 := rgX+40, rgY+156
	colW := (rgW - 120) / 2
	yMax := rgY + rgH - 70
	x, y := x0, y0
	ft := face(27, policeGothique)
	fn := face(18, policeTexte)
	fi := face(17, policeItalique)
	saut := func(h float64) {
		if y+h > yMax && x == x0 {
			x, y = x0+colW+40, y0
		}
	}
	entete := true
	lignes := strings.Split(pagesRegle[g.pageRgl].texte, "\n")
	g.lignesRgl = lignes
	for li, l := range lignes {
		switch {
		case l == "===":
			x, y = x0+colW+40, y0
		case strings.HasPrefix(l, "# "):
			if y > y0 {
				y += 8
			}
			saut(64)
			ecrire(dst, strings.TrimPrefix(l, "# "), x, y, ft, coulSang)
			y += 36
			entete = true
		case strings.HasPrefix(l, "| "):
			cells := strings.Split(strings.TrimPrefix(l, "| "), " | ")
			saut(26)
			if entete {
				rect(dst, x, y-1, colW, 25, coulJungle)
			} else {
				ligne(dst, x, y+23, x+colW, y+23, 1, coulParch3)
			}
			pos := g.colonnesTableau(li, colW)
			for i, c := range cells {
				f, col := fn, coulEncre
				if entete {
					f, col = face(16, policeGras), coulParch
				}
				ecrire(dst, c, x+8+colW*pos[i], y+1, f, col)
			}
			y += 26
			entete = false
		case strings.HasPrefix(l, "- "):
			lignes := couper(strings.TrimPrefix(l, "- "), fn, colW-22)
			saut(float64(len(lignes)) * 23)
			losange(dst, x+6, y+12, 4, coulOrSombre)
			for _, s := range lignes {
				ecrire(dst, s, x+18, y, fn, coulEncre)
				y += 23
			}
			y += 3
		case strings.HasPrefix(l, "> "):
			lignes := couper(strings.TrimPrefix(l, "> "), fi, colW-30)
			h := float64(len(lignes))*22 + 12
			saut(h)
			rect(dst, x, y, colW, h, coulParch2)
			rect(dst, x, y, 5, h, coulOr)
			for i, s := range lignes {
				ecrire(dst, s, x+16, y+6+float64(i)*22, fi, coulEncre)
			}
			y += h + 8
		default:
			lignes := couper(l, fn, colW)
			saut(float64(len(lignes)) * 23)
			for _, s := range lignes {
				ecrire(dst, s, x, y, fn, coulEncre)
				y += 23
			}
			y += 4
			entete = true
		}
	}
}

// colonnesTableau : position (fraction de la colonne de texte) de chaque
// colonne du tableau qui contient la ligne li, d'après sa cellule la plus
// large ; l'espace restant est réparti également.
func (g *Game) colonnesTableau(li int, colW float64) []float64 {
	l := g.lignesRgl
	debut, fin := li, li
	for debut > 0 && strings.HasPrefix(l[debut-1], "| ") {
		debut--
	}
	for fin+1 < len(l) && strings.HasPrefix(l[fin+1], "| ") {
		fin++
	}
	var larg []float64
	for _, t := range l[debut : fin+1] {
		for i, c := range strings.Split(strings.TrimPrefix(t, "| "), " | ") {
			if i >= len(larg) {
				larg = append(larg, 0)
			}
			larg[i] = max(larg[i], largeur(c, face(18, policeTexte))+16)
		}
	}
	tot := 0.0
	for _, w := range larg {
		tot += w
	}
	marge := max(colW-16-tot, 0) / float64(len(larg))
	pos := make([]float64, len(larg))
	x := 0.0
	for i, w := range larg {
		pos[i] = x / colW
		x += w + marge
	}
	return pos
}
