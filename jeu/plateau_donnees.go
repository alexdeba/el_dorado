// Code généré par outils/decoupe ; NE PAS MODIFIER.

package jeu

// Grille des cases sur l'image du plateau redressé (assets/gen/plateau.jpg).
const (
	LargImage, HautImage = 2500, 2120
	GrilleX0, GrilleY0   = 18, 100
	GrilleDX, GrilleDY   = 95.025, 82.21
)

// Terrain de chaque case, rangée par rangée (les impaires décalées d'une
// demi-case vers la droite) : m mer, a marais, j jungle, p plaine, t plateau, n montagne.
var rangeesPlateau = [NbRang]string{
	"mmmmnjjjppppjjmmmmmmmmmmmmm", // 0
	"mmmmjjjpjppjjjjjmmmmmmmmmmm", // 1
	"mmmmntjpjjjjjjjjjpmmmmmmmmm", // 2
	"mmmjnjjjjjppjjpjjmmmmmmmmmm", // 3
	"mmmtnjjjjjjpjpjjjpmmmmmmmmm", // 4
	"mmpnjjjajjjjjaaaajjmmmmmmmm", // 5
	"mmmnjjjaaajjaaaaaajjjmmmmmm", // 6
	"mmtnjjjajjjjjjaajajjjttmmmm", // 7
	"mmtntjjjjjjjjjjjjtjjtttttmm", // 8
	"mmmnnjjajjjjjjttjtjtttttpmm", // 9
	"mmmmtjtjaajjjjttjtjttttttmm", // 10
	"mmmmnttjjapjjttttjtjttjjmmm", // 11
	"mmmmmnnnjjapnjtttttjttjtmmm", // 12
	"mmmmmnnnjpajjapnpnjtjjtmmmm", // 13
	"mmmmmmmnnjpnpaapjnptptpmmmm", // 14
	"mmmmmmmmtnnjpatnpjpjtjmmmmm", // 15
	"mmmmmmmmnnnjpaapjjajjjmmmmm", // 16
	"mmmmmmmmnnjppatjjtttnmmmmmm", // 17
	"mmmmmmmmnnjppppptttnpmmmmmm", // 18
	"mmmmmmmnnnppppptttmmmmmmmmm", // 19
	"mmmmmmmmnnppppppptmmmmmmmmm", // 20
	"mmmmmmmnnppppapppmmmmmmmmmm", // 21
	"mmmmmmmmnppppppppmmmmmmmmmm", // 22
	"mmmmmmmpnppppppmmmmmmmmmmmm", // 23
	"mmmmmmmmnppppppmmmmmmmmmmmm", // 24
}

// Tronçons de fleuve entre deux cases voisines, de l'amont vers l'aval,
// avec le numéro du fleuve (on ne change de fleuve qu'en s'arrêtant).
var tronconsPlateau = []Troncon{
	{Case{9, 5}, Case{10, 6}, 1},
	{Case{13, 5}, Case{14, 6}, 1},
	{Case{14, 5}, Case{15, 5}, 1},
	{Case{15, 5}, Case{16, 5}, 1},
	{Case{6, 6}, Case{7, 6}, 1},
	{Case{7, 6}, Case{7, 7}, 1},
	{Case{8, 6}, Case{9, 6}, 1},
	{Case{9, 6}, Case{9, 5}, 1},
	{Case{10, 6}, Case{11, 6}, 1},
	{Case{11, 6}, Case{12, 6}, 1},
	{Case{12, 6}, Case{13, 6}, 1},
	{Case{13, 6}, Case{13, 5}, 1},
	{Case{14, 6}, Case{14, 5}, 1},
	{Case{3, 7}, Case{4, 7}, 1},
	{Case{4, 7}, Case{5, 7}, 1},
	{Case{5, 7}, Case{6, 6}, 1},
	{Case{7, 7}, Case{8, 6}, 1},
	{Case{3, 8}, Case{3, 7}, 1},
	{Case{3, 9}, Case{3, 8}, 1},
	{Case{4, 10}, Case{3, 9}, 1},
	{Case{5, 8}, Case{5, 7}, 2},
	{Case{5, 9}, Case{5, 8}, 2},
	{Case{5, 10}, Case{5, 9}, 2},
	{Case{5, 11}, Case{5, 10}, 2},
	{Case{6, 12}, Case{5, 11}, 2},
	{Case{4, 5}, Case{5, 5}, 3},
	{Case{5, 5}, Case{6, 6}, 3},
	{Case{7, 8}, Case{7, 7}, 4},
	{Case{4, 4}, Case{5, 4}, 5},
	{Case{5, 4}, Case{6, 4}, 5},
	{Case{6, 4}, Case{7, 4}, 5},
	{Case{7, 4}, Case{7, 5}, 5},
	{Case{8, 4}, Case{8, 5}, 5},
	{Case{7, 5}, Case{8, 4}, 5},
	{Case{8, 5}, Case{9, 5}, 5},
	{Case{6, 3}, Case{7, 3}, 6},
	{Case{7, 3}, Case{8, 4}, 6},
	{Case{8, 3}, Case{7, 3}, 7},
	{Case{10, 5}, Case{10, 6}, 8},
	{Case{9, 2}, Case{9, 3}, 9},
	{Case{9, 3}, Case{9, 4}, 9},
	{Case{9, 4}, Case{10, 4}, 9},
	{Case{10, 4}, Case{11, 4}, 9},
	{Case{11, 4}, Case{11, 5}, 9},
	{Case{11, 5}, Case{12, 5}, 9},
	{Case{12, 5}, Case{12, 6}, 9},
	{Case{10, 2}, Case{11, 2}, 10},
	{Case{11, 2}, Case{12, 2}, 10},
	{Case{12, 2}, Case{12, 3}, 10},
	{Case{12, 3}, Case{12, 4}, 10},
	{Case{12, 4}, Case{11, 5}, 10},
	{Case{10, 7}, Case{11, 7}, 11},
	{Case{11, 7}, Case{12, 6}, 11},
	{Case{9, 8}, Case{9, 9}, 11},
	{Case{10, 8}, Case{10, 7}, 11},
	{Case{6, 9}, Case{7, 9}, 11},
	{Case{7, 9}, Case{8, 9}, 11},
	{Case{8, 9}, Case{9, 8}, 11},
	{Case{9, 9}, Case{10, 8}, 11},
	{Case{7, 10}, Case{7, 9}, 12},
	{Case{12, 7}, Case{13, 6}, 13},
	{Case{11, 8}, Case{12, 8}, 13},
	{Case{12, 8}, Case{12, 7}, 13},
	{Case{10, 9}, Case{11, 8}, 13},
	{Case{10, 10}, Case{10, 9}, 13},
	{Case{9, 11}, Case{10, 10}, 13},
	{Case{10, 11}, Case{9, 11}, 13},
	{Case{11, 11}, Case{10, 11}, 13},
	{Case{10, 12}, Case{9, 11}, 14},
	{Case{9, 13}, Case{10, 12}, 14},
	{Case{14, 3}, Case{14, 4}, 15},
	{Case{14, 4}, Case{14, 5}, 15},
	{Case{15, 6}, Case{15, 5}, 16},
	{Case{14, 7}, Case{15, 6}, 16},
	{Case{14, 8}, Case{14, 7}, 16},
	{Case{13, 9}, Case{14, 8}, 16},
	{Case{13, 10}, Case{13, 9}, 16},
	{Case{12, 11}, Case{13, 10}, 16},
	{Case{13, 11}, Case{12, 11}, 16},
	{Case{14, 9}, Case{14, 8}, 17},
	{Case{15, 10}, Case{14, 9}, 17},
	{Case{16, 6}, Case{16, 5}, 18},
	{Case{15, 7}, Case{16, 7}, 18},
	{Case{16, 7}, Case{16, 6}, 18},
	{Case{16, 8}, Case{15, 7}, 18},
	{Case{16, 9}, Case{16, 8}, 18},
	{Case{16, 10}, Case{16, 9}, 18},
	{Case{15, 11}, Case{16, 10}, 18},
	{Case{15, 12}, Case{15, 11}, 18},
	{Case{17, 6}, Case{16, 6}, 19},
	{Case{18, 6}, Case{18, 5}, 20},
	{Case{17, 7}, Case{18, 6}, 20},
	{Case{18, 7}, Case{17, 7}, 20},
	{Case{18, 8}, Case{18, 7}, 20},
	{Case{18, 9}, Case{18, 8}, 20},
	{Case{18, 10}, Case{18, 9}, 20},
	{Case{17, 11}, Case{18, 10}, 20},
	{Case{17, 12}, Case{17, 11}, 20},
	{Case{14, 13}, Case{15, 13}, 20},
	{Case{15, 13}, Case{16, 13}, 20},
	{Case{16, 13}, Case{17, 12}, 20},
	{Case{15, 14}, Case{14, 13}, 20},
	{Case{16, 14}, Case{16, 13}, 21},
	{Case{19, 10}, Case{18, 9}, 22},
	{Case{18, 11}, Case{19, 10}, 22},
	{Case{18, 12}, Case{18, 11}, 22},
	{Case{19, 8}, Case{18, 7}, 23},
	{Case{23, 9}, Case{24, 10}, 24},
	{Case{21, 10}, Case{22, 10}, 24},
	{Case{22, 10}, Case{23, 10}, 24},
	{Case{23, 10}, Case{23, 9}, 24},
	{Case{20, 11}, Case{21, 10}, 24},
	{Case{20, 12}, Case{20, 11}, 24},
	{Case{20, 13}, Case{20, 12}, 24},
	{Case{20, 14}, Case{20, 13}, 24},
	{Case{19, 15}, Case{20, 14}, 24},
	{Case{20, 16}, Case{19, 15}, 24},
	{Case{20, 10}, Case{20, 11}, 25},
	{Case{7, 13}, Case{8, 14}, 26},
	{Case{8, 14}, Case{8, 15}, 26},
	{Case{10, 14}, Case{10, 15}, 27},
	{Case{9, 15}, Case{10, 16}, 27},
	{Case{10, 15}, Case{9, 15}, 27},
	{Case{10, 16}, Case{11, 16}, 27},
	{Case{11, 16}, Case{12, 16}, 27},
	{Case{12, 16}, Case{12, 17}, 27},
	{Case{12, 17}, Case{13, 18}, 27},
	{Case{13, 18}, Case{13, 19}, 27},
	{Case{13, 19}, Case{13, 20}, 27},
	{Case{13, 20}, Case{12, 21}, 27},
	{Case{12, 21}, Case{13, 22}, 27},
	{Case{12, 22}, Case{12, 23}, 27},
	{Case{13, 22}, Case{12, 22}, 27},
	{Case{12, 13}, Case{13, 13}, 28},
	{Case{13, 13}, Case{14, 14}, 28},
	{Case{14, 14}, Case{13, 15}, 28},
	{Case{13, 15}, Case{14, 16}, 28},
	{Case{14, 16}, Case{13, 17}, 28},
	{Case{13, 17}, Case{14, 18}, 28},
	{Case{14, 18}, Case{13, 19}, 28},
	{Case{14, 15}, Case{13, 15}, 29},
	{Case{10, 17}, Case{11, 18}, 30},
	{Case{10, 18}, Case{10, 17}, 30},
	{Case{11, 18}, Case{12, 18}, 30},
	{Case{12, 18}, Case{12, 19}, 30},
	{Case{12, 19}, Case{13, 20}, 30},
	{Case{16, 16}, Case{15, 17}, 31},
	{Case{17, 16}, Case{16, 16}, 31},
	{Case{18, 16}, Case{17, 16}, 31},
	{Case{15, 17}, Case{15, 18}, 31},
	{Case{17, 17}, Case{18, 16}, 31},
	{Case{18, 17}, Case{17, 17}, 31},
	{Case{15, 18}, Case{14, 19}, 31},
	{Case{18, 18}, Case{18, 17}, 31},
	{Case{14, 19}, Case{15, 20}, 31},
	{Case{14, 20}, Case{14, 21}, 31},
	{Case{15, 20}, Case{14, 20}, 31},
	{Case{14, 21}, Case{14, 22}, 31},
	{Case{14, 22}, Case{13, 22}, 31},
	{Case{17, 15}, Case{17, 16}, 32},
	{Case{15, 15}, Case{16, 16}, 33},
	{Case{16, 18}, Case{15, 18}, 34},
	{Case{16, 20}, Case{15, 20}, 35},
	{Case{9, 19}, Case{10, 19}, 36},
	{Case{10, 19}, Case{11, 20}, 36},
	{Case{11, 20}, Case{11, 21}, 36},
	{Case{11, 21}, Case{12, 22}, 36},
	{Case{9, 2}, Case{8, 3}, 37},
	{Case{4, 4}, Case{4, 5}, 38},
	{Case{8, 4}, Case{9, 4}, 39},
	{Case{9, 5}, Case{10, 5}, 40},
	{Case{7, 8}, Case{7, 9}, 41},
	{Case{18, 9}, Case{19, 8}, 42},
	{Case{9, 13}, Case{10, 14}, 43},
	{Case{15, 14}, Case{15, 15}, 44},
}

var cascadesPlateau = []Case{
	{10, 10},
	{14, 8},
	{16, 18},
	{17, 17},
	{18, 17},
}

var villesPlateau = []Ville{
	{"Quito", Case{4, 4}},
	{"Lima", Case{4, 10}},
	{"Belem", Case{18, 5}},
	{"Salvador de Bahia", Case{23, 12}},
	{"Rio de Janeiro", Case{20, 18}},
	{"Santa Maria de los Buenos Aires", Case{12, 23}},
}
