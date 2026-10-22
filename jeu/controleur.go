package jeu

// Controleur : les décisions d'un joueur, demandées par le moteur au fil du
// tour (voir l'organigramme de la règle).
type Controleur interface {
	Deplacer(e *Etat, j int, coups []Coup) Coup     // déplacement (ou rester, construire)
	Annoncer(e *Etat, j int) bool                   // sur une case de la région centrale : « El Dorado » ?
	Contact(e *Etat, j int, natifs int) (bool, int) // prendre contact ? avec combien de cadeaux
	Victime(e *Etat, j int) Membre                  // qui meurt
	Combattre(e *Etat, j, adversaire int) bool      // poursuivre le combat (sinon fuir)
}
