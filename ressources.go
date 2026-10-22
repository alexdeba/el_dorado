package main

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

//go:embed assets/gen assets/polices
var fichiers embed.FS

var (
	imgPlateau     *ebiten.Image // la carte redressée (2500 × 2120)
	imgCouv        *ebiten.Image
	imgFrise       *ebiten.Image
	imgNatif       *ebiten.Image
	imgExpeditions [6]*ebiten.Image

	policeTitre, policeGothique, policeTexte, policeGras, policeItalique *Police
)

// Palette : cuir de selle, parchemin des cartes espagnoles, or des Incas,
// vert de la forêt, rouge du sang des combats, turquoise des fleuves.
var (
	coulCuir      = color.RGBA{0x26, 0x18, 0x0f, 0xff}
	coulCuir2     = color.RGBA{0x3b, 0x26, 0x17, 0xff}
	coulCuir3     = color.RGBA{0x55, 0x38, 0x22, 0xff}
	coulParch     = color.RGBA{0xf1, 0xe3, 0xc2, 0xff}
	coulParch2    = color.RGBA{0xe3, 0xcf, 0xa2, 0xff}
	coulParch3    = color.RGBA{0xc8, 0xab, 0x74, 0xff}
	coulEncre     = color.RGBA{0x2b, 0x1d, 0x14, 0xff}
	coulSepia     = color.RGBA{0x75, 0x55, 0x33, 0xff}
	coulOr        = color.RGBA{0xd9, 0xa4, 0x1e, 0xff}
	coulOr2       = color.RGBA{0xf6, 0xd8, 0x6e, 0xff}
	coulOrSombre  = color.RGBA{0x8c, 0x63, 0x12, 0xff}
	coulJungle    = color.RGBA{0x1d, 0x45, 0x33, 0xff}
	coulJungle2   = color.RGBA{0x2c, 0x63, 0x48, 0xff}
	coulSang      = color.RGBA{0xa3, 0x2a, 0x1b, 0xff}
	coulTurquoise = color.RGBA{0x2b, 0xb3, 0xb0, 0xff}
	coulOrange    = color.RGBA{0xff, 0x8a, 0x1e, 0xff}
	coulEau       = color.RGBA{0x9c, 0xee, 0xff, 0xff} // coups en jangada
	coulBlanc     = color.RGBA{0xff, 0xff, 0xff, 0xff}
	coulNoir      = color.RGBA{0x00, 0x00, 0x00, 0xff}
)

// couleurs des six pions-expéditions de l'encart
var (
	coulExpeditions = [6]color.RGBA{
		{0x22, 0xa8, 0xde, 0xff}, {0xf3, 0xd5, 0x1c, 0xff}, {0xe0, 0x4a, 0x1c, 0xff},
		{0x1d, 0x8c, 0x4a, 0xff}, {0x4a, 0x2a, 0x8a, 0xff}, {0xb9, 0xa4, 0x2a, 0xff},
	}
	nomsCouleurs = [6]string{"Azur", "Or", "Pourpre", "Émeraude", "Violet", "Olive"}
)

func chargerImage(nom string) *ebiten.Image {
	b, err := fichiers.ReadFile(nom)
	if err != nil {
		log.Fatal(err)
	}
	img, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		log.Fatal(nom, err)
	}
	return ebiten.NewImageFromImage(img)
}

// Police : une police (variable ou non) et sa graisse (0 : celle par défaut).
type Police struct {
	src   *text.GoTextFaceSource
	poids float32
}

var sources = map[string]*text.GoTextFaceSource{}

func chargerPolice(nom string, poids float32) *Police {
	if s, ok := sources[nom]; ok {
		return &Police{s, poids}
	}
	b, err := fichiers.ReadFile("assets/polices/" + nom)
	if err != nil {
		log.Fatal(err)
	}
	s, err := text.NewGoTextFaceSource(bytes.NewReader(b))
	if err != nil {
		log.Fatal(err)
	}
	sources[nom] = s
	return &Police{s, poids}
}

func chargerRessources() {
	imgPlateau = chargerImage("assets/gen/plateau.jpg")
	imgCouv = chargerImage("assets/gen/couverture.jpg")
	imgFrise = chargerImage("assets/gen/frise.png")
	imgNatif = chargerImage("assets/gen/natif.png")
	for i := range imgExpeditions {
		imgExpeditions[i] = chargerImage(fmt.Sprintf("assets/gen/expedition%d.png", i))
	}
	policeTitre = chargerPolice("CinzelDecorative-Bold.ttf", 0)
	policeGothique = chargerPolice("PirataOne-Regular.ttf", 0)
	policeTexte = chargerPolice("EBGaramond.ttf", 500)
	policeGras = chargerPolice("EBGaramond.ttf", 700)
	policeItalique = chargerPolice("EBGaramond-Italique.ttf", 500)
}

// ---------------------------------------------------------------- texte

var tagGraisse = text.MustParseTag("wght")

type cleFace struct {
	p *Police
	t float64
}

var faces = map[cleFace]*text.GoTextFace{}

func face(taille float64, p *Police) *text.GoTextFace {
	k := cleFace{p, taille}
	if f, ok := faces[k]; ok {
		return f
	}
	f := &text.GoTextFace{Source: p.src, Size: taille}
	if p.poids > 0 {
		f.SetVariation(tagGraisse, p.poids)
	}
	faces[k] = f
	return f
}

func ecrire(dst *ebiten.Image, s string, x, y float64, f text.Face, c color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(c)
	text.Draw(dst, s, f, op)
}

func ecrireCentre(dst *ebiten.Image, s string, cx, y float64, f text.Face, c color.Color) {
	ecrire(dst, s, cx-text.Advance(s, f)/2, y, f, c)
}

func ecrireDroite(dst *ebiten.Image, s string, x, y float64, f text.Face, c color.Color) {
	ecrire(dst, s, x-text.Advance(s, f), y, f, c)
}

// ecrireHalo : texte cerné.
func ecrireHalo(dst *ebiten.Image, s string, x, y float64, f text.Face, c, halo color.Color, ep float64) {
	for _, d := range [][2]float64{{-1, 0}, {1, 0}, {0, -1}, {0, 1}, {-.7, -.7}, {.7, .7}, {-.7, .7}, {.7, -.7}} {
		ecrire(dst, s, x+d[0]*ep, y+d[1]*ep, f, halo)
	}
	ecrire(dst, s, x, y, f, c)
}

func largeur(s string, f text.Face) float64 { return text.Advance(s, f) }

// mots : les mots d'un paragraphe ; la ponctuation haute reste avec le mot
// qui précède, et « n° 12 » reste d'un bloc.
func mots(s string) []string {
	var l []string
	for _, m := range strings.Fields(s) {
		if n := len(l); n > 0 && (m == ";" || m == ":" || m == "!" || m == "?" || m == "»" || m == "%" || l[n-1] == "«" || l[n-1] == "n°") {
			l[n-1] += " " + m
			continue
		}
		l = append(l, m)
	}
	return l
}

// couper : un texte en lignes de largeur maximale w.
func couper(s string, f text.Face, w float64) []string {
	var lignes []string
	for _, para := range strings.Split(s, "\n") {
		cur := ""
		for _, mot := range mots(para) {
			essai := mot
			if cur != "" {
				essai = cur + " " + mot
			}
			if text.Advance(essai, f) > w && cur != "" {
				lignes = append(lignes, cur)
				cur = mot
			} else {
				cur = essai
			}
		}
		lignes = append(lignes, cur)
	}
	return lignes
}

// paragraphe : écrit un texte coupé, renvoie la hauteur utilisée.
func paragraphe(dst *ebiten.Image, s string, x, y, w float64, f text.Face, c color.Color, inter float64) float64 {
	l := couper(s, f, w)
	for i, t := range l {
		ecrire(dst, t, x, y+float64(i)*inter, f, c)
	}
	return float64(len(l)) * inter
}

// ---------------------------------------------------------------- formes

func rect(dst *ebiten.Image, x, y, w, h float64, c color.Color) {
	vector.FillRect(dst, float32(x), float32(y), float32(w), float32(h), c, false)
}

func cadre(dst *ebiten.Image, x, y, w, h, ep float64, c color.Color) {
	vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), float32(ep), c, false)
}

func disque(dst *ebiten.Image, x, y, r float64, c color.Color) {
	vector.FillCircle(dst, float32(x), float32(y), float32(r), c, true)
}

func cercle(dst *ebiten.Image, x, y, r, ep float64, c color.Color) {
	vector.StrokeCircle(dst, float32(x), float32(y), float32(r), float32(ep), c, true)
}

func ligne(dst *ebiten.Image, x0, y0, x1, y1, ep float64, c color.Color) {
	vector.StrokeLine(dst, float32(x0), float32(y0), float32(x1), float32(y1), float32(ep), c, true)
}

func avecAlpha(c color.RGBA, a uint8) color.RGBA {
	f := float64(a) / 255
	return color.RGBA{uint8(float64(c.R) * f), uint8(float64(c.G) * f), uint8(float64(c.B) * f), a}
}

func melange(a, b color.RGBA, t float64) color.RGBA {
	m := func(x, y uint8) uint8 { return uint8(float64(x)*(1-t) + float64(y)*t + .5) }
	return color.RGBA{m(a.R, b.R), m(a.G, b.G), m(a.B, b.B), m(a.A, b.A)}
}

// dessinerImage : img mise à l'échelle dans le rectangle (x, y, w, h).
func dessinerImage(dst, img *ebiten.Image, x, y, w, h float64, cs *ebiten.ColorScale) {
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	op.GeoM.Scale(w/float64(img.Bounds().Dx()), h/float64(img.Bounds().Dy()))
	op.GeoM.Translate(x, y)
	if cs != nil {
		op.ColorScale = *cs
	}
	dst.DrawImage(img, op)
}

// polygone plein (anticrénelé).
func polygone(dst *ebiten.Image, pts [][2]float64, c color.Color) {
	var p vector.Path
	for i, q := range pts {
		if i == 0 {
			p.MoveTo(float32(q[0]), float32(q[1]))
		} else {
			p.LineTo(float32(q[0]), float32(q[1]))
		}
	}
	p.Close()
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(c)
	vector.FillPath(dst, &p, nil, op)
}

func contour(dst *ebiten.Image, pts [][2]float64, ep float64, c color.Color) {
	var p vector.Path
	for i, q := range pts {
		if i == 0 {
			p.MoveTo(float32(q[0]), float32(q[1]))
		} else {
			p.LineTo(float32(q[0]), float32(q[1]))
		}
	}
	p.Close()
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(c)
	vector.StrokePath(dst, &p, &vector.StrokeOptions{Width: float32(ep), LineJoin: vector.LineJoinRound}, op)
}

// ---------------------------------------------------------------- ornements

// frise : le motif en grecque de la bordure de la carte, répété.
func frise(dst *ebiten.Image, x, y, w, h float64) {
	b := imgFrise.Bounds()
	tw := h * float64(b.Dx()) / float64(b.Dy())
	sub := dst.SubImage(image.Rect(int(x), int(y), int(x+w), int(y+h))).(*ebiten.Image)
	for px := x; px < x+w; px += tw {
		dessinerImage(sub, imgFrise, px, y, tw, h, nil)
	}
	ligne(dst, x, y, x+w, y, 1.5, coulOrSombre)
	ligne(dst, x, y+h, x+w, y+h, 1.5, coulOrSombre)
}

// losange : petit ornement d'angle.
func losange(dst *ebiten.Image, x, y, r float64, c color.Color) {
	polygone(dst, [][2]float64{{x, y - r}, {x + r, y}, {x, y + r}, {x - r, y}}, c)
}

// carte : un cartouche de parchemin bordé d'or, coins ornés.
func carte(dst *ebiten.Image, x, y, w, h float64) {
	rect(dst, x+3, y+4, w, h, avecAlpha(coulNoir, 0x60))
	rect(dst, x, y, w, h, coulParch)
	// grain : bandes légères
	for i := 0.0; i < h; i += 7 {
		rect(dst, x, y+i, w, 2, avecAlpha(coulParch2, 0x40))
	}
	cadre(dst, x+1, y+1, w-2, h-2, 2, coulOrSombre)
	cadre(dst, x+5, y+5, w-10, h-10, 1, coulOr)
	for _, c := range [][2]float64{{x + 5, y + 5}, {x + w - 5, y + 5}, {x + 5, y + h - 5}, {x + w - 5, y + h - 5}} {
		losange(dst, c[0], c[1], 5, coulOrSombre)
		losange(dst, c[0], c[1], 2.5, coulOr2)
	}
}

// titreCarte : intitulé gothique d'un cartouche, avec filet.
func titreCarte(dst *ebiten.Image, s string, x, y, w float64) {
	f := face(24, policeGothique)
	ecrire(dst, s, x+16, y+6, f, coulSang)
	lw := largeur(s, f)
	ligne(dst, x+24+lw, y+22, x+w-16, y+22, 1, coulParch3)
	losange(dst, x+w-16, y+22, 3, coulOr)
}

// fondCuir : panneau de cuir, dégradé et coutures.
func fondCuir(dst *ebiten.Image, x, y, w, h float64) {
	n := 24
	for i := range n {
		t := float64(i) / float64(n-1)
		rect(dst, x, y+h*float64(i)/float64(n), w, h/float64(n)+1, melange(coulCuir2, coulCuir, t))
	}
	for py := y + 8; py < y+h-8; py += 14 {
		ligne(dst, x+6, py, x+6, py+7, 1, avecAlpha(coulOrSombre, 0x90))
		ligne(dst, x+w-6, py, x+w-6, py+7, 1, avecAlpha(coulOrSombre, 0x90))
	}
}

// ---------------------------------------------------------------- dés

func de(dst *ebiten.Image, x, y, t float64, v int, fond, point color.RGBA) {
	rect(dst, x+2, y+3, t, t, avecAlpha(coulNoir, 0x70))
	polygone(dst, [][2]float64{{x + 6, y}, {x + t - 6, y}, {x + t, y + 6}, {x + t, y + t - 6}, {x + t - 6, y + t}, {x + 6, y + t}, {x, y + t - 6}, {x, y + 6}}, fond)
	r := t * 0.09
	c := t / 2
	q := t * 0.26
	pos := map[int][][2]float64{
		1: {{c, c}},
		2: {{q, q}, {t - q, t - q}},
		3: {{q, q}, {c, c}, {t - q, t - q}},
		4: {{q, q}, {t - q, q}, {q, t - q}, {t - q, t - q}},
		5: {{q, q}, {t - q, q}, {c, c}, {q, t - q}, {t - q, t - q}},
		6: {{q, q}, {t - q, q}, {q, c}, {t - q, c}, {q, t - q}, {t - q, t - q}},
	}[v]
	for _, p := range pos {
		disque(dst, x+p[0], y+p[1], r, point)
	}
}

// ---------------------------------------------------------------- boutons

type StyleBouton int

const (
	BoutonNormal StyleBouton = iota
	BoutonOr                 // action principale
	BoutonDanger
	BoutonDiscret
)

type Bouton struct {
	X, Y, W, H float64
	Texte      string
	Actif      bool
	Choisi     bool
	Style      StyleBouton
	Taille     float64
	Touche     string // raccourci clavier, rappelé en petit dans le coin
	Action     func()
}

func (b *Bouton) Contient(x, y int) bool {
	fx, fy := float64(x), float64(y)
	return fx >= b.X && fx < b.X+b.W && fy >= b.Y && fy < b.Y+b.H
}

func dessinerBouton(dst *ebiten.Image, b *Bouton, survol bool) {
	fond, bord, txt := coulJungle, coulOr, coulParch
	switch b.Style {
	case BoutonOr:
		fond, bord, txt = coulOr, coulOrSombre, coulEncre
	case BoutonDanger:
		fond, bord, txt = coulSang, coulOr, coulParch
	case BoutonDiscret:
		fond, bord, txt = coulParch2, coulOrSombre, coulEncre
	}
	if b.Choisi {
		fond, bord, txt = coulOr2, coulSang, coulEncre
	}
	if !b.Actif {
		fond = melange(fond, coulParch3, 0.6)
		txt = melange(txt, coulSepia, 0.6)
		bord = coulSepia
	} else if survol {
		fond = melange(fond, coulBlanc, 0.18)
	}
	x, y, w, h := b.X, b.Y, b.W, b.H
	ch := 6.0
	pts := [][2]float64{{x + ch, y}, {x + w - ch, y}, {x + w, y + ch}, {x + w, y + h - ch}, {x + w - ch, y + h}, {x + ch, y + h}, {x, y + h - ch}, {x, y + ch}}
	ombre := make([][2]float64, len(pts))
	for i, p := range pts {
		ombre[i] = [2]float64{p[0] + 2, p[1] + 3}
	}
	polygone(dst, ombre, avecAlpha(coulNoir, 0x80))
	polygone(dst, pts, fond)
	contour(dst, pts, 2, bord)
	// liseré intérieur
	ligne(dst, x+ch+2, y+3, x+w-ch-2, y+3, 1, avecAlpha(coulBlanc, 0x50))
	t := b.Taille
	if t == 0 {
		t = 22
	}
	lignes := strings.Split(b.Texte, "\n")
	marge := 14.0
	if b.Touche != "" {
		marge = 34
		ecrireDroite(dst, b.Touche, x+w-9, y+h-19, face(13, policeItalique), melange(txt, fond, 0.4))
	}
	// le texte rapetisse s'il déborde
	for t > 10 {
		lw := 0.0
		for _, l := range lignes {
			lw = max(lw, largeur(l, face(t, policeGras)))
		}
		if lw <= w-marge {
			break
		}
		t -= 0.5
	}
	f := face(t, policeGras)
	y0 := y + h/2 - float64(len(lignes))*t*0.62
	for i, l := range lignes {
		ecrireCentre(dst, l, x+w/2, y0+float64(i)*t*1.2, f, txt)
	}
}
