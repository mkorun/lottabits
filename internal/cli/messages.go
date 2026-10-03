package cli

import (
	"fmt"
	"strings"
)

// The message catalog: user-interface text only. Translations never contain data such as numbers, words or mappings;
// those are passed as arguments. TestCatalogsMatch checks that both languages have the same keys and format verbs.
var catalogs = map[string]map[string]string{
	"en": { //nolint:gosec // G101: user-interface text whose keys mention "password", no credentials
		"usage": `LottaBits – physical randomness without dice.

Usage:
  lottabits seed       [--details]          46 draws (01–64) -> 24 BIP39 words
  lottabits password                        1 draw (01–88) per character
  lottabits passphrase --wordlist en|de     2 draws (01–88) per word
  lottabits selftest                        check vectors and data hashes on this device
  lottabits version                         show version and data identifiers

Options for every command:
  --lang en|de    language of messages (never changes a word list)

Draws are read from standard input, never from arguments. On a terminal you are asked for each entry.
`,
		"err_unknown_command":   "unknown command %q",
		"err_option":            "invalid option: %v",
		"err_arguments":         "draws are read from standard input, not from command-line arguments",
		"err_missing_wordlist":  "--wordlist is required: choose en or de explicitly",
		"err_wordlist":          "unknown word list %q (choose en or de)",
		"err_not_a_number":      "draw %d: %q is not a number from 01 to %02d",
		"err_exact_count":       "%d draws needed, got %d",
		"err_min_count":         "at least %d draw needed, got %d",
		"err_even_count":        "an even number of draws (at least %d) needed, got %d",
		"err_input_too_large":   "input is larger than %d bytes",
		"err_read":              "cannot read input: %v",
		"err_aborted":           "input ended before all draws were entered",
		"prompt_seed":           "Word %02d – two draws (01–64): ",
		"prompt_password":       "Character %02d – draw (01–88, empty line to finish): ",
		"prompt_passphrase":     "Word %02d – two draws (01–88, empty line to finish): ",
		"entry_invalid":         "  %q is not a number from 01 to %02d. Please enter again.\n",
		"entry_count":           "  Please enter exactly %d draws.\n",
		"col_word":              "Word",
		"col_draws":             "Draws",
		"col_number":            "Number",
		"seed_title":            "LottaBits seed (%s)",
		"seed_extra":            "Extra bits: %s -> block %d, word numbers %d–%d",
		"seed_wallet":           "Hardware wallet: pick the final word numbered %d–%d, or enter the extra bits %s in this order.",
		"seed_source":           "Entropy source: %d token draws",
		"seed_rng":              "Software-generated randomness: none",
		"seed_entropy":          "Entropy: %s",
		"seed_checksum":         "Checksum: 0x%02x",
		"label_password":        "Password:",
		"label_classes":         "Classes:",
		"classes_legend":        "(D digit, U upper case, L lower case, S symbol)",
		"label_draws":           "Draws:",
		"label_strength":        "Strength:",
		"label_passphrase":      "Passphrase:",
		"label_wordlist":        "Word list:",
		"strength_draws":        "%d draws ≈ %s bits",
		"strength_words":        "%d words ≈ %s bits",
		"warn_password_short":   "Warning: fewer than the recommended %d characters.",
		"warn_passphrase_short": "Warning: fewer than the recommended %d words.",
		"after_interactive":     "Write the result down, then close this window and clear the terminal scroll-back.",
		"selftest_ok":           "ok    %s",
		"selftest_fail":         "FAIL  %s: %v",
		"selftest_passed":       "Self-test passed: %d checks.",
		"selftest_failed":       "SELF-TEST FAILED: %d of %d checks. Do not use this binary.",
	},
	"de": { //nolint:gosec // G101: user-interface text whose keys mention "password", no credentials
		"usage": `LottaBits – physischer Zufall ohne Würfel.

Aufruf:
  lottabits seed       [--details]          46 Ziehungen (01–64) -> 24 BIP39-Wörter
  lottabits password                        1 Ziehung (01–88) pro Zeichen
  lottabits passphrase --wordlist en|de     2 Ziehungen (01–88) pro Wort
  lottabits selftest                        Testvektoren und Daten-Hashes auf diesem Gerät prüfen
  lottabits version                         Version und Daten-Kennungen anzeigen

Option für jeden Befehl:
  --lang en|de    Sprache der Meldungen (ändert nie eine Wortliste)

Ziehungen werden von der Standardeingabe gelesen, nie aus Argumenten. Im Terminal wird jede Eingabe einzeln abgefragt.
`,
		"err_unknown_command":   "unbekannter Befehl %q",
		"err_option":            "ungültige Option: %v",
		"err_arguments":         "Ziehungen werden von der Standardeingabe gelesen, nicht aus Kommandozeilen-Argumenten",
		"err_missing_wordlist":  "--wordlist ist erforderlich: en oder de ausdrücklich wählen",
		"err_wordlist":          "unbekannte Wortliste %q (en oder de wählen)",
		"err_not_a_number":      "Ziehung %d: %q ist keine Zahl von 01 bis %02d",
		"err_exact_count":       "%d Ziehungen nötig, erhalten: %d",
		"err_min_count":         "mindestens %d Ziehung nötig, erhalten: %d",
		"err_even_count":        "eine gerade Anzahl Ziehungen (mindestens %d) nötig, erhalten: %d",
		"err_input_too_large":   "Eingabe ist größer als %d Byte",
		"err_read":              "Eingabe nicht lesbar: %v",
		"err_aborted":           "Eingabe endete, bevor alle Ziehungen eingegeben waren",
		"prompt_seed":           "Wort %02d – zwei Ziehungen (01–64): ",
		"prompt_password":       "Zeichen %02d – Ziehung (01–88, leere Zeile beendet): ",
		"prompt_passphrase":     "Wort %02d – zwei Ziehungen (01–88, leere Zeile beendet): ",
		"entry_invalid":         "  %q ist keine Zahl von 01 bis %02d. Bitte erneut eingeben.\n",
		"entry_count":           "  Bitte genau %d Ziehungen eingeben.\n",
		"col_word":              "Wort",
		"col_draws":             "Ziehungen",
		"col_number":            "Nummer",
		"seed_title":            "LottaBits-Seed (%s)",
		"seed_extra":            "Zusatzbits: %s -> Block %d, Wortnummern %d–%d",
		"seed_wallet":           "Hardware-Wallet: das Schlusswort mit Nummer %d–%d wählen oder die Zusatzbits %s in dieser Reihenfolge eingeben.",
		"seed_source":           "Entropiequelle: %d Token-Ziehungen",
		"seed_rng":              "Software-erzeugter Zufall: keiner",
		"seed_entropy":          "Entropie: %s",
		"seed_checksum":         "Prüfsumme: 0x%02x",
		"label_password":        "Passwort:",
		"label_classes":         "Klassen:",
		"classes_legend":        "(D Ziffer, U Großbuchstabe, L Kleinbuchstabe, S Sonderzeichen)",
		"label_draws":           "Ziehungen:",
		"label_strength":        "Stärke:",
		"label_passphrase":      "Passphrase:",
		"label_wordlist":        "Wortliste:",
		"strength_draws":        "%d Ziehungen ≈ %s Bit",
		"strength_words":        "%d Wörter ≈ %s Bit",
		"warn_password_short":   "Warnung: weniger als die empfohlenen %d Zeichen.",
		"warn_passphrase_short": "Warnung: weniger als die empfohlenen %d Wörter.",
		"after_interactive":     "Ergebnis notieren, dann dieses Fenster schließen und den Terminal-Verlauf löschen.",
		"selftest_ok":           "ok    %s",
		"selftest_fail":         "FEHLER  %s: %v",
		"selftest_passed":       "Selbsttest bestanden: %d Prüfungen.",
		"selftest_failed":       "SELBSTTEST FEHLGESCHLAGEN: %d von %d Prüfungen. Dieses Programm nicht verwenden.",
	},
}

// Languages lists the accepted values of --lang.
var languages = []string{"en", "de"}

type messages struct{ lang string }

// t formats the message with the given key in the current language.
func (m messages) t(key string, args ...any) string {
	return fmt.Sprintf(catalogs[m.lang][key], args...)
}

// decimal formats a number with one decimal place and the language's decimal separator.
func (m messages) decimal(x float64) string {
	s := fmt.Sprintf("%.1f", x)
	if m.lang == "de" {
		s = strings.Replace(s, ".", ",", 1)
	}
	return s
}
