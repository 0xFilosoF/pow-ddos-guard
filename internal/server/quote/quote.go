package quote

import "math/rand/v2"

//nolint:gochecknoglobals // okay for example
var quotesBook = []string{
	"Silence often teaches what words cannot",
	"What you practice daily becomes your destiny",
	"Patience is strength in slow motion",
	"A calm mind sees clearer paths",
	"Discipline creates the freedom you seek",
	"Listen more than you speak; wisdom lives there",
	"Small choices shape great lives",
	"Clarity comes from honesty with yourself",
	"Growth begins where comfort ends",
	"Guard your time; it is your life in hours",
	"Kindness costs little, but pays forever",
	"Master your thoughts, or they will master you",
	"Consistency outperforms intensity",
	"Peace is built, not found",
	"The way you start matters less than the way you persist",
	"Simplicity reveals what truly matters",
	"Your habits write your future silently",
	"Truth requires courage, not volume",
	"Learn to pause before you react",
	"Wisdom grows when ego shrinks",
}

// Random Word of Wisdom, here we go.
func Random() string {
	//nolint:gosec // simple rand okay for example
	return quotesBook[rand.IntN(len(quotesBook))]
}
