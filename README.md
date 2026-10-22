# El Dorado

Adaptation en Go (Ebitengine) du jeu « El Dorado », paru en encart dans un
magazine et signé *Strate et Pop*, d'après sa règle et la photo de l'encart.
Vous menez une expédition de conquistadors contre une à cinq expéditions dirigées par l'ordinateur (Orellana, Quesada, Belalcázar,
Federmann, Aguirre, Hutten) : découvrir où se cache la cité d'or au cœur du
Mato Grosso, puis en rapporter le trésor dans une ville.

## Compiler

Il faut [Go](https://go.dev/dl/) 1.27.1 ou plus récent. Le plateau, les pions et les polices sont
intégrés au binaire : il se suffit à lui-même.

```
git clone https://github.com/alexdeba/el_dorado.git
cd el_dorado
go build -o eldorado .
./eldorado                        # F11 : plein écran
```

`go run .` lance le jeu sans garder de binaire.

Selon le système :

- **Linux** : Ebitengine a besoin d'un compilateur C et des en-têtes OpenGL
  et X11. Sous Debian ou Ubuntu :
  ```
  sudo apt install gcc pkg-config libgl1-mesa-dev libxcursor-dev libxi-dev \
      libxinerama-dev libxrandr-dev libxxf86vm-dev
  ```
- **macOS** : les outils en ligne de commande de Xcode
  (`xcode-select --install`).
- **Windows** : rien de plus. `go build -o eldorado.exe .` donne le
  binaire. On peut aussi le produire depuis Linux ou macOS :
  `GOOS=windows go build -o eldorado.exe .`

## Déroulement

1. **Accueil** : la couverture (`assets/cover.jpg`) ; choisissez le nombre
   d'expéditions rivales.
2. **Organisation** : 1 000 sols d'or pour les arquebusiers, médecins,
   jésuites, guides (80), cadeaux (20) et jangadas (160) ; cliquez la ville de
   départ sur la carte, choisissez votre code secret et votre pion.
3. **Le jeu** : à chaque tour, déplacez-vous (cases en surbrillance), puis les
   dés décident : événement (nombre pair) ou rencontre d'une tribu (impair).
   Le panneau de droite donne la consigne et les choix de chaque étape
   (déplacement, contact ou fuite et cadeaux, victime, combat, annonce
   d'El Dorado), le journal de route de votre expédition, les derniers dés et
   la chronique de la partie.

Touches : R la règle (8 pages, flèches ← → ; la première, « En bref », résume le but et le déroulement d'une partie), F la feuille commune des
renseignements, A accélérer l'ordinateur, molette sur la chronique, F11 plein
écran.

Zoom : barre d'espace (zoom ×2,5 autour du curseur, ou de votre pion si le
curseur est hors du plateau ; un second appui revient à la carte entière),
molette sur le plateau (autour du curseur), touches + et −, 0 ou
« Tout voir » pour revenir à la carte entière, boutons en haut à gauche du
plateau. On déplace la vue en glissant avec le bouton droit (ou du milieu) ou
avec les flèches. Dès que l'on zoome, une mini-carte apparaît en bas à gauche :
la vue y est encadrée, les expéditions (ou les villes, pendant
l'organisation) y sont pointées ; un clic ou un glisser y déplace la vue.

## Organisation du code

- `jeu/` : le moteur de règles, sans interface.
  - `plateau.go` : cases hexagonales (pointe en haut, rangées impaires décalées),
    voisins, distances, fleuves orientés, cascades, villes ; la région centrale
    (18 cases candidates décrites par zone, fleuve et terrain).
  - `plateau_donnees.go` : terrains, tronçons de fleuve, cascades, villes et
    grille, relevés sur la photo de l'encart (voir « Le plateau »).
  - `tables.go` : les six tableaux de la carte.
  - `etat.go`, `tour.go` : expéditions, pions-natifs, feuille commune ; le tour
    suit l'organigramme de la règle (déplacement, événement ou rencontre,
    attitude, combat, renseignement, annonce, retour).
  - `deplacement.go` : à pied, en jangada (fleuves, marais, côte), construction.
  - `verite.go` : la zone d'El Dorado se fixe en cours de partie ; `Croyance`
    énumère les codes secrets possibles des autres pour estimer chaque zone.
  - `ia.go` : l'ordinateur.
  - `controleur.go` : l'interface d'un joueur, que le moteur appelle.
- `partie.go` : le moteur joue dans sa goroutine ; le joueur humain répond par
  un canal (requêtes `ReqDeplacer`, `ReqContact`, `ReqVictime`, `ReqCombat`,
  `ReqAnnoncer`).
- Interface : `vue.go` (plateau, coups, pions, infobulles), `camera.go`
  (zoom, cadrage, mini-carte), `panneau.go`,
  `organisation.go`, `accueil.go`, `regles.go`, `feuille.go`,
  `ressources.go` (palette, polices, cartouches, boutons, dés).
- Habillage : le cuir et le parchemin des cartes de conquistadors, la frise en
  grecque de la bordure de la carte (découpée sur la photo) en bandeau, l'or
  des Incas ; titres Cinzel Decorative, intitulés gothiques Pirata One, texte
  EB Garamond (OFL, `assets/polices`). Les pions sont ceux de l'encart,
  découpés sur la photo.

## Le plateau

La carte affichée est la photo de l'encart, redressée et rendue plus nette
(`assets/gen/plateau.jpg`). Les cases (hexagones, 25 rangées de 27), les
terrains, les fleuves orientés vers l'aval, les cascades et les villes y ont
été relevés par un outil qui n'est pas distribué ; le résultat est dans
`jeu/plateau_donnees.go`.

## Choix d'adaptation

La règle est appliquée telle quelle ; là où elle est muette, ambiguë ou
contredite par la carte :

- **Région centrale** : la règle annonce, par zone, six cases (avec ou sans
  fleuve × jungle, marais, plaine ou plateau). Sur la carte, la zone Est n'a
  pas de marais bien net : ses deux cases d'un turquoise plus bleu (12,11 et
  13,12) en tiennent lieu ; 11,14, longée par le fleuve, compte comme case à
  fleuve. Le découpage Ouest / Est / Sud suit les villes citées par le
  tableau 5 (Lima, Belem, Buenos Aires). L'infobulle de chaque case candidate
  donne sa description, pour que le joueur raisonne comme sur le plateau.
- **Tableau 2** : on lit la colonne du plus grand nombre d'arquebusiers
  imprimé qui ne dépasse pas le sien, la rangée des dizaines de natifs
  (11 à 19 : rangée 10 ; 60 à 66 : rangée 60). Sans arquebusier : première
  colonne et encore −1. Le dé modifié est borné à 1..6.
- **Tableau 6** : au-delà de 11 cases de la case centrale, on lit la
  colonne 11 ; en deçà de 3, la colonne 3.
- **Vérité** : la bonne zone est celle du premier joueur dont les fiabilités
  réelles cumulées sur une même zone atteignent 100 %, comme le dit la règle.
  Les deux renseignements de la région centrale sont sûrs et tirés une seule
  fois pour toute la partie.
- **Tribus** : une tribu exterminée laisse un pion-natif barré (plus de
  rencontre). Une tribu ne donne qu'un renseignement par joueur, qu'elle soit
  amie ou prisonnière ; une tribu amie accueille ensuite sans dé. Natifs
  enfuis ou prisonniers sont absents tant que l'expédition reste sur la case.
- **Jangada** : la règle ne cite que Bahia pour la côte ; tout départ côtier
  peut la suivre (2 cases par tour). Une jangada se construit sur une case de
  fleuve ou de marais ; un tour de retard passé sur place compte pour la
  construction. Le matériel acheté en plus sert à en construire d'autres.
  Rio n'est reliée à aucun fleuve sur la carte : on en part à pied ou par la côte.
- **Fleuves** : on change de fleuve en s'arrêtant (la règle l'impose aux
  confluents) ; de Quito partent deux fleuves, dont l'un ne compte qu'une case
  avant le confluent.
- **Annonce** : on annonce « El Dorado » juste après son déplacement, sur
  l'une des 18 cases. Si les renseignements ne désignent encore aucune case,
  l'annonce est manquée (deux tours de découragement, codes dévoilés).
- **Combats entre expéditions** : celle qui n'a pas bougé ce tour-ci ne peut
  pas fuir (elle n'a pas de case précédente) ; le vainqueur prend le trésor du
  vaincu ; la fuite ne déclenche pas de nouveau combat. Les victimes sont
  choisies par chaque joueur, comme pour les natifs.
- **Retards** : un retard ou le découragement fait sauter tout le tour
  (déplacement compris).
- Une fois le trésor trouvé, les tribus ne donnent plus de renseignement.
- Garde-fou : la partie s'arrête au tour 400 (jamais atteint en pratique).

## L'ordinateur

Chaque expédition achète 7 ou 8 arquebusiers, un médecin, un jésuite, un guide
et des cadeaux (parfois une jangada depuis une ville sur un fleuve), part de la
ville libre la plus proche du centre, et planifie ses trajets tour par tour
(recherche en largeur sur les états case × jangada). Elle cherche des tribus à
3 ou 4 cases de la case centrale (renseignements les plus fiables), entre dans
la région centrale pour les deux renseignements sûrs, annonce El Dorado quand
sa propre estimation (`Croyance`) dépasse un seuil qui dépend de son audace,
puis rapporte le trésor dans la ville la plus proche. Elle fuit les grosses
tribus quand elle est affaiblie, et n'attaque une expédition rivale qu'en
nette supériorité (ou pour lui prendre le trésor). Sur 40 parties entre
ordinateurs, environ 40 % voient le trésor rapporté ; dans les autres, toutes
les expéditions périssent : la règle est meurtrière (un mort sur un jet sur
quatre).

## Crédits et licence

*El Dorado* est un jeu paru en encart dans un magazine, signé *Strate et
Pop*. Le jeu original, sa règle et ses illustrations appartiennent à leurs
auteurs. Cette adaptation est un hommage non commercial.

Les scans du magazine ne sont pas distribués dans ce dépôt. Seules les
images dont le jeu a besoin le sont (`assets/gen`).

Le code est sous licence MIT (voir `LICENSE`). Les polices de
`assets/polices` (Cinzel Decorative, Pirata One, EB Garamond) gardent leur
licence SIL OFL (`assets/polices/OFL.txt`).
