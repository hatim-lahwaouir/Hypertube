package types

var ImdbGenres map[string]uint64 = map[string]uint64{
	"Action":      1,
	"Adult":       2,
	"Adventure":   3,
	"Animation":   4,
	"Biography":   5,
	"Comedy":      6,
	"Crime":       7,
	"Documentary": 8,
	"Drama":       9,
	"Family":      10,
	"Fantasy":     11,
	"Film-Noir":   12,
	"Game-Show":   13,
	"History":     14,
	"Horror":      15,
	"Music":       16,
	"Musical":     17,
	"Mystery":     18,
	"News":        19,
	"Reality-TV":  20,
	"Romance":     21,
	"Sci-Fi":      22,
	"Short":       23,
	"Sport":       24,
	"Talk-Show":   25,
	"Thriller":    26,
	"War":         27,
	"Western":     28,
}

func GetGenres() map[string]uint64 {
	return ImdbGenres
}
