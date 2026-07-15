// Package i18n holds every player-facing string in each supported language, so
// the UI can be shown in the player's language. English is the default; adding a
// language is just one more Strings value plus an entry in dicts.
package i18n

// Lang identifies a supported language.
type Lang int

const (
	EN Lang = iota // English (default)
	TR             // Turkish
)

// Langs is the selectable order, used when cycling languages.
var Langs = []Lang{EN, TR}

// Label is the language's own name, shown in the language menu.
func (l Lang) Label() string {
	switch l {
	case TR:
		return "Türkçe"
	default:
		return "English"
	}
}

// Next returns the following language, wrapping around.
func (l Lang) Next() Lang {
	return Langs[(int(l)+1)%len(Langs)]
}

// For returns the string set for a language (English if unknown).
func For(l Lang) Strings {
	if s, ok := dicts[l]; ok {
		return s
	}
	return dicts[EN]
}

// Strings is the full set of player-facing text. Fields ending in a verb like
// "Fmt" are format templates filled with fmt.Sprintf by the UI.
type Strings struct {
	// shared
	Tagline string

	// welcome menu
	WPlay, WHowTo, WWhatSSH, WAbout, WLanguage, WQuit string
	WLeaderboard, WNickname                           string
	WNav, WInfoBack                                   string
	WHowToTitle, WHowToBody                           string
	WWhatSSHTitle, WWhatSSHBody                       string
	WAboutTitle, WAboutBody                           string

	// leaderboard & nickname
	WLbTitle, WLbEmpty, WLbYouRankFmt           string
	WNickTitle, WNickHelp, WNickTaken, WNickSet string

	// lobby
	LOpenRooms, LNoRooms, LPlayer string
	LBotWaiting                   string
	LHumanWaitingFmt              string // "%s · %d/2 waiting"
	LFooter                       string
	LCode, LCodeHelp              string
	LPasswordSoon                 string
	LErrNoRoom, LErrRoomFull      string
	LQuitConfirm                  string

	// game — placement
	VsFmt            string // "vs %s"
	PlaceFleet       string
	PlaceHelpFmt     string // "... rotate (%s) ..."
	OrientH, OrientV string
	Ready            string
	OppPlacingFmt    string // "%s is still placing..."

	// game — waiting room
	Room, WaitingOpp, ShareCode, BackHelp string

	// game — battle
	YourWaters, EnemyWaters           string
	YourTurn                          string
	OppAimingFmt                      string // "%s IS AIMING..."
	BattleHelp                        string
	LegendShip, LegendHit, LegendMiss string
	AlreadyFired                      string
	MissFmt, HitFmt, SunkFmt          string

	// game — over
	Victory, Defeat       string
	WinMsgFmt, LoseMsgFmt string
	OverHelp              string
	Opponent              string

	// battle log
	LogTitle                                   string
	LogYouMissFmt, LogYouHitFmt, LogYouSunkFmt string // arg: coord / coord / ship
	LogOppMissFmt, LogOppHitFmt                string // args: name, coord
	LogOppSunkFmt                              string // args: name, ship

	// score & rematch
	ScoreFmt          string // args: yourScore, oppScore, oppName
	RematchHelp       string
	RematchWaitingFmt string // arg: oppName
	OppLeft           string
}

var dicts = map[Lang]Strings{
	EN: {
		Tagline: "battleship over ssh",

		WPlay:        "Play",
		WHowTo:       "How to play",
		WWhatSSH:     "What is SSH?",
		WAbout:       "About",
		WLanguage:    "Language",
		WQuit:        "Quit",
		WLeaderboard: "Leaderboard",
		WNickname:    "Nickname",
		WNav:         "↑↓/jk navigate · enter select · q quit",
		WInfoBack:    "press any key to go back",

		WLbTitle:      "LEADERBOARD — most wins",
		WLbEmpty:      "No games played yet. Be the first!",
		WLbYouRankFmt: "Your rank: #%d  ",
		WNickTitle:    "Choose your nickname",
		WNickHelp:     "type a name · enter to save · esc to cancel",
		WNickTaken:    "That nickname is already taken.",
		WNickSet:      "Nickname saved!",

		WHowToTitle: "How to play",
		WHowToBody: "Place your fleet of 5 ships on the grid, then take turns firing\n" +
			"at your opponent's waters. First to sink the enemy fleet wins.\n\n" +
			"• Placing: arrows/hjkl move · r rotates · enter drops the ship\n" +
			"• Firing:  arrows/hjkl aim · enter fires\n\n" +
			"Play a bot to warm up, or create a room and send the code to a\n" +
			"friend for a 1v1.",

		WWhatSSHTitle: "What is SSH?",
		WWhatSSHBody: "SSH is the tool you already use to log into servers. torpido\n" +
			"turns that same connection into a game: there is nothing to\n" +
			"install and no account to make.\n\n" +
			"Just run:  ssh torpido.dev\n\n" +
			"Your terminal is the whole game.",

		WAboutTitle: "About",
		WAboutBody: "torpido — battleship played entirely in your terminal, over SSH.\n" +
			"Built in Go with Bubble Tea and Wish.\n\n" +
			"Source: github.com/ensardev/ssh-torpido",

		LOpenRooms:       "OPEN ROOMS:",
		LNoRooms:         "(no rooms)",
		LPlayer:          "player",
		LBotWaiting:      "bot · waiting for a challenger",
		LHumanWaitingFmt: "· %d/2 waiting",
		LFooter:          "↑↓ move · enter join · c create · h quick · k code · q menu",
		LCode:            "CODE: ",
		LCodeHelp:        "type the letters · enter to join · esc to cancel",
		LPasswordSoon:    "Password rooms are coming soon 🔜",
		LErrNoRoom:       "No room with that code.",
		LErrRoomFull:     "That room is full.",
		LQuitConfirm:     "Really quit torpido? (y/n)",

		VsFmt:         "vs %s",
		PlaceFleet:    "Place your fleet:",
		PlaceHelpFmt:  "arrows/hjkl move · r rotate (%s) · enter place · q back",
		OrientH:       "horizontal",
		OrientV:       "vertical",
		Ready:         "READY",
		OppPlacingFmt: "%s is still placing their fleet…",

		Room:       "ROOM: ",
		WaitingOpp: "Waiting for an opponent…",
		ShareCode:  "Send this code to a friend — they join with ‘join by code’.",
		BackHelp:   "q back to lobby",

		YourWaters:   "YOUR WATERS",
		EnemyWaters:  "ENEMY WATERS",
		YourTurn:     "YOUR TURN",
		OppAimingFmt: "%s IS AIMING…",
		BattleHelp:   "arrows/hjkl aim · enter fire · q back to lobby",
		LegendShip:   "ship",
		LegendHit:    "hit",
		LegendMiss:   "miss",
		AlreadyFired: "You already fired there.",
		MissFmt:      "%s — miss.",
		HitFmt:       "%s — direct hit! 💥",
		SunkFmt:      "You sank the %s!",

		Victory:    "★  VICTORY  ★",
		Defeat:     "✖  DEFEAT  ✖",
		WinMsgFmt:  "You destroyed %s's fleet!",
		LoseMsgFmt: "%s beat you.",
		OverHelp:   "enter/q back to lobby",
		Opponent:   "opponent",

		LogTitle:      "BATTLE LOG",
		LogYouMissFmt: "You fired at %s — miss",
		LogYouHitFmt:  "You hit %s",
		LogYouSunkFmt: "You sank the enemy %s",
		LogOppMissFmt: "%s fired at %s — miss",
		LogOppHitFmt:  "%s hit you at %s",
		LogOppSunkFmt: "%s sank your %s",

		ScoreFmt:          "You %d — %d %s",
		RematchHelp:       "r rematch · q leave",
		RematchWaitingFmt: "Waiting for %s to accept the rematch…",
		OppLeft:           "Your opponent left.",
	},

	TR: {
		Tagline: "terminal amiral battı",

		WPlay:        "Oyna",
		WHowTo:       "Nasıl oynanır",
		WWhatSSH:     "SSH nedir?",
		WAbout:       "Hakkında",
		WLanguage:    "Dil",
		WQuit:        "Çık",
		WLeaderboard: "Skor Tablosu",
		WNickname:    "Takma ad",
		WNav:         "↑↓/jk gez · enter seç · q çık",
		WInfoBack:    "geri dönmek için bir tuşa bas",

		WLbTitle:      "SKOR TABLOSU — en çok galibiyet",
		WLbEmpty:      "Henüz maç oynanmadı. İlk sen ol!",
		WLbYouRankFmt: "Senin sıran: #%d  ",
		WNickTitle:    "Takma adını seç",
		WNickHelp:     "bir ad yaz · enter kaydet · esc iptal",
		WNickTaken:    "Bu takma ad zaten alınmış.",
		WNickSet:      "Takma ad kaydedildi!",

		WHowToTitle: "Nasıl oynanır",
		WHowToBody: "5 gemilik donanmanı ızgaraya yerleştir, sonra sırayla rakibinin\n" +
			"sularına ateş et. Düşman donanmasını ilk batıran kazanır.\n\n" +
			"• Yerleştirme: ok/hjkl taşı · r döndür · enter gemiyi bırak\n" +
			"• Ateş:        ok/hjkl nişan al · enter ateş\n\n" +
			"Isınmak için bota karşı oyna, ya da oda kurup kodu bir\n" +
			"arkadaşına göndererek 1v1 yap.",

		WWhatSSHTitle: "SSH nedir?",
		WWhatSSHBody: "SSH, sunuculara bağlanmak için zaten kullandığın araç. torpido\n" +
			"bu bağlantıyı bir oyuna çeviriyor: kuracak bir şey yok, açılacak\n" +
			"bir hesap yok.\n\n" +
			"Sadece şunu çalıştır:  ssh torpido.dev\n\n" +
			"Terminalin tüm oyun.",

		WAboutTitle: "Hakkında",
		WAboutBody: "torpido — tamamen terminalinde, SSH üzerinden oynanan amiral battı.\n" +
			"Go, Bubble Tea ve Wish ile yazıldı.\n\n" +
			"Kaynak: github.com/ensardev/ssh-torpido",

		LOpenRooms:       "AÇIK ODALAR:",
		LNoRooms:         "(oda yok)",
		LPlayer:          "oyuncu",
		LBotWaiting:      "bot · rakip bekliyor",
		LHumanWaitingFmt: "· %d/2 bekliyor",
		LFooter:          "↑↓ gez · enter gir · c oda kur · h eşleş · k kod · q menü",
		LCode:            "KOD: ",
		LCodeHelp:        "harfleri yaz · enter katıl · esc iptal",
		LPasswordSoon:    "Şifreli odalar yakında geliyor 🔜",
		LErrNoRoom:       "Bu kodla oda yok.",
		LErrRoomFull:     "Oda dolu.",
		LQuitConfirm:     "torpido'dan çıkmak istediğine emin misin? (e/h)",

		VsFmt:         "%s'e karşı",
		PlaceFleet:    "Donanmanı yerleştir:",
		PlaceHelpFmt:  "ok/hjkl taşı · r döndür (%s) · enter yerleştir · q çık",
		OrientH:       "yatay",
		OrientV:       "dikey",
		Ready:         "HAZIRSIN",
		OppPlacingFmt: "%s hâlâ donanmasını yerleştiriyor…",

		Room:       "ODA: ",
		WaitingOpp: "Rakip bekleniyor…",
		ShareCode:  "Bu kodu arkadaşına gönder — o da ‘kodla katıl’ ile girsin.",
		BackHelp:   "q lobiye dön",

		YourWaters:   "SENİN SULARIN",
		EnemyWaters:  "DÜŞMAN SULARI",
		YourTurn:     "SIRA: SEN",
		OppAimingFmt: "%s NİŞAN ALIYOR…",
		BattleHelp:   "ok/hjkl nişan al · enter ateş · q lobiye dön",
		LegendShip:   "gemi",
		LegendHit:    "isabet",
		LegendMiss:   "ıska",
		AlreadyFired: "Oraya zaten ateş ettin.",
		MissFmt:      "%s — ıska.",
		HitFmt:       "%s — tam isabet! 💥",
		SunkFmt:      "%s gemisini batırdın!",

		Victory:    "★  ZAFER  ★",
		Defeat:     "✖  MAĞLUBİYET  ✖",
		WinMsgFmt:  "%s donanmasını yok ettin!",
		LoseMsgFmt: "%s seni yendi.",
		OverHelp:   "enter/q lobiye dön",
		Opponent:   "rakip",

		LogTitle:      "SAVAŞ KAYDI",
		LogYouMissFmt: "%s'e ateş ettin — ıska",
		LogYouHitFmt:  "%s — isabet",
		LogYouSunkFmt: "Düşmanın %s gemisini batırdın",
		LogOppMissFmt: "%s, %s'e ateş etti — ıska",
		LogOppHitFmt:  "%s seni %s'te vurdu",
		LogOppSunkFmt: "%s senin %s gemini batırdı",

		ScoreFmt:          "Sen %d — %d %s",
		RematchHelp:       "r rövanş · q çık",
		RematchWaitingFmt: "%s'in rövanşı kabul etmesi bekleniyor…",
		OppLeft:           "Rakibin ayrıldı.",
	},
}
